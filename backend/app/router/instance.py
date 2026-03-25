import json
from datetime import datetime, timezone
from typing import Optional

from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session
import logging
from pydantic import BaseModel
from app.database import get_db
from app.models.template import Template
from app.models.user import User
from app.schema.instance import (
    InstanceCreate,
    InstanceCreateWithConfig,
    InstanceUpdate,
    InstanceResponse,
    InstanceListResponse,
)
from app.service.instance import InstanceService
from app.service.auth import AuthService
from app.service.queue import queue_service

logger = logging.getLogger(__name__)

router = APIRouter()
instance_service = InstanceService()
auth_service = AuthService()


def get_current_user(token: str = Depends(auth_service.get_current_user)):
    return token


class BatchStartByTemplateRequest(BaseModel):
    template_id: int
    user_ids: list[int]


def _parse_template_config(raw_config: Optional[str]) -> dict:
    if not raw_config:
        return {}
    if isinstance(raw_config, dict):
        return raw_config
    try:
        return json.loads(raw_config)
    except Exception:
        return {}


def _parse_storage_limit_gi(value, default: int = 1) -> int:
    if value is None:
        return default
    if isinstance(value, (int, float)):
        return int(value)

    text = str(value).strip()
    if not text:
        return default

    upper_text = text.upper()
    for suffix in ("GI", "G"):
        if upper_text.endswith(suffix):
            number_part = text[: -len(suffix)].strip()
            return int(float(number_part))

    return int(float(text))


@router.post("/", response_model=InstanceResponse, status_code=status.HTTP_201_CREATED)
def create_instance(instance: InstanceCreate, db: Session = Depends(get_db)):
    try:
        result = instance_service.create(db, instance)
        logger.info(
            f"[Instance] Instance '{instance.name}' created successfully, ID: {result.id}"
        )
        return instance_service.to_response(db, result)
    except ValueError as e:
        logger.error(
            f"[Instance] Failed to create instance '{instance.name}': {str(e)}"
        )
        raise HTTPException(status_code=400, detail=str(e))
    except Exception as e:
        logger.error(
            f"[Instance] Failed to create instance '{instance.name}': {str(e)}"
        )
        raise HTTPException(status_code=500, detail=f"启动实例失败: {str(e)}")


