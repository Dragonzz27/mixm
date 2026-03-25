import logging
import uuid
from datetime import datetime
from typing import Any, Dict, List, Optional

from app.celery_app import celery_app
from app.service.queue_store import (
    QueueTask,
    TaskStatus,
    create_task_record,
    get_all_task_records,
    get_redis_client,
    get_task_record,
    get_user_task_records,
)

logger = logging.getLogger(__name__)


class InstanceQueueService:
    def __init__(self):
        self._redis = get_redis_client()

    def add_task(
        self,
        instance_name: str,
        user_id: int,
        namespace: str,
        username: str = "",
        repo_id: Optional[int] = None,
        repo_tag: Optional[str] = None,
        cpu_limit: float = 1.0,
        memory_limit: str = "1Gi",
        storage_limit: int = 1,
        gpu_limit: int = 0,
        gpu_memory_limit: int = 0,
    ) -> QueueTask:
        payload = {
            "instance_name": instance_name,
            "user_id": user_id,
            "namespace": namespace,
            "repo_id": repo_id,
            "repo_tag": repo_tag,
            "cpu_limit": cpu_limit,
            "memory_limit": memory_limit,
            "storage_limit": storage_limit,
            "gpu_limit": gpu_limit,
            "gpu_memory_limit": gpu_memory_limit,
        }

        task_id = uuid.uuid4().hex
        task_record = {
            "task_id": task_id,
            "instance_name": instance_name,
            "user_id": user_id,
            "namespace": namespace,
            "username": username,
            "repo_id": repo_id,
            "repo_tag": repo_tag,
            "cpu_limit": cpu_limit,
            "memory_limit": memory_limit,
            "storage_limit": storage_limit,
            "gpu_limit": gpu_limit,
            "gpu_memory_limit": gpu_memory_limit,
            "status": TaskStatus.PENDING.value,
            "instance_id": "",
            "error": "",
            "created_at": datetime.utcnow().isoformat(),
            "started_at": "",
            "completed_at": "",
        }
        create_task_record(self._redis, task_record)

        celery_app.send_task(
            "instance.create",
            kwargs=payload,
            task_id=task_id,
            queue="instance_create",
        )

        logger.info("[Queue] Task %s added to MQ queue: %s", task_id, instance_name)
        return QueueTask(task_record)

    def get_task(self, task_id: str) -> Optional[QueueTask]:
        return get_task_record(self._redis, task_id)

    def get_user_tasks(self, user_id: int) -> List[QueueTask]:
        return get_user_task_records(self._redis, user_id)

    def get_all_tasks(self) -> List[QueueTask]:
        return get_all_task_records(self._redis)

    def get_queue_position(self, task_id: str) -> int:
        tasks = self.get_all_tasks()
        pending_or_processing = [
            task
            for task in tasks
            if task.status in (TaskStatus.PENDING, TaskStatus.PROCESSING)
        ]
        for index, task in enumerate(reversed(pending_or_processing)):
            if task.task_id == task_id:
                return index
        return -1

    def _build_status(self, tasks: List[QueueTask]) -> Dict[str, Any]:
        pending = sum(1 for task in tasks if task.status == TaskStatus.PENDING)
        processing = sum(1 for task in tasks if task.status == TaskStatus.PROCESSING)
        completed = sum(1 for task in tasks if task.status == TaskStatus.COMPLETED)
        failed = sum(1 for task in tasks if task.status == TaskStatus.FAILED)

        return {
            "pending": pending,
            "processing": processing,
            "completed": completed,
            "failed": failed,
            "total": len(tasks),
        }

    def get_queue_status(self, user_id: Optional[int] = None) -> Dict[str, Any]:
        tasks = self.get_user_tasks(user_id) if user_id is not None else self.get_all_tasks()
        return self._build_status(tasks)

    def clear_completed(self, older_than_hours: int = 24):
        # Distributed queue mode keeps history in Redis for tracing;
        # clean-up policy can be added with key TTL or periodic task.
        return None


queue_service = InstanceQueueService()
