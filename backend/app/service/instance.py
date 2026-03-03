import json
import logging
import secrets
import string
import time
from datetime import datetime
from typing import List, Optional

from kubernetes import client, config as k8s_config
from kubernetes.client.rest import ApiException
from sqlalchemy.orm import Session

from app.config import get_settings
from app.models.instance import Instance
from app.models.user import User
from app.models.image import ImageRepository
from app.schema.instance import InstanceCreate, InstanceUpdate

logger = logging.getLogger(__name__)


class InstanceService:
    def get_by_id(self, db: Session, instance_id: str) -> Optional[Instance]:
        return db.query(Instance).filter(Instance.id == instance_id).first()

    def get_all(self, db: Session, skip: int = 0, limit: int = 100) -> List[Instance]:
        return db.query(Instance).offset(skip).limit(limit).all()

    def get_count(self, db: Session) -> int:
        return db.query(Instance).count()

    def get_by_user(self, db: Session, user_id: int) -> List[Instance]:
        return db.query(Instance).filter(Instance.user_id == user_id).all()

    def _parse_memory(self, memory_str: str) -> int:
        memory_str = memory_str.upper().strip()
        if memory_str.endswith("GI"):
            return int(memory_str[:-2]) * 1024
        if memory_str.endswith("G"):
            return int(memory_str[:-1]) * 1024
        if memory_str.endswith("MI"):
            return int(memory_str[:-2])
        if memory_str.endswith("M"):
            return int(memory_str[:-1])
        return int(memory_str)

    def _get_image_full_name(self, db: Session, repo_id: int, repo_tag: str) -> str:
        repo = db.query(ImageRepository).filter(ImageRepository.id == repo_id).first()
        if not repo:
            return f"ubuntu:{repo_tag}"
        registry_prefix = get_settings().REGISTRY_PREFIX
        return f"{registry_prefix}/{repo.name}:{repo_tag}"

    def _get_memory_mb(self, memory_str: str) -> int:
        return self._parse_memory(memory_str)

    def _get_cpu_request(self, cpu_limit: float) -> str:
        cpu_request = max(0.5, cpu_limit * 0.5)
        return str(cpu_request)

    def _get_memory_request(self, memory_limit: str) -> str:
        memory_mb = self._parse_memory(memory_limit)
        memory_request_mb = max(256, memory_mb // 2)
        if memory_request_mb >= 1024:
            return f"{memory_request_mb // 1024}Gi"
        return f"{memory_request_mb}Mi"


    def create(
        self,
        db: Session,
        instance: InstanceCreate,
        repo_id: Optional[int] = None,
        repo_tag: Optional[str] = None,
        cpu_limit: float = 1.0,
        memory_limit: str = "1Gi",
        storage_limit: int = 1,
        gpu_limit: int = 0,
        gpu_memory_limit: int = 0,
    ) -> Instance:
        # Step 1: 解析镜像来源（仓库镜像 / 直接镜像）
        if repo_id and repo_tag:
            image = self._get_image_full_name(db, repo_id, repo_tag)
            db_repo_id = repo_id
            db_repo_tag = repo_tag
        else:
            image = instance.image or "ubuntu:24.04"
            db_repo_id = None
            db_repo_tag = None

        # Step 2: 计算实例基础参数并生成随机密码
        namespace = instance.namespace or "default"
        replicas = instance.replicas or 1
        memory_limit_mb = self._get_memory_mb(memory_limit)
        password_charset = string.ascii_letters + string.digits
        password = "".join(secrets.choice(password_charset) for _ in range(12))
        ssh_port = get_settings().K8S_SSH_PORT
        USER_PERSONAL_PVC_CAPACITY = get_settings().USER_PERSONAL_PVC_CAPACITY   # 单个PVC的固定容量，避免每个实例都创建一个独占PVC导致资源浪费
        PVC_ACCESS_MODE = get_settings().PVC_ACCESS_MODE   # PVC访问模式，默认为ReadWriteOnce，适用于大多数场景；如果需要多个实例共享同一个PVC，可以改为ReadWriteMany，但需要底层存储支持
        storage_sc_name = get_settings().STORAGE_CLASS_NAME  # 因为在 WSL2 环境下无法使用 Longhorn，改用自带的本地存储

        # Step 3: 创建数据库实例记录（初始状态 pending）
        db_instance = Instance(
            name=instance.name,
            user_id=instance.user_id,
            image=image,
            repo_id=db_repo_id,
            repo_tag=db_repo_tag,
            namespace=namespace,
            replicas=replicas,
            cpu_limit=cpu_limit,
            memory_limit=memory_limit,
            memory_limit_mb=memory_limit_mb,
            storage_limit=storage_limit,
            gpu_limit=gpu_limit,
            gpu_memory_limit=gpu_memory_limit,
            ssh_port=ssh_port,
            config="",
            status="pending",
            started_at=datetime.now(),
        )
        db.add(db_instance)
        db.flush()

        # Step 4: 回填系统字段并提交一次，确保拿到完整实例ID
        instance_id = db_instance.id
        db_instance.ssh_username = "root"
        db.commit()
        db.refresh(db_instance)

        # Step 5: 识别镜像类型并生成安全 hostname
        image_name = image.lower()
        is_codeserver_image = (
            "codeserver" in image_name or "code-server" in image_name
        )

        safe_instance_suffix = "".join(
            ch for ch in instance_id.lower() if ch.isalnum()
        )[:12] or "instance"
        target_hostname = f"mixm-{safe_instance_suffix}"

        # Step 6: 生成通用启动命令（设置 SSH 与 root 登录）
        startup_common_command = (
            # 1. 开启错误即退出模式
            "set -e; "
            
            # 2. 定义变量
            f"TARGET_PASS='{password}'; "

            # 3. 设置 root 密码
            "echo \"root:$TARGET_PASS\" | chpasswd; "

            # 4. 配置 SSH AllowUsers
            "if grep -q '^AllowUsers ' /etc/ssh/sshd_config; then "
            "sed -i \"s/^AllowUsers .*/AllowUsers root/\" /etc/ssh/sshd_config; "
            "else "
            "echo \"AllowUsers root\" >> /etc/ssh/sshd_config; "
            "fi; "
            
            # 5. 启用密码认证
            "sed -i 's/^#\\?PasswordAuthentication.*/PasswordAuthentication yes/' /etc/ssh/sshd_config; "

            # 6. 允许 root 通过密码登录
            "sed -i 's/^#\\?PermitRootLogin.*/PermitRootLogin yes/' /etc/ssh/sshd_config; "
            
            # 7. 禁用公钥认证 (根据你的原逻辑)
            "sed -i 's/^#\\?PubkeyAuthentication.*/PubkeyAuthentication no/' /etc/ssh/sshd_config; "
            
            # 8. 创建 SSH 运行目录
            "mkdir -p /var/run/sshd; "
        )

        # Step 7: 根据镜像类型拼接最终启动命令
        if is_codeserver_image:
            startup_command = (
                startup_common_command
                + "touch /root/.bashrc; "
                + "if ! grep -q 'MIXM_AUTO_VENV' /root/.bashrc; then "
                + "printf '\\n# MIXM_AUTO_VENV\\nsource /opt/venv/bin/activate\\n' >> /root/.bashrc; "
                + "fi; "
                + "/usr/sbin/sshd -D -e & "
                + "exec /bin/bash -lc \"export VIRTUAL_ENV=/opt/venv; export PATH=\\\"/opt/venv/bin:$PATH\\\"; PASSWORD=\\\"$TARGET_PASS\\\" code-server --bind-addr 0.0.0.0:8080 --auth password /workspace\""
            )
        else:
            startup_command = startup_common_command + "exec /usr/sbin/sshd -D -e"

        # Step 8: 计算 request/limit 及容器服务端口
        cpu_request = self._get_cpu_request(cpu_limit)
        memory_request = self._get_memory_request(memory_limit)
        container_port = 8080 if is_codeserver_image else 22

        # Step 9: 生成并写入实例运行配置
        config_data = {
            "startup_mode": "direct",
            "image": image,
            "container_port": container_port,
            "working_dir": "/workspace",
            "cpu_limit": str(cpu_limit),
            "memory_limit": memory_limit,
            "cpu_request": cpu_request,
            "memory_request": memory_request,
            "storage": f"{storage_limit}Gi",
            "ssh_username": db_instance.ssh_username,
            "password": password,
        }

        if gpu_limit > 0:
            config_data["gpu_limit"] = gpu_limit
            config_data["gpu_memory_limit"] = gpu_memory_limit

        gpu_resource_key = "nvidia.com/gpu"
        gpu_memory_resource_key = "nvidia.com/gpumem"

        deplpy_limits = {
            "cpu": str(cpu_limit),
            "memory": memory_limit,
            "ephemeral-storage": f"{storage_limit}Gi",
        }

        if gpu_limit > 0:
            deplpy_limits[gpu_resource_key] = str(gpu_limit)
            if gpu_memory_limit > 0:
                deplpy_limits[gpu_memory_resource_key] = str(gpu_memory_limit)

        deployment_requests = {
            "cpu": cpu_request,
            "memory": memory_request,
        }

        if gpu_limit > 0:
            deployment_requests[gpu_resource_key] = str(gpu_limit)
            if gpu_memory_limit > 0:
                deployment_requests[gpu_memory_resource_key] = str(gpu_memory_limit)

        # Step 10: 持久化 config 配置
        db_instance.config = json.dumps(config_data)
        db.commit()
        db.refresh(db_instance)

        try:
            # Step 11: 初始化 Kubernetes 客户端
            k8s_config.load_kube_config(config_file=get_settings().K8S_CONFIG_PATH)
            apps_v1 = client.AppsV1Api()
            core_v1 = client.CoreV1Api()
            storage_v1 = client.StorageV1Api()
            custom_api = client.CustomObjectsApi()
            net_v1 = client.NetworkingV1Api()
            target_longhorn_replicas = 1

            # Step 12: 生成本次实例资源名称
            pvc_name = f"edu-pvc-cs-{str(instance.user_id)}"  # 限制每个用户拥有自己的pvc，避免重复创建和资源浪费
            deploy_name = f"edu-deploy-cs-{instance_id}"
            svc_name = f"edu-svc-cs-{instance_id}"
            middleware_name = f"edu-mid-cs-{instance_id}"
            ingress_name = f"edu-ing-cs-{instance_id}"
            access_path = f"/instance/{instance_id}"

            # Step 13: 定义容器端口与 Service 端口
            container_ports = [
                client.V1ContainerPort(container_port=22, name="ssh")
            ]
            service_ports = [
                client.V1ServicePort(port=80, target_port=22, name="ssh")
            ]
            if is_codeserver_image:
                container_ports.append(
                    client.V1ContainerPort(container_port=8080, name="http")
                )
                service_ports = [
                    client.V1ServicePort(port=80, target_port=8080, name="http"),
                    client.V1ServicePort(port=22, target_port=22, name="ssh"),
                ]

            # Step 14: 构建用户级 PVC 定义
            pvc = client.V1PersistentVolumeClaim(
                metadata=client.V1ObjectMeta(
                    name=pvc_name,
                    annotations={"longhorn.io/numberOfReplicas": "1"},
                ),
                spec=client.V1PersistentVolumeClaimSpec(
                    access_modes=[PVC_ACCESS_MODE], 
                    resources=client.V1ResourceRequirements(
                        requests={"storage": USER_PERSONAL_PVC_CAPACITY}
                    ),
                    storage_class_name=storage_sc_name, 
                ),
            )

            
            # ====== 新增：检查PVC是否已存在 ======
            wait_for_first_consumer = False
            is_longhorn_storage_class = False
            try:
                            # Step 15: 读取 StorageClass，判断是否 Longhorn 与绑定策略
                try:
                    storage_class = storage_v1.read_storage_class(name=storage_sc_name)
                    wait_for_first_consumer = self._is_wait_for_first_consumer(
                        storage_class
                    )
                    is_longhorn_storage_class = self._is_longhorn_storage_class(
                        storage_class
                    )
                    print(f"✅ Found StorageClass: {storage_sc_name}")
                except client.rest.ApiException as e:
                    if e.status == 404:
                        raise Exception(f"❌ StorageClass '{storage_sc_name}' not found! K3s usually comes with 'local-path'.")
                    else:
                        raise e

                # Step 16: 检查 PVC 是否已存在
                core_v1.read_namespaced_persistent_volume_claim(
                    name=pvc_name,
                    namespace=namespace
                )
                print(f"PVC {pvc_name} already exists, skipping creation.")
            except client.rest.ApiException as e:
                if e.status == 404:
                    # Step 17: PVC 不存在时创建 PVC
                    try:
                        print(f"PVC {pvc_name} not found, creating new PVC...")
                        core_v1.create_namespaced_persistent_volume_claim(
                            namespace=namespace,
                            body=pvc,
                        )
                    except ApiException as e:
                        if e.status != 409:
                            raise e
                else:
                    raise e # 其他错误（如权限问题）直接抛出

            # Step 18: 必要时等待 PVC 进入 Bound
            if wait_for_first_consumer:
                logger.info(
                    "[K8s] StorageClass %s uses WaitForFirstConsumer, skip pre-deployment PVC Bound wait",
                    storage_sc_name,
                )
            else:
                pvc_ready = self._wait_for_pvc_bound(
                    core_v1,
                    pvc_name,
                    namespace,
                    timeout=180,
                )
                if not pvc_ready:
                    raise Exception(
                        f"PVC {pvc_name} is not Bound in namespace {namespace}"
                    )
                print(f"✅ PVC {pvc_name} is ready or already exists.检测")

            # Step 19: Longhorn 场景下强制卷副本数为 1
            if is_longhorn_storage_class:
                enforced = self._ensure_longhorn_volume_replicas(
                    core_v1=core_v1,
                    custom_api=custom_api,
                    pvc_name=pvc_name,
                    pvc_namespace=namespace,
                    replicas=target_longhorn_replicas,
                )
                if not enforced:
                    raise Exception(
                        f"Failed to enforce Longhorn numberOfReplicas={target_longhorn_replicas} for PVC {pvc_name}"
                    )

            # Step 20: 创建 Deployment
            deployment = client.V1Deployment(
                api_version="apps/v1",
                kind="Deployment",
                metadata=client.V1ObjectMeta(name=deploy_name),
                spec=client.V1DeploymentSpec(
                    replicas=replicas,
                    selector=client.V1LabelSelector(match_labels={"app": deploy_name}),
                    template=client.V1PodTemplateSpec(
                        metadata=client.V1ObjectMeta(labels={"app": deploy_name}),
                        spec=client.V1PodSpec(
                            hostname=target_hostname,
                            containers=[
                                client.V1Container(
                                    name="ubuntu-runtime",
                                    image=image,
                                    command=["bash", "-lc", startup_command],
                                    working_dir="/workspace",
                                    ports=container_ports,
                                    resources=client.V1ResourceRequirements(
                                        limits=deplpy_limits,
                                        requests=deployment_requests,
                                    ),
                                    volume_mounts=[
                                        client.V1VolumeMount(
                                            name="project-data",
                                            mount_path="/workspace",
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

            # Step 21: 创建 Service
            service = client.V1Service(
                metadata=client.V1ObjectMeta(name=svc_name),
                spec=client.V1ServiceSpec(
                    type="ClusterIP",
                    selector={"app": deploy_name},
                    ports=service_ports,
                ),
            )
            try:
                core_v1.create_namespaced_service(namespace=namespace, body=service)
            except ApiException as e:
                if e.status != 409:
                    raise e

            # Step 22: 创建 Traefik Middleware
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

            # Step 23: 创建 Ingress
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

            # Step 24: 等待 Deployment 就绪
            ready = self._wait_for_deployment(
                apps_v1,
                deploy_name,
                namespace,
                timeout=300,
            )
            if not ready:
                raise Exception("Deployment failed to become ready")

        except Exception as e:
            logger.error(f"[K8s] Error creating instance {instance_id}: {str(e)}")
            db.rollback()
            raise Exception(f"Failed to create instance: {str(e)}")

        # Step 25: 最终将状态更新为 running 并返回
        db_instance.status = "running"
        db.commit()
        db.refresh(db_instance)
        return db_instance

    def _wait_for_deployment(
        self,
        apps_v1: client.AppsV1Api,
        name: str,
        namespace: str,
        timeout: int = 300,
        interval: int = 5,
    ) -> bool:
        start_time = time.time()

        while time.time() - start_time < timeout:
            try:
                deployment = apps_v1.read_namespaced_deployment(
                    name=name,
                    namespace=namespace,
                )
                status = deployment.status
                if status:
                    available_replicas = status.available_replicas or 0
                    ready_replicas = status.ready_replicas or 0
                    desired_replicas = deployment.spec.replicas or 1
                    if (
                        ready_replicas >= desired_replicas
                        and available_replicas >= desired_replicas
                    ):
                        return True
            except ApiException as e:
                if e.status != 404:
                    logger.error(f"[K8s] Error checking deployment status: {e}")
                    return False
            except Exception as e:
                logger.error(f"[K8s] Unexpected error checking deployment: {e}")
                return False

            time.sleep(interval)

        logger.warning(f"[K8s] Timeout waiting for deployment {name}")
        return False

    def _wait_for_pvc_bound(
        self,
        core_v1: client.CoreV1Api,
        name: str,
        namespace: str,
        timeout: int = 180,
        interval: int = 5,
    ) -> bool:
        start_time = time.time()

        while time.time() - start_time < timeout:
            try:
                pvc = core_v1.read_namespaced_persistent_volume_claim(
                    name=name,
                    namespace=namespace,
                )
                phase = pvc.status.phase if pvc and pvc.status else None
                if phase == "Bound":
                    return True
            except ApiException as e:
                if e.status != 404:
                    logger.error(f"[K8s] Error checking PVC status: {e}")
                    return False
            except Exception as e:
                logger.error(f"[K8s] Unexpected error checking PVC: {e}")
                return False

            time.sleep(interval)

        logger.warning(f"[K8s] Timeout waiting for PVC {name} to be Bound")
        return False

    def _is_wait_for_first_consumer(self, storage_class) -> bool:
        if not storage_class:
            return False
        binding_mode = getattr(storage_class, "volume_binding_mode", None)
        return binding_mode == "WaitForFirstConsumer"

    def _is_longhorn_storage_class(self, storage_class) -> bool:
        if not storage_class:
            return False
        provisioner = getattr(storage_class, "provisioner", "") or ""
        return "longhorn" in provisioner.lower()

    def _ensure_longhorn_volume_replicas(
        self,
        core_v1: client.CoreV1Api,
        custom_api: client.CustomObjectsApi,
        pvc_name: str,
        pvc_namespace: str,
        replicas: int,
        timeout: int = 180,
        interval: int = 5,
    ) -> bool:
        start_time = time.time()
        volume_name = None

        while time.time() - start_time < timeout:
            try:
                pvc = core_v1.read_namespaced_persistent_volume_claim(
                    name=pvc_name,
                    namespace=pvc_namespace,
                )
                volume_name = pvc.spec.volume_name if pvc and pvc.spec else None
                if volume_name:
                    break
            except ApiException as e:
                if e.status != 404:
                    logger.error(f"[K8s] Error reading PVC {pvc_name}: {e}")
                    return False
            except Exception as e:
                logger.error(f"[K8s] Unexpected error reading PVC {pvc_name}: {e}")
                return False

            time.sleep(interval)

        if not volume_name:
            logger.warning(
                "[Longhorn] Timeout waiting PVC %s/%s to get PV binding",
                pvc_namespace,
                pvc_name,
            )
            return False

        for api_version in ["v1beta2", "v1beta1"]:
            try:
                volume = custom_api.get_namespaced_custom_object(
                    group="longhorn.io",
                    version=api_version,
                    namespace="longhorn-system",
                    plural="volumes",
                    name=volume_name,
                )
                current_replicas = (
                    volume.get("spec", {}).get("numberOfReplicas")
                    if isinstance(volume, dict)
                    else None
                )
                if current_replicas == replicas:
                    return True

                custom_api.patch_namespaced_custom_object(
                    group="longhorn.io",
                    version=api_version,
                    namespace="longhorn-system",
                    plural="volumes",
                    name=volume_name,
                    body={"spec": {"numberOfReplicas": replicas}},
                )
                logger.info(
                    "[Longhorn] Patched volume %s numberOfReplicas: %s -> %s",
                    volume_name,
                    current_replicas,
                    replicas,
                )
                return True
            except ApiException as e:
                if e.status == 404:
                    continue
                logger.error(
                    "[Longhorn] Error patching volume %s with %s: %s",
                    volume_name,
                    api_version,
                    e,
                )
                return False
            except Exception as e:
                logger.error(
                    "[Longhorn] Unexpected error patching volume %s: %s",
                    volume_name,
                    e,
                )
                return False

        logger.error(
            "[Longhorn] Volume CRD not found for volume %s in longhorn-system",
            volume_name,
        )
        return False

    def update(
        self,
        db: Session,
        instance_id: str,
        instance: InstanceUpdate,
    ) -> Optional[Instance]:
        db_instance = self.get_by_id(db, instance_id)
        if not db_instance:
            return None

        update_data = instance.model_dump(exclude_unset=True)
        for field, value in update_data.items():
            setattr(db_instance, field, value)

        db.commit()
        db.refresh(db_instance)
        return db_instance

    def update_status(
        self,
        db: Session,
        instance_id: str,
        status: str,
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
                name=deploy_name,
                namespace=namespace,
                body=body,
            )
        except Exception as e:
            logger.error(f"[K8s] Error scaling instance {instance_id}: {str(e)}")

        db_instance.status = status
        db.commit()
        db.refresh(db_instance)
        return db_instance

    def delete(self, db: Session, instance_id: str) -> bool:
        db_instance = self.get_by_id(db, instance_id)
        if not db_instance:
            return False

        try:
            kubeconfig_path = get_settings().K8S_CONFIG_PATH
            k8s_config.load_kube_config(config_file=kubeconfig_path)
            apps_v1 = client.AppsV1Api()
            core_v1 = client.CoreV1Api()
            custom_api = client.CustomObjectsApi()
            net_v1 = client.NetworkingV1Api()

            namespace = db_instance.namespace or "default"
            object_id = db_instance.id

            deploy_name = f"deploy-cs-{object_id}"
            svc_name = f"svc-cs-{object_id}"
            middleware_name = f"mid-cs-{object_id}"
            ingress_name = f"ing-cs-{object_id}"

            try:
                net_v1.delete_namespaced_ingress(name=ingress_name, namespace=namespace)
            except Exception:
                pass
            try:
                core_v1.delete_namespaced_service(name=svc_name, namespace=namespace)
            except Exception:
                pass
            try:
                custom_api.delete_namespaced_custom_object(
                    group="traefik.io",
                    version="v1alpha1",
                    namespace=namespace,
                    plural="middlewares",
                    name=middleware_name,
                )
            except Exception:
                pass
            try:
                apps_v1.delete_namespaced_deployment(
                    name=deploy_name, namespace=namespace
                )
            except Exception:
                pass
        except Exception as e:
            logger.error(
                f"[K8s] Error deleting instance resources {instance_id}: {str(e)}"
            )

        db.delete(db_instance)
        db.commit()
        return True

    def to_response(self, db: Session, instance: Instance) -> dict:
        user = db.query(User).filter(User.id == instance.user_id).first()
        access_path = f"/instance/{instance.id}"
        ssh_host = get_settings().K8S_SSH_GATEWAY_HOST
        code_server_host = get_settings().K8S_CODE_SERVER_HOST
        ssh_password = None
        if instance.config:
            try:
                config_data = (
                    json.loads(instance.config)
                    if isinstance(instance.config, str)
                    else instance.config
                )
                ssh_password = config_data.get("password")
            except Exception:
                ssh_password = None
        return {
            "id": instance.id,
            "name": instance.name,
            "user_id": instance.user_id,
            "username": user.username if user else None,
            "image": instance.image,
            "repo_id": instance.repo_id,
            "repo_tag": instance.repo_tag,
            "status": instance.status,
            "namespace": instance.namespace,
            "replicas": instance.replicas,
            "cpu_limit": instance.cpu_limit,
            "memory_limit": instance.memory_limit,
            "storage_limit": instance.storage_limit,
            "gpu_limit": instance.gpu_limit,
            "gpu_memory_limit": instance.gpu_memory_limit,
            "cpu_usage": instance.cpu_usage,
            "memory_usage": instance.memory_usage,
            "ssh_port": instance.ssh_port,
            "ssh_username": instance.ssh_username,
            "ssh_password": ssh_password,
            "ssh_connection": f"ssh {instance.ssh_username}@{ssh_host} -p {instance.ssh_port}"
            if instance.ssh_username and instance.ssh_port
            else None,
            "config": instance.config,
            "hostIp": code_server_host,
            "endpoint": access_path,
            "started_at": instance.started_at,
            "created_at": instance.created_at,
            "updated_at": instance.updated_at,
        }
