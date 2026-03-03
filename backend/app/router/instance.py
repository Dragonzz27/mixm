from datetime import datetime, timezone
from typing import Optional

from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session
import logging
from pydantic import BaseModel
from app.database import get_db
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
def get_queue_status():
    """获取队列整体状态"""
    return queue_service.get_queue_status()


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
