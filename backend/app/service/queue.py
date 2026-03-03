import logging
import uuid
import threading
import time
from datetime import datetime
from typing import List, Optional, Dict, Any
from enum import Enum

logger = logging.getLogger(__name__)


class TaskStatus(str, Enum):
    PENDING = "pending"
    PROCESSING = "processing"
    COMPLETED = "completed"
    FAILED = "failed"


class QueueTask:
    def __init__(
        self,
        task_id: str,
        instance_name: str,
        template_id: int,
        user_id: int,
        namespace: str,
        template_name: str = "",
        username: str = "",
    ):
        self.task_id = task_id
        self.instance_name = instance_name
        self.template_id = template_id
        self.user_id = user_id
        self.namespace = namespace
        self.template_name = template_name
        self.username = username
        self.status = TaskStatus.PENDING
        self.instance_id: Optional[int] = None
        self.error: Optional[str] = None
        self.created_at = datetime.now()
        self.started_at: Optional[datetime] = None
        self.completed_at: Optional[datetime] = None

    def to_dict(self) -> Dict[str, Any]:
        return {
            "task_id": self.task_id,
            "instance_name": self.instance_name,
            "template_id": self.template_id,
            "user_id": self.user_id,
            "namespace": self.namespace,
            "template_name": self.template_name,
            "username": self.username,
            "status": self.status.value,
            "instance_id": self.instance_id,
            "error": self.error,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "started_at": self.started_at.isoformat() if self.started_at else None,
            "completed_at": self.completed_at.isoformat()
            if self.completed_at
            else None,
        }


class InstanceQueueService:
    _instance = None
    _lock = threading.Lock()

    def __new__(cls):
        if cls._instance is None:
            with cls._lock:
                if cls._instance is None:
                    cls._instance = super().__new__(cls)
                    cls._instance._initialized = False
        return cls._instance

    def __init__(self):
        if self._initialized:
            return
        self._initialized = True
        self._queue: List[QueueTask] = []
        self._processing_thread: Optional[threading.Thread] = None
        self._stop_event = threading.Event()
        self._start_worker()

    def _start_worker(self):
        self._processing_thread = threading.Thread(
            target=self._process_queue, daemon=True
        )
        self._processing_thread.start()
        logger.info("[Queue] Instance queue worker started")

    def _process_queue(self):
        while not self._stop_event.is_set():
            task = self._get_next_pending_task()
            if task:
                self._execute_task(task)
            else:
                time.sleep(2)

    def _get_next_pending_task(self) -> Optional[QueueTask]:
        with self._lock:
            for t in self._queue:
                if t.status == TaskStatus.PENDING:
                    return t
            return None

    def _execute_task(self, task: QueueTask):
        task.status = TaskStatus.PROCESSING
        task.started_at = datetime.now()
        logger.info(
            f"[Queue] Processing task {task.task_id} for instance {task.instance_name}"
        )

        from app.database import SessionLocal
        from app.schema.instance import InstanceCreate

        db = SessionLocal()
        try:
            instance_data = InstanceCreate(
                name=task.instance_name,
                template_id=task.template_id,
                user_id=task.user_id,
                namespace=task.namespace,
            )

            from app.service.instance import InstanceService

            instance_service = InstanceService()
            result = instance_service.create(db, instance_data)

            task.instance_id = result.id
            task.status = TaskStatus.COMPLETED
            task.completed_at = datetime.now()
            logger.info(
                f"[Queue] Task {task.task_id} completed, instance ID: {result.id}"
            )

        except Exception as e:
            task.status = TaskStatus.FAILED
            task.error = str(e)
            task.completed_at = datetime.now()
            logger.error(f"[Queue] Task {task.task_id} failed: {str(e)}")
        finally:
            db.close()

    def add_task(
        self,
        instance_name: str,
        template_id: int,
        user_id: int,
        namespace: str,
        template_name: str = "",
        username: str = "",
    ) -> QueueTask:
        task_id = str(uuid.uuid4())[:8]
        task = QueueTask(
            task_id=task_id,
            instance_name=instance_name,
            template_id=template_id,
            user_id=user_id,
            namespace=namespace,
            template_name=template_name,
            username=username,
        )
        with self._lock:
            self._queue.append(task)
        logger.info(f"[Queue] Task {task_id} added to queue: {instance_name}")
        return task

    def get_task(self, task_id: str) -> Optional[QueueTask]:
        with self._lock:
            for task in self._queue:
                if task.task_id == task_id:
                    return task
        return None

    def get_user_tasks(self, user_id: int) -> List[QueueTask]:
        with self._lock:
            return [t for t in self._queue if t.user_id == user_id]

    def get_all_tasks(self) -> List[QueueTask]:
        with self._lock:
            return list(self._queue)

    def get_queue_position(self, task_id: str) -> int:
        with self._lock:
            for i, task in enumerate(self._queue):
                if task.task_id == task_id:
                    pending_count = 0
                    for t in self._queue[:i]:
                        if t.status in (TaskStatus.PENDING, TaskStatus.PROCESSING):
                            pending_count += 1
                    return pending_count
            return -1

    def get_queue_status(self) -> Dict[str, Any]:
        with self._lock:
            pending = sum(1 for t in self._queue if t.status == TaskStatus.PENDING)
            processing = sum(
                1 for t in self._queue if t.status == TaskStatus.PROCESSING
            )
            completed = sum(1 for t in self._queue if t.status == TaskStatus.COMPLETED)
            failed = sum(1 for t in self._queue if t.status == TaskStatus.FAILED)
            return {
                "pending": pending,
                "processing": processing,
                "completed": completed,
                "failed": failed,
                "total": len(self._queue),
            }

    def clear_completed(self, older_than_hours: int = 24):
        from datetime import timedelta

        cutoff = datetime.now() - timedelta(hours=older_than_hours)
        with self._lock:
            self._queue = [
                t
                for t in self._queue
                if t.status in (TaskStatus.PENDING, TaskStatus.PROCESSING)
                or (t.completed_at and t.completed_at > cutoff)
            ]


queue_service = InstanceQueueService()
