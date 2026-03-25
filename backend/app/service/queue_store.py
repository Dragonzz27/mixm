import json
from datetime import datetime
from enum import Enum
from typing import Any, Dict, List, Optional

from redis import Redis

from app.config import get_settings

TASK_KEY_PREFIX = "instance_task"
USER_TASKS_KEY_PREFIX = "instance_user_tasks"


class TaskStatus(str, Enum):
    PENDING = "pending"
    PROCESSING = "processing"
    COMPLETED = "completed"
    FAILED = "failed"


class QueueTask:
    def __init__(self, data: Dict[str, Any]):
        self.task_id = data.get("task_id")
        self.instance_name = data.get("instance_name")
        self.user_id = int(data.get("user_id")) if data.get("user_id") is not None else None
        self.namespace = data.get("namespace")
        self.username = data.get("username") or ""
        self.repo_id = _to_int(data.get("repo_id"))
        self.repo_tag = data.get("repo_tag")
        self.cpu_limit = _to_float(data.get("cpu_limit"), default=1.0)
        self.memory_limit = data.get("memory_limit") or "1Gi"
        self.storage_limit = _to_int(data.get("storage_limit"), default=1)
        self.gpu_limit = _to_int(data.get("gpu_limit"), default=0)
        self.gpu_memory_limit = _to_int(data.get("gpu_memory_limit"), default=0)
        self.status = TaskStatus(data.get("status") or TaskStatus.PENDING.value)
        self.instance_id = data.get("instance_id")
        self.error = data.get("error")
        self.created_at = data.get("created_at")
        self.started_at = data.get("started_at")
        self.completed_at = data.get("completed_at")

    def to_dict(self) -> Dict[str, Any]:
        return {
            "task_id": self.task_id,
            "instance_name": self.instance_name,
            "user_id": self.user_id,
            "namespace": self.namespace,
            "username": self.username,
            "repo_id": self.repo_id,
            "repo_tag": self.repo_tag,
            "cpu_limit": self.cpu_limit,
            "memory_limit": self.memory_limit,
            "storage_limit": self.storage_limit,
            "gpu_limit": self.gpu_limit,
            "gpu_memory_limit": self.gpu_memory_limit,
            "status": self.status.value,
            "instance_id": self.instance_id,
            "error": self.error,
            "created_at": self.created_at,
            "started_at": self.started_at,
            "completed_at": self.completed_at,
        }


def _to_int(value: Any, default: Optional[int] = None) -> Optional[int]:
    if value is None or value == "":
        return default
    try:
        return int(value)
    except (TypeError, ValueError):
        return default


def _to_float(value: Any, default: float = 1.0) -> float:
    if value is None or value == "":
        return default
    try:
        return float(value)
    except (TypeError, ValueError):
        return default


def _task_key(task_id: str) -> str:
    return f"{TASK_KEY_PREFIX}:{task_id}"


def _user_tasks_key(user_id: int) -> str:
    return f"{USER_TASKS_KEY_PREFIX}:{user_id}"


def _utc_now_iso() -> str:
    return datetime.utcnow().isoformat()


def get_redis_client() -> Redis:
    settings = get_settings()
    return Redis.from_url(settings.QUEUE_REDIS_URL, decode_responses=True)


def create_task_record(redis_client: Redis, task_data: Dict[str, Any]) -> None:
    task_id = task_data["task_id"]
    payload = dict(task_data)
    for key, value in payload.items():
        if value is None:
            payload[key] = ""
        elif isinstance(value, (dict, list)):
            payload[key] = json.dumps(value, ensure_ascii=True)

    redis_client.hset(_task_key(task_id), mapping=payload)
    redis_client.rpush(_user_tasks_key(int(task_data["user_id"])), task_id)


def get_task_record(redis_client: Redis, task_id: str) -> Optional[QueueTask]:
    data = redis_client.hgetall(_task_key(task_id))
    if not data:
        return None
    return QueueTask(data)


def get_user_task_records(redis_client: Redis, user_id: int) -> List[QueueTask]:
    task_ids = redis_client.lrange(_user_tasks_key(user_id), 0, -1)
    tasks: List[QueueTask] = []
    for task_id in reversed(task_ids):
        task = get_task_record(redis_client, task_id)
        if task:
            tasks.append(task)
    return tasks


def get_all_task_records(redis_client: Redis) -> List[QueueTask]:
    tasks: List[QueueTask] = []
    keys = redis_client.keys(f"{TASK_KEY_PREFIX}:*")
    for key in keys:
        task_id = key.split(":", 1)[1]
        task = get_task_record(redis_client, task_id)
        if task:
            tasks.append(task)
    tasks.sort(key=lambda item: item.created_at or "", reverse=True)
    return tasks


def update_task_status(
    redis_client: Redis,
    task_id: str,
    status: TaskStatus,
    instance_id: Optional[str] = None,
    error: Optional[str] = None,
) -> None:
    updates: Dict[str, Any] = {"status": status.value}
    if status == TaskStatus.PROCESSING:
        updates["started_at"] = _utc_now_iso()
    if status in (TaskStatus.COMPLETED, TaskStatus.FAILED):
        updates["completed_at"] = _utc_now_iso()
    if instance_id is not None:
        updates["instance_id"] = str(instance_id)
    if error is not None:
        updates["error"] = error
    redis_client.hset(_task_key(task_id), mapping=updates)
