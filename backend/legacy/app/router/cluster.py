from fastapi import APIRouter, Depends
from datetime import datetime
from typing import Optional

from kubernetes import client, config as k8s_config
from requests import Session

from app.database import get_db
from app.service.instance import InstanceService
from app.config import get_settings
import logging

logger = logging.getLogger(__name__)

router = APIRouter()
instance_service = InstanceService()

KUBECONFIG_PATH = get_settings().K8S_CONFIG_PATH


def _get_k8s_client():
    try:
        k8s_config.load_kube_config(config_file=KUBECONFIG_PATH)
    except Exception as e:
        logger.warning(f"Failed to load kubeconfig: {e}")
        try:
            k8s_config.load_incluster_config()
        except Exception:
            pass
    return client


def _parse_memory(memory_str: str) -> int:
    if not memory_str:
        return 0
    memory_str = memory_str.upper().strip()
    if memory_str.endswith("GI"):
        return int(memory_str[:-2]) * 1024 * 1024 * 1024
    if memory_str.endswith("G"):
        return int(memory_str[:-1]) * 1024 * 1024 * 1024
    if memory_str.endswith("MI"):
        return int(memory_str[:-2]) * 1024 * 1024
    if memory_str.endswith("M"):
        return int(memory_str[:-1]) * 1024 * 1024
    if memory_str.endswith("KI"):
        return int(memory_str[:-2]) * 1024
    if memory_str.endswith("K"):
        return int(memory_str[:-1]) * 1024
    return int(memory_str)


def _parse_cpu(cpu_str: str) -> int:
    if not cpu_str:
        return 0
    cpu_str = cpu_str.strip()
    if cpu_str.endswith("m"):
        return int(cpu_str[:-1])
    return int(float(cpu_str) * 1000)


def _format_age(creation_timestamp) -> str:
    if not creation_timestamp:
        return "unknown"
    now = datetime.now(creation_timestamp.tzinfo)
    delta = now - creation_timestamp
    days = delta.days
    hours = delta.seconds // 3600
    if days > 0:
        return f"{days}d {hours}h"
    elif hours > 0:
        return f"{hours}h {delta.seconds % 3600 // 60}m"
    else:
        return f"{delta.seconds // 60}m"


@router.get("/pods")
def list_pods(namespace: Optional[str] = None):
    k8s = _get_k8s_client()
    core_v1 = k8s.CoreV1Api()

    pods = {}
    namespaces = [namespace] if namespace else []

    if not namespace:
        try:
            ns_list = core_v1.list_namespace()
            namespaces = [ns.metadata.name for ns in ns_list.items]
        except Exception:
            namespaces = ["default"]

    for ns in namespaces:
        try:
            pod_list = core_v1.list_namespaced_pod(namespace=ns)
            for pod in pod_list.items:
                pod_name = pod.metadata.name
                if pod_name not in pods:
                    pods[pod_name] = {
                        "name": pod_name,
                        "namespace": ns,
                        "node": pod.spec.node_name or "N/A",
                        "status": pod.status.phase.lower(),
                        "containers": [c.name for c in pod.spec.containers],
                        "restarts": sum(
                            cs.restart_count
                            for cs in pod.status.container_statuses or []
                        ),
                        "age": _format_age(pod.metadata.creation_timestamp),
                    }
        except Exception as e:
            logger.error(f"Failed to list pods in namespace {ns}: {e}")
            continue

    return {"total": len(pods), "pods": list(pods.values())}


