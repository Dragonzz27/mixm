import logging
from typing import Optional

from app.database import SessionLocal
from app.schema.instance import InstanceCreate
from app.service.instance import InstanceService
from app.service.queue_store import TaskStatus, get_redis_client, update_task_status
from app.celery_app import celery_app

logger = logging.getLogger(__name__)


@celery_app.task(bind=True, name="instance.create")
def create_instance_task(
    self,
    instance_name: str,
    user_id: int,
    namespace: str,
    repo_id: Optional[int] = None,
    repo_tag: Optional[str] = None,
    cpu_limit: float = 1.0,
    memory_limit: str = "1Gi",
    storage_limit: int = 1,
    gpu_limit: int = 0,
    gpu_memory_limit: int = 0,
):
    task_id = self.request.id
    if not task_id:
        raise RuntimeError("Celery task id is missing")

    redis_client = get_redis_client()
    update_task_status(redis_client, task_id, TaskStatus.PROCESSING)

    db = SessionLocal()
    try:
        instance_data = InstanceCreate(
            name=instance_name,
            user_id=user_id,
            namespace=namespace,
        )

        instance_service = InstanceService()
        result = instance_service.create(
            db,
            instance_data,
            repo_id=repo_id,
            repo_tag=repo_tag,
            cpu_limit=cpu_limit,
            memory_limit=memory_limit,
            storage_limit=storage_limit,
            gpu_limit=gpu_limit,
            gpu_memory_limit=gpu_memory_limit,
        )

        update_task_status(
            redis_client,
            task_id,
            TaskStatus.COMPLETED,
            instance_id=str(result.id),
        )
        return {"instance_id": str(result.id)}
    except Exception as exc:
        error_message = str(exc)
        update_task_status(
            redis_client,
            task_id,
            TaskStatus.FAILED,
            error=error_message,
        )
        logger.error("[Queue] Task %s failed: %s", task_id, error_message)
        raise
    finally:
        db.close()
