import logging
import secrets
import string
import json
import time
from typing import List, Optional, Union, cast

from sqlalchemy.orm import Session
from sqlalchemy import select
from kubernetes import client, config as k8s_config
from kubernetes.client.rest import ApiException

from app.models.instance import Instance
from app.models.template import Template
from app.models.user import User
from app.schema.instance import InstanceCreate, InstanceUpdate
from app.config import get_settings
from datetime import datetime

logger = logging.getLogger(__name__)


class InstanceService:
    def get_by_id(self, db: Session, instance_id: int) -> Optional[Instance]:
        return db.query(Instance).filter(Instance.id == instance_id).first()

    def get_all(self, db: Session, skip: int = 0, limit: int = 100) -> List[Instance]:
        return db.query(Instance).offset(skip).limit(limit).all()

    def get_count(self, db: Session) -> int:
        return db.query(Instance).count()

    def get_by_user(self, db: Session, user_id: int) -> List[Instance]:
        return db.query(Instance).filter(Instance.user_id == user_id).all()

    def get_by_template(self, db: Session, template_id: int) -> List[Instance]:
        return db.query(Instance).filter(Instance.template_id == template_id).all()

    def _parse_memory(self, memory_str: str) -> int:
        memory_str = memory_str.upper().strip()
        if memory_str.endswith("GI"):
            return int(memory_str[:-2]) * 1024
        elif memory_str.endswith("G"):
            return int(memory_str[:-1]) * 1024
        elif memory_str.endswith("MI"):
            return int(memory_str[:-2])
        elif memory_str.endswith("M"):
            return int(memory_str[:-1])
        return int(memory_str)

    def create(self, db: Session, instance: InstanceCreate) -> Instance:
        template = (
            db.query(Template).filter(Template.id == instance.template_id).first()
        )
        if not template:
            raise ValueError(f"Template with id {instance.template_id} not found")

        template_config_str = (
            cast(str, template.config) if template.config is not None else "{}"
        )
        template_config = json.loads(template_config_str)

        password = secrets.token_urlsafe(12)
        config_data = {"password": password}
        config_data.update(template_config)

        db_instance = Instance(
            name=instance.name,
            template_id=instance.template_id,
            user_id=instance.user_id,
            namespace=instance.namespace or template_config.get("namespace", "default"),
            replicas=instance.replicas or template_config.get("replicas", 1),
            memory_limit=instance.memory_limit
            or self._parse_memory(template_config.get("memory_limit", "1Gi")),
            config=json.dumps(config_data),
            status="pending",
            started_at=datetime.now(),
        )
        db.add(db_instance)
        db.flush()

        instance_id = db_instance.id
        user_id = db_instance.user_id

        image = template_config.get("image", "codercom/code-server:latest")
        container_port = template_config.get("container_port", 8080)
        working_dir = template_config.get("working_dir", "/home/coder/project")
        cpu_limit = template_config.get("cpu_limit", "1")
        memory_limit = template_config.get("memory_limit", "1Gi")
        cpu_request = template_config.get("cpu_request", "0.5")
        memory_request = template_config.get("memory_request", "512Mi")
        storage = template_config.get("storage", "1Gi")
        template_env = template_config.get("env", [])

        try:
            k8s_config.load_kube_config(config_file=get_settings().K8S_CONFIG_PATH)
            apps_v1 = client.AppsV1Api()
            core_v1 = client.CoreV1Api()
            custom_api = client.CustomObjectsApi()
            net_v1 = client.NetworkingV1Api()

            namespace = db_instance.namespace

            pvc_name = f"pvc-cs-{instance_id}"
            deploy_name = f"deploy-cs-{instance_id}"
            svc_name = f"svc-cs-{instance_id}"
            middleware_name = f"mid-cs-{instance_id}"
            ingress_name = f"ing-cs-{instance_id}"
            access_path = f"/instance/{instance_id}"

            pvc = client.V1PersistentVolumeClaim(
                metadata=client.V1ObjectMeta(name=pvc_name),
                spec=client.V1PersistentVolumeClaimSpec(
                    access_modes=["ReadWriteOnce"],
                    resources=client.V1ResourceRequirements(
                        requests={"storage": storage}
                    ),
                ),
            )
            try:
                core_v1.create_namespaced_persistent_volume_claim(
                    namespace=namespace, body=pvc
                )
            except ApiException as e:
                if e.status != 409:
                    raise e

            # 2. Deployment
            env_list = [
                client.V1EnvVar(name="PASSWORD", value=password),
                client.V1EnvVar(name="VSCODE_PROXY_URI", value=access_path),
                client.V1EnvVar(name="CS_DISABLE_PROXY_URI_VALIDATION", value="1"),
                client.V1EnvVar(name="CS_ALLOW_IFRAME", value="true"),
            ]
            for env_item in template_env:
                if isinstance(env_item, dict):
                    env_list.append(
                        client.V1EnvVar(
                            name=env_item.get("name"), value=env_item.get("value")
                        )
                    )

            deployment = client.V1Deployment(
                api_version="apps/v1",
                kind="Deployment",
                metadata=client.V1ObjectMeta(name=deploy_name),
                spec=client.V1DeploymentSpec(
                    replicas=1,
                    selector=client.V1LabelSelector(match_labels={"app": deploy_name}),
                    template=client.V1PodTemplateSpec(
                        metadata=client.V1ObjectMeta(labels={"app": deploy_name}),
                        spec=client.V1PodSpec(
                            containers=[
                                client.V1Container(
                                    name="code-server",
                                    image=image,
                                    working_dir=working_dir,
                                    ports=[
                                        client.V1ContainerPort(
                                            container_port=container_port
                                        )
                                    ],
                                    env=env_list,
                                    resources=client.V1ResourceRequirements(
                                        limits={
                                            "cpu": cpu_limit,
                                            "memory": memory_limit,
                                        },
                                        requests={
                                            "cpu": cpu_request,
                                            "memory": memory_request,
                                        },
                                    ),
                                    volume_mounts=[
                                        client.V1VolumeMount(
                                            name="project-data",
                                            mount_path=working_dir,
                                        )
                                    ],
                                )
                            ],
                            volumes=[
                                client.V1Volume(
                                    name="project-data",
                                    persistent_volume_claim=client.V1PersistentVolumeClaimVolumeSource(
                                        claim_name=pvc_name
                                    ),
                                )
                            ],
                        ),
                    ),
                ),
            )
            try:
                apps_v1.create_namespaced_deployment(
                    namespace=namespace, body=deployment
                )
            except ApiException as e:
                if e.status != 409:
                    raise e

            # 3. Service
            service = client.V1Service(
                metadata=client.V1ObjectMeta(name=svc_name),
                spec=client.V1ServiceSpec(
                    type="ClusterIP",
                    selector={"app": deploy_name},
                    ports=[client.V1ServicePort(port=80, target_port=container_port)],
                ),
            )
            try:
                core_v1.create_namespaced_service(namespace=namespace, body=service)
            except ApiException as e:
                if e.status != 409:
                    raise e

            # 4. Middleware
            middleware_body = {
                "apiVersion": "traefik.io/v1alpha1",
                "kind": "Middleware",
                "metadata": {"name": middleware_name},
                "spec": {
                    "stripPrefix": {"prefixes": [access_path], "forceSlash": True}
                },
            }
            try:
                custom_api.create_namespaced_custom_object(
                    group="traefik.io",
                    version="v1alpha1",
                    namespace=namespace,
                    plural="middlewares",
                    body=middleware_body,
                )
            except ApiException as e:
                if e.status != 409:
                    raise e

            # 5. Ingress
            ingress = client.V1Ingress(
                metadata=client.V1ObjectMeta(
                    name=ingress_name,
                    annotations={
                        "traefik.ingress.kubernetes.io/router.middlewares": f"{namespace}-{middleware_name}@kubernetescrd"
                    },
                ),
                spec=client.V1IngressSpec(
                    ingress_class_name="traefik",
                    rules=[
                        client.V1IngressRule(
                            http=client.V1HTTPIngressRuleValue(
                                paths=[
                                    client.V1HTTPIngressPath(
                                        path=access_path,
                                        path_type="Prefix",
                                        backend=client.V1IngressBackend(
                                            service=client.V1IngressServiceBackend(
                                                name=svc_name,
                                                port=client.V1ServiceBackendPort(
                                                    number=80
                                                ),
                                            )
                                        ),
                                    )
                                ]
                            )
                        )
                    ],
                ),
            )
            try:
                net_v1.create_namespaced_ingress(namespace=namespace, body=ingress)
            except ApiException as e:
                if e.status != 409:
                    raise e

            # 6. 等待 Deployment 就绪
            logger.info(f"[K8s] Waiting for deployment {deploy_name} to be ready...")
            ready = self._wait_for_deployment(
                apps_v1, deploy_name, namespace, timeout=300
            )

            if not ready:
                logger.error(
                    f"[K8s] Deployment {deploy_name} failed to become ready within timeout"
                )
                raise Exception("Deployment failed to become ready")

            logger.info(f"[K8s] Deployment {deploy_name} is ready")

        except Exception as e:
            logger.error(f"[K8s] Error creating instance {instance_id}: {str(e)}")
            db.rollback()
            raise Exception(f"Failed to create instance: {str(e)}")
        # ---------------- Demo Implementation End ----------------

        db_instance.status = "running"
        db.commit()
        db.refresh(db_instance)
        logger.info(
            f"[Instance] Instance '{db_instance.name}' (ID: {instance_id}) created and resources deployed successfully."
        )
        return db_instance

    def _wait_for_deployment(
        self,
        apps_v1: client.AppsV1Api,
        name: str,
        namespace: str,
        timeout: int = 300,
        interval: int = 5,
    ) -> bool:
        """等待 Deployment 就绪

        Args:
            apps_v1: K8s AppsV1Api 客户端
            name: Deployment 名称
            namespace: 命名空间
            timeout: 超时时间（秒）
            interval: 检查间隔（秒）

        Returns:
            bool: 是否成功就绪
        """
        start_time = time.time()

        while time.time() - start_time < timeout:
            try:
                deployment = apps_v1.read_namespaced_deployment(
                    name=name, namespace=namespace
                )

                # 检查 deployment 状态
                status = deployment.status
                if status:
                    available_replicas = status.available_replicas or 0
                    ready_replicas = status.ready_replicas or 0
                    desired_replicas = deployment.spec.replicas or 1

                    logger.debug(
                        f"[K8s] Deployment {name}: "
                        f"available={available_replicas}, "
                        f"ready={ready_replicas}, "
                        f"desired={desired_replicas}"
                    )

                    # 检查是否所有副本都已就绪且可用
                    if (
                        ready_replicas >= desired_replicas
                        and available_replicas >= desired_replicas
                    ):
                        return True

                    # 检查是否有失败的状况
                    if status.conditions:
                        for condition in status.conditions:
                            if (
                                condition.type == "Progressing"
                                and condition.status == "False"
                            ):
                                if (
                                    "Failed" in condition.reason
                                    or "Error" in condition.message
                                ):
                                    logger.error(
                                        f"[K8s] Deployment {name} failed: {condition.message}"
                                    )
                                    return False
            except ApiException as e:
                if e.status != 404:  # 404 表示 deployment 还未创建
                    logger.error(f"[K8s] Error checking deployment status: {e}")
                    return False
            except Exception as e:
                logger.error(f"[K8s] Unexpected error checking deployment: {e}")
                return False

            time.sleep(interval)

        logger.warning(f"[K8s] Timeout waiting for deployment {name}")
        return False

    def update(
        self, db: Session, instance_id: int, instance: InstanceUpdate
    ) -> Optional[Instance]:
        db_instance = self.get_by_id(db, instance_id)
        if not db_instance:
            return None

        # TODO: 同步更新 K8s 资源
        # logger.info(f"[K8s] Updating instance {instance_id}")
        # from kubernetes import client, config
        # config.load_kube_config()
        # apps_v1 = client.AppsV1Api()
        # 如需滚动更新: apps_v1.patch_namespaced_deployment(...)

        update_data = instance.model_dump(exclude_unset=True)
        for field, value in update_data.items():
            setattr(db_instance, field, value)

        db.commit()
        db.refresh(db_instance)
        return db_instance

    def update_status(
        self, db: Session, instance_id: int, status: str
    ) -> Optional[Instance]:
        db_instance = self.get_by_id(db, instance_id)
        if not db_instance:
            return None

        try:
            k8s_config.load_kube_config()
            apps_v1 = client.AppsV1Api()
            deploy_name = f"deploy-cs-{db_instance.id}"
            namespace = db_instance.namespace or "default"

            replicas = 1 if status == "running" else 0
            body = {"spec": {"replicas": replicas}}

            apps_v1.patch_namespaced_deployment_scale(
                name=deploy_name, namespace=namespace, body=body
            )
            logger.info(f"[K8s] Scaled deployment {deploy_name} to {replicas}")
        except Exception as e:
            logger.error(f"[K8s] Error scaling instance {instance_id}: {str(e)}")

        db_instance.status = status
        db.commit()
        db.refresh(db_instance)
        logger.info(f"[Instance] Instance {instance_id} status updated to: {status}")
        return db_instance

    def delete(self, db: Session, instance_id: int) -> bool:
        db_instance = self.get_by_id(db, instance_id)
        if not db_instance:
            return False

        # ---------------- Demo Implementation Start ----------------
        try:
            kubeconfig_path = get_settings().K8S_CONFIG_PATH
            k8s_config.load_kube_config(config_file=kubeconfig_path)
            apps_v1 = client.AppsV1Api()
            core_v1 = client.CoreV1Api()
            custom_api = client.CustomObjectsApi()
            net_v1 = client.NetworkingV1Api()

            instance_id = db_instance.id
            namespace = db_instance.namespace or "default"

            pvc_name = f"pvc-cs-{instance_id}"
            deploy_name = f"deploy-cs-{instance_id}"
            svc_name = f"svc-cs-{instance_id}"
            middleware_name = f"mid-cs-{instance_id}"
            ingress_name = f"ing-cs-{instance_id}"

            # Delete Ingress
            try:
                net_v1.delete_namespaced_ingress(name=ingress_name, namespace=namespace)
            except:
                pass

            # Delete Service
            try:
                core_v1.delete_namespaced_service(name=svc_name, namespace=namespace)
            except:
                pass

            # Delete Middleware
            try:
                custom_api.delete_namespaced_custom_object(
                    group="traefik.io",
                    version="v1alpha1",
                    namespace=namespace,
                    plural="middlewares",
                    name=middleware_name,
                )
            except:
                pass

            # Delete Deployment
            try:
                apps_v1.delete_namespaced_deployment(
                    name=deploy_name, namespace=namespace
                )
            except:
                pass

            # Delete PVC
            try:
                core_v1.delete_namespaced_persistent_volume_claim(
                    name=pvc_name, namespace=namespace
                )
            except:
                pass

            logger.info(f"[K8s] All resources deleted for instance {instance_id}")
        except Exception as e:
            logger.error(
                f"[K8s] Error deleting instance resources {instance_id}: {str(e)}"
            )
        # ---------------- Demo Implementation End ----------------

        db.delete(db_instance)
        db.commit()
        logger.info(f"[Instance] Instance {instance_id} deleted from database")
        return True

    def to_response(self, db: Session, instance: Instance) -> dict:
        template = (
            db.query(Template).filter(Template.id == instance.template_id).first()
        )
        user = db.query(User).filter(User.id == instance.user_id).first()

        access_path = f"/instance/{instance.id}"
        endpoint = access_path
        port = get_settings().K8S_EXPORT_PORT
        return {
            "id": instance.id,
            "name": instance.name,
            "template_id": instance.template_id,
            "template_name": template.name if template else None,
            "user_id": instance.user_id,
            "username": user.username if user else None,
            "status": instance.status,
            "namespace": instance.namespace,
            "replicas": instance.replicas,
            "cpu_usage": instance.cpu_usage,
            "memory_usage": instance.memory_usage,
            "memory_limit": instance.memory_limit,
            "config": instance.config,
            "hostIp": get_settings().K8S_EXPORT_HOST,  # 这里直接返回 IP 和端口，前端拼接成完整 URL
            "endpoint": endpoint,
            "port": port,
            "started_at": instance.started_at,
            "created_at": instance.created_at,
            "updated_at": instance.updated_at,
        }