@router.get("/nodes")
def list_nodes():
    k8s = _get_k8s_client()
    core_v1 = k8s.CoreV1Api()
    custom_api = k8s.CustomObjectsApi()

    nodes = {}
    pod_resources = {}

    try:
        pod_list = core_v1.list_pod_for_all_namespaces()
        for pod in pod_list.items:
            node_name = pod.spec.node_name
            if not node_name:
                continue
            if node_name not in pod_resources:
                pod_resources[node_name] = {"cpu": 0, "memory": 0, "pods": 0}
            pod_resources[node_name]["pods"] += 1
            for container in pod.spec.containers:
                resources = container.resources
                if resources and resources.requests:
                    cpu_request = resources.requests.get("cpu", "0")
                    memory_request = resources.requests.get("memory", "0")
                    pod_resources[node_name]["cpu"] += _parse_cpu(cpu_request)
                    pod_resources[node_name]["memory"] += _parse_memory(memory_request)
    except Exception as e:
        logger.error(f"Failed to get pods for node resources: {e}")

    node_metrics = {}
    try:
        metrics = custom_api.list_cluster_custom_object("metrics.k8s.io", "v1", "nodes")
        for metric in metrics.get("items", []):
            node_name = metric.get("metadata", {}).get("name")
            usage = metric.get("usage", {})
            node_metrics[node_name] = {
                "cpu": _parse_cpu(usage.get("cpu", "0")),
                "memory": _parse_memory(usage.get("memory", "0")),
            }
    except Exception as e:
        logger.warning(f"Failed to get node metrics from metrics-server: {e}")

    try:
        node_list = core_v1.list_node()
        for node in node_list.items:
            node_name = node.metadata.name

            cpu_capacity = 0
            memory_capacity = 0
            if node.status.capacity:
                cpu_capacity = int(node.status.capacity.get("cpu", "0"))
                memory_capacity = _parse_memory(node.status.capacity.get("memory", "0"))

            cpu_usage = 0
            memory_usage = 0

            if node.status.allocatable:
                allocatable_cpu = int(node.status.allocatable.get("cpu", "0"))
                allocatable_memory = _parse_memory(
                    node.status.allocatable.get("memory", "0")
                )

                cpu_capacity = allocatable_cpu * 1000
                memory_capacity = allocatable_memory

            if node_name in node_metrics:
                cpu_usage = node_metrics[node_name]["cpu"]
                memory_usage = node_metrics[node_name]["memory"]
            elif node_name in pod_resources:
                cpu_usage = pod_resources[node_name]["cpu"]
                memory_usage = pod_resources[node_name]["memory"]

            pods_count = pod_resources.get(node_name, {}).get("pods", 0)

            conditions = {c.type: c.status for c in node.status.conditions}
            status = "Ready" if conditions.get("Ready") == "True" else "NotReady"

            nodes[node_name] = {
                "name": node_name,
                "status": status,
                "cpu_capacity": cpu_capacity,
                "memory_capacity": memory_capacity,
                "cpu_usage": cpu_usage,
                "memory_usage": memory_usage,
                "pods_count": pods_count,
            }
    except Exception as e:
        logger.error(f"Failed to list nodes: {e}")

    return {"total": len(nodes), "nodes": list(nodes.values())}


@router.get("/deployments")
def list_deployments():
    """获取部署列表"""
    k8s = _get_k8s_client()
    apps_v1 = k8s.AppsV1Api()

    deployments = []
    try:
        deployment_list = apps_v1.list_deployment_for_all_namespaces()
        for deployment in deployment_list.items:
            deployment_name = deployment.metadata.name
            namespace = deployment.metadata.namespace

            replicas = deployment.spec.replicas or 0
            available_replicas = deployment.status.available_replicas or 0

            status = "运行中" if available_replicas > 0 else "未运行"

            deployments.append(
                {
                    "name": deployment_name,
                    "namespace": namespace,
                    "replicas": replicas,
                    "available_replicas": available_replicas,
                    "status": status,
                    "age": _format_age(deployment.metadata.creation_timestamp),
                    "labels": deployment.metadata.labels or {},
                }
            )
    except Exception as e:
        logger.error(f"Failed to list deployments: {e}")
        return {"success": False, "message": f"获取部署列表失败: {str(e)}"}

    return {"total": len(deployments), "deployments": deployments}


@router.delete("/deployments/{namespace}/{deployment_name}")
def delete_deployment(namespace: str, deployment_name: str):
    """删除指定部署"""
    if namespace == "kube-system":
        return {"success": False, "message": "不能删除 kube-system 命名空间的部署"}

    k8s = _get_k8s_client()
    apps_v1 = k8s.AppsV1Api()

    try:
        apps_v1.delete_namespaced_deployment(
            name=deployment_name,
            namespace=namespace,
            body=client.V1DeleteOptions(propagation_policy="Foreground"),
        )
        logger.info(
            f"[K8s] Deleted deployment {deployment_name} in namespace {namespace}"
        )
        return {"success": True, "message": f"部署 {deployment_name} 已删除"}
    except Exception as e:
        logger.error(f"Failed to delete deployment {deployment_name}: {e}")
        return {"success": False, "message": f"删除失败: {str(e)}"}