@router.post("/start", status_code=status.HTTP_202_ACCEPTED)
def start_instance_direct(
    request: InstanceCreateWithConfig,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    """直接启动实例（异步队列模式）：可指定 user_id，否则使用当前登录用户"""
    try:
        target_user = current_user
        if request.user_id is not None:

            db_user = db.query(User).filter(User.id == request.user_id).first()
            if not db_user:
                raise HTTPException(status_code=404, detail="指定用户不存在")
            target_user = db_user

        now_timestamp = int(datetime.now(timezone.utc).timestamp())
        instance_name = (
            f"{request.name}-{now_timestamp}"
            if request.name
            else f"instance-{now_timestamp}"
        )

        task = queue_service.add_task(
            instance_name=instance_name,
            user_id=target_user.id,
            namespace="default",
            username=target_user.username,
            repo_id=request.repo_id,
            repo_tag=request.repo_tag,
            cpu_limit=request.cpu_limit,
            memory_limit=request.memory_limit,
            storage_limit=request.storage_limit,
            gpu_limit=request.gpu_limit,
            gpu_memory_limit=request.gpu_memory_limit,
        )

        logger.info(
            f"[Queue] Task {task.task_id} added for instance '{instance_name}' by user {target_user.username}"
        )

        return {
            "task_id": task.task_id,
            "instance_name": instance_name,
            "user_id": target_user.id,
            "status": task.status.value,
            "message": "实例已加入创建队列，请稍后刷新查看进度",
            "config": {
                "repo_id": request.repo_id,
                "repo_tag": request.repo_tag,
                "cpu_limit": request.cpu_limit,
                "memory_limit": request.memory_limit,
                "storage_limit": request.storage_limit,
                "gpu_limit": request.gpu_limit,
                "gpu_memory_limit": request.gpu_memory_limit,
            },
        }
    except ValueError as e:
        logger.error(f"[Instance] Failed to create instance '{request.name}': {str(e)}")
        raise HTTPException(status_code=400, detail=str(e))
    except Exception as e:
        logger.error(f"[Instance] Failed to create instance '{request.name}': {str(e)}")
        raise HTTPException(status_code=500, detail=f"启动实例失败: {str(e)}")


@router.post("/start/batch-template", status_code=status.HTTP_202_ACCEPTED)
def start_instances_by_template_batch(
    request: BatchStartByTemplateRequest,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    """按模板批量启动实例（加入MQ队列）"""
    if not current_user.is_superuser:
        raise HTTPException(status_code=403, detail="仅管理员可执行批量启动")

    if not request.user_ids:
        raise HTTPException(status_code=400, detail="user_ids 不能为空")

    template = db.query(Template).filter(Template.id == request.template_id).first()
    if not template:
        raise HTTPException(status_code=404, detail="模板不存在")

    template_config = _parse_template_config(template.config)
    repo_id = template_config.get("repo_id")
    repo_tag = template_config.get("repo_tag")

    if not repo_id or not repo_tag:
        raise HTTPException(
            status_code=400,
            detail="模板配置缺少 repo_id 或 repo_tag，无法批量启动",
        )

    user_ids = sorted(set(request.user_ids))
    users = db.query(User).filter(User.id.in_(user_ids)).all()
    user_map = {u.id: u for u in users}
    missing_user_ids = [uid for uid in user_ids if uid not in user_map]
    if missing_user_ids:
        raise HTTPException(
            status_code=404,
            detail=f"以下用户不存在: {missing_user_ids}",
        )

    try:
        cpu_limit = float(template_config.get("cpu_limit", 1.0))
        memory_limit = template_config.get("memory_limit") or template_config.get(
            "memory_limit_str", "1Gi"
        )
        storage_limit = _parse_storage_limit_gi(template_config.get("storage", 1))
        gpu_limit = int(template_config.get("gpu_limit", 0))
        gpu_memory_limit = int(template_config.get("gpu_memory_limit", 0))

        # Align validation behavior with single-start API.
        InstanceCreateWithConfig(
            name="batch-validate",
            repo_id=int(repo_id),
            repo_tag=str(repo_tag),
            cpu_limit=cpu_limit,
            memory_limit=str(memory_limit),
            storage_limit=storage_limit,
            gpu_limit=gpu_limit,
            gpu_memory_limit=gpu_memory_limit,
        )
    except Exception as e:
        raise HTTPException(
            status_code=400,
            detail=f"模板资源配置不合法: {str(e)}",
        )

    queued_tasks = []
    now_timestamp = int(datetime.now(timezone.utc).timestamp())
    for user_id in user_ids:
        target_user = user_map[user_id]
        instance_name = f"{template.name}-{target_user.username}-{now_timestamp}"

        task = queue_service.add_task(
            instance_name=instance_name,
            user_id=target_user.id,
            namespace="default",
            username=target_user.username,
            repo_id=repo_id,
            repo_tag=repo_tag,
            cpu_limit=cpu_limit,
            memory_limit=memory_limit,
            storage_limit=storage_limit,
            gpu_limit=gpu_limit,
            gpu_memory_limit=gpu_memory_limit,
        )

        queued_tasks.append(
            {
                "task_id": task.task_id,
                "user_id": target_user.id,
                "username": target_user.username,
                "instance_name": instance_name,
                "status": task.status.value,
            }
        )

    return {
        "template_id": request.template_id,
        "queued_count": len(queued_tasks),
        "tasks": queued_tasks,
        "message": f"已批量加入 {len(queued_tasks)} 个实例创建任务",
    }


@router.get("/", response_model=InstanceListResponse)
def list_instances(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    instances = instance_service.get_all(db, skip, limit)
    total = instance_service.get_count(db)
    return {
        "total": total,
        "instances": [instance_service.to_response(db, i) for i in instances],
    }


@router.get("/me", response_model=InstanceListResponse)
def get_my_instances(
    db: Session = Depends(get_db), current_user: User = Depends(get_current_user)
):
    """获取当前登录用户的实例"""
    instances = instance_service.get_by_user(db, current_user.id)
    return {
        "total": len(instances),
        "instances": [instance_service.to_response(db, i) for i in instances],
    }


@router.get("/{instance_id}", response_model=InstanceResponse)
def get_instance(instance_id: str, db: Session = Depends(get_db)):
    db_instance = instance_service.get_by_id(db, instance_id)
    if not db_instance:
        raise HTTPException(status_code=404, detail="Instance not found")
    return instance_service.to_response(db, db_instance)


@router.put("/{instance_id}", response_model=InstanceResponse)
def update_instance(
    instance_id: str, instance: InstanceUpdate, db: Session = Depends(get_db)
):
    db_instance = instance_service.update(db, instance_id, instance)
    if not db_instance:
        raise HTTPException(status_code=404, detail="Instance not found")
    return instance_service.to_response(db, db_instance)


@router.post("/{instance_id}/start", response_model=InstanceResponse)
def start_instance(instance_id: str, db: Session = Depends(get_db)):
    db_instance = instance_service.update_status(db, instance_id, "running")
    if not db_instance:
        raise HTTPException(status_code=404, detail="Instance not found")
    logger.info(f"[Instance] Instance {instance_id} started")
    return instance_service.to_response(db, db_instance)


@router.post("/{instance_id}/stop", response_model=InstanceResponse)
def stop_instance(instance_id: str, db: Session = Depends(get_db)):
    db_instance = instance_service.update_status(db, instance_id, "stopped")
    if not db_instance:
        raise HTTPException(status_code=404, detail="Instance not found")
    logger.info(f"[Instance] Instance {instance_id} stopped")
    return instance_service.to_response(db, db_instance)


@router.delete("/{instance_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_instance(instance_id: str, db: Session = Depends(get_db)):
    success = instance_service.delete(db, instance_id)
    if not success:
        raise HTTPException(status_code=404, detail="Instance not found")
    logger.info(f"[Instance] Instance {instance_id} deleted")


@router.get("/user/{user_id}", response_model=InstanceListResponse)
def get_user_instances(user_id: int, db: Session = Depends(get_db)):
    instances = instance_service.get_by_user(db, user_id)
    return {
        "total": len(instances),
        "instances": [instance_service.to_response(db, i) for i in instances],
    }


@router.get("/queue/status", status_code=status.HTTP_200_OK)
def get_queue_status(current_user: User = Depends(get_current_user)):
    """获取当前用户的队列状态"""
    return queue_service.get_queue_status(user_id=current_user.id)


@router.get("/queue/tasks", status_code=status.HTTP_200_OK)
def get_queue_tasks(current_user: User = Depends(get_current_user)):
    """获取当前用户的队列任务列表"""
    tasks = queue_service.get_user_tasks(current_user.id)
    return {"total": len(tasks), "tasks": [t.to_dict() for t in tasks]}


@router.get("/queue/task/{task_id}", status_code=status.HTTP_200_OK)
def get_task_status(task_id: str, current_user: User = Depends(get_current_user)):
    """获取指定任务的状态"""
    task = queue_service.get_task(task_id)
    if not task:
        raise HTTPException(status_code=404, detail="Task not found")

    if task.user_id != current_user.id:
        raise HTTPException(status_code=403, detail="Access denied")

    return task.to_dict()