@router.delete("/pods/{namespace}/{pod_name}")
def delete_pod(namespace: str, pod_name: str, db: Session = Depends(get_db)):
    """强制删除指定 Pod 及其所属实例（非 kube-system 命名空间）"""
    if namespace == "kube-system":
        return {"success": False, "message": "不能删除 kube-system 命名空间的 Pod"}

    k8s = _get_k8s_client()
    core_v1 = k8s.CoreV1Api()

    try:
        pod = core_v1.read_namespaced_pod(name=pod_name, namespace=namespace)

        labels = pod.metadata.labels or {}
        app_label = labels.get("app", "")

        if app_label.startswith("deploy-cs-"):
            instance_id_str = app_label.replace("deploy-cs-", "")
            success = instance_service.delete(db, instance_id_str)
            if success:
                logger.info(
                    f"[K8s] Deleted instance {instance_id_str} via pod {pod_name}"
                )
                return {"success": True, "message": f"实例 {instance_id_str} 已删除"}

        core_v1.delete_namespaced_pod(
            name=pod_name,
            namespace=namespace,
            body=client.V1DeleteOptions(propagation_policy="Foreground"),
        )
        logger.info(f"[K8s] Deleted pod {pod_name} in namespace {namespace}")
        return {"success": True, "message": f"Pod {pod_name} 已删除"}
    except Exception as e:
        logger.error(f"Failed to delete pod {pod_name}: {e}")
        return {"success": False, "message": f"删除失败: {str(e)}"}


@router.get("/pods/{namespace}/{pod_name}/logs")
def get_pod_logs(
    namespace: str,
    pod_name: str,
    container: Optional[str] = None,
    tail_lines: int = 50,
):
    """获取Pod日志"""
    k8s = _get_k8s_client()
    core_v1 = k8s.CoreV1Api()

    try:
        pod = core_v1.read_namespaced_pod(name=pod_name, namespace=namespace)

        container_name = container
        if not container_name:
            if pod.spec.containers:
                container_name = pod.spec.containers[0].name
            else:
                return {"success": False, "message": "Pod没有容器"}

        logs = core_v1.read_namespaced_pod_log(
            name=pod_name,
            namespace=namespace,
            container=container_name,
            tail_lines=tail_lines,
        )

        return {
            "success": True,
            "logs": logs,
            "container": container_name,
            "pod_name": pod_name,
            "namespace": namespace,
        }
    except Exception as e:
        logger.error(f"Failed to get pod logs for {pod_name}: {e}")
        return {"success": False, "message": f"获取日志失败: {str(e)}"}


@router.get("/pods/{namespace}/{pod_name}/describe")
def describe_pod(namespace: str, pod_name: str):
    """获取Pod详细信息"""
    k8s = _get_k8s_client()
    core_v1 = k8s.CoreV1Api()

    try:
        pod = core_v1.read_namespaced_pod(name=pod_name, namespace=namespace)

        info = {
            "name": pod.metadata.name,
            "namespace": pod.metadata.namespace,
            "node": pod.spec.node_name or "N/A",
            "status": pod.status.phase,
            "ip": pod.status.pod_ip or "N/A",
            "host_ip": pod.status.host_ip or "N/A",
            "start_time": pod.status.start_time.isoformat()
            if pod.status.start_time
            else None,
            "labels": pod.metadata.labels or {},
            "annotations": pod.metadata.annotations or {},
            "containers": [],
            "conditions": [],
            "events": [],
        }

        for container in pod.spec.containers:
            container_info = {
                "name": container.name,
                "image": container.image,
                "image_pull_policy": container.image_pull_policy,
                "command": " ".join(container.command) if container.command else None,
                "args": " ".join(container.args) if container.args else None,
                "ports": [
                    {"container_port": c.container_port, "protocol": c.protocol}
                    for c in container.ports
                ]
                if container.ports
                else [],
                "resources": {
                    "requests": container.resources.requests
                    if container.resources and container.resources.requests
                    else {},
                    "limits": container.resources.limits
                    if container.resources and container.resources.limits
                    else {},
                }
                if container.resources
                else {},
            }
            info["containers"].append(container_info)

        for condition in pod.status.conditions or []:
            info["conditions"].append(
                {
                    "type": condition.type,
                    "status": condition.status,
                    "reason": condition.reason,
                    "message": condition.message,
                    "last_transition_time": condition.last_transition_time.isoformat()
                    if condition.last_transition_time
                    else None,
                }
            )

        try:
            events = core_v1.list_namespaced_event(
                namespace=namespace,
                field_selector=f"involvedObject.name={pod_name},involvedObject.kind=Pod",
            )
            for event in events.items:
                info["events"].append(
                    {
                        "type": event.type,
                        "reason": event.reason,
                        "message": event.message,
                        "source": event.source.component if event.source else None,
                        "count": event.count,
                        "first_timestamp": event.first_timestamp.isoformat()
                        if event.first_timestamp
                        else None,
                        "last_timestamp": event.last_timestamp.isoformat()
                        if event.last_timestamp
                        else None,
                    }
                )
        except Exception as e:
            logger.warning(f"Failed to get events for pod {pod_name}: {e}")

        return {"success": True, "pod": info}
    except Exception as e:
        logger.error(f"Failed to describe pod {pod_name}: {e}")
        return {"success": False, "message": f"获取Pod详情失败: {str(e)}"}


@router.get("/pv")
def list_pv():
    """获取所有PV列表"""
    k8s = _get_k8s_client()
    core_v1 = k8s.CoreV1Api()

    pvs = []
    try:
        pv_list = core_v1.list_persistent_volume()
        for pv in pv_list.items:
            pvs.append(
                {
                    "name": pv.metadata.name,
                    "capacity": pv.spec.capacity,
                    "access_modes": pv.spec.access_modes,
                    "reclaim_policy": pv.spec.persistent_volume_reclaim_policy,
                    "status": pv.status.phase,
                    "claim": f"{pv.spec.claim_ref.namespace}/{pv.spec.claim_ref.name}"
                    if pv.spec.claim_ref
                    else None,
                    "storage_class": pv.spec.storage_class_name,
                    "age": _format_age(pv.metadata.creation_timestamp),
                }
            )
    except Exception as e:
        logger.error(f"Failed to list pv: {e}")
        return {"success": False, "message": f"获取PV列表失败: {str(e)}"}

    return {"total": len(pvs), "pvs": pvs}


@router.delete("/pv/{pv_name}")
def delete_pv(pv_name: str):
    """删除PV"""
    k8s = _get_k8s_client()
    core_v1 = k8s.CoreV1Api()

    try:
        pv = core_v1.read_persistent_volume(name=pv_name)
        if pv.spec.persistent_volume_reclaim_policy == "Retain":
            return {
                "success": False,
                "message": "PV的回收策略为Retain，请先手动处理数据",
            }

        if pv.spec.claim_ref:
            return {
                "success": False,
                "message": f"PV已绑定到PVC {pv.spec.claim_ref.namespace}/{pv.spec.claim_ref.name}，请先删除该PVC",
            }

        core_v1.delete_persistent_volume(
            name=pv_name, body=client.V1DeleteOptions(propagation_policy="Foreground")
        )
        logger.info(f"[K8s] Deleted pv {pv_name}")
        return {
            "success": True,
            "message": f"PV {pv_name} 已删除（可能需要几秒钟完全消失）",
        }
    except Exception as e:
        logger.error(f"Failed to delete pv {pv_name}: {e}")
        return {"success": False, "message": f"删除失败: {str(e)}"}


@router.get("/pvc")
def list_pvc(namespace: Optional[str] = None):
    """获取所有PVC列表"""
    k8s = _get_k8s_client()
    core_v1 = k8s.CoreV1Api()

    pvcs = []
    namespaces = [namespace] if namespace else []

    if not namespace:
        try:
            ns_list = core_v1.list_namespace()
            namespaces = [ns.metadata.name for ns in ns_list.items]
        except Exception:
            namespaces = ["default"]

    for ns in namespaces:
        try:
            pvc_list = core_v1.list_namespaced_persistent_volume_claim(namespace=ns)
            for pvc in pvc_list.items:
                pvcs.append(
                    {
                        "name": pvc.metadata.name,
                        "namespace": ns,
                        "storage_class": pvc.spec.storage_class_name,
                        "capacity": pvc.status.capacity
                        if pvc.status and pvc.status.capacity
                        else {},
                        "access_modes": pvc.spec.access_modes,
                        "volume_mode": pvc.spec.volume_mode,
                        "status": pvc.status.phase if pvc.status else "Unknown",
                        "volume": pvc.spec.volume_name,
                        "age": _format_age(pvc.metadata.creation_timestamp),
                    }
                )
        except Exception as e:
            logger.error(f"Failed to list pvc in namespace {ns}: {e}")
            continue

    return {"total": len(pvcs), "pvcs": pvcs}


@router.delete("/pvc/{namespace}/{pvc_name}")
def delete_pvc(namespace: str, pvc_name: str):
    """删除PVC"""
    if namespace == "kube-system":
        return {"success": False, "message": "不能删除 kube-system 命名空间的 PVC"}

    k8s = _get_k8s_client()
    core_v1 = k8s.CoreV1Api()

    try:
        core_v1.delete_namespaced_persistent_volume_claim(
            name=pvc_name,
            namespace=namespace,
            body=client.V1DeleteOptions(propagation_policy="Foreground"),
        )
        logger.info(f"[K8s] Deleted pvc {pvc_name} in namespace {namespace}")
        return {"success": True, "message": f"PVC {pvc_name} 已删除"}
    except Exception as e:
        logger.error(f"Failed to delete pvc {pvc_name}: {e}")
        return {"success": False, "message": f"删除失败: {str(e)}"}
