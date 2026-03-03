from datetime import datetime

from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session
from typing import List, Dict, Any
import logging
from pydantic import BaseModel
from app.database import get_db
from app.models.instance import Instance
from app.models.user import User
from app.models.template import Template
from app.schema.instance import (
    InstanceCreate,
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


class InstanceStartRequest(BaseModel):
    template_id: int
    name: str = ""


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
def start_instance_by_template(
    request: InstanceStartRequest,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    """通过模板ID启动实例，当前登录用户自动关联（异步队列模式）"""
    try:
        template = db.query(Template).filter(Template.id == request.template_id).first()
        if not template:
            raise HTTPException(status_code=404, detail="Template not found")

        now_timestamp = int(datetime.utcnow().timestamp())
        instance_name = (
            f"{request.name}-{now_timestamp}"
            if request.name
            else f"instance-{now_timestamp}"
        )

        task = queue_service.add_task(
            instance_name=instance_name,
            template_id=request.template_id,
            user_id=current_user.id,
            namespace="default",
            template_name=template.name,
            username=current_user.username,
        )

        logger.info(
            f"[Queue] Task {task.task_id} added for instance '{instance_name}' by user {current_user.username}"
        )

        return {
            "task_id": task.task_id,
            "instance_name": instance_name,
            "status": task.status.value,
            "message": "实例已加入创建队列，请稍后刷新查看进度",
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
def get_instance(instance_id: int, db: Session = Depends(get_db)):
    db_instance = instance_service.get_by_id(db, instance_id)
    if not db_instance:
        raise HTTPException(status_code=404, detail="Instance not found")
    return instance_service.to_response(db, db_instance)


@router.put("/{instance_id}", response_model=InstanceResponse)
def update_instance(
    instance_id: int, instance: InstanceUpdate, db: Session = Depends(get_db)
):
    # TODO: 调用 K8s API 更新资源 (如滚动更新)
    # logger.info(f"[K8s] Updating instance {instance_id}")
    # 从 K8s 获取最新状态并同步...

    db_instance = instance_service.update(db, instance_id, instance)
    if not db_instance:
        raise HTTPException(status_code=404, detail="Instance not found")
    return instance_service.to_response(db, db_instance)


@router.post("/{instance_id}/start", response_model=InstanceResponse)
def start_instance(instance_id: int, db: Session = Depends(get_db)):
    # TODO: 调用 K8s API 启动资源
    # logger.info(f"[K8s] Starting instance {instance_id}")
    # from kubernetes import client, config
    # config.load_kube_config()
    # apps_v1 = client.AppsV1Api()
    # deployment = apps_v1.read_namespaced_deployment(name=..., namespace=...)
    # apps_v1.patch_namespaced_deployment_scale(...) 或 rollout restart...

    db_instance = instance_service.update_status(db, instance_id, "running")
    if not db_instance:
        raise HTTPException(status_code=404, detail="Instance not found")
    logger.info(f"[Instance] Instance {instance_id} started")
    return instance_service.to_response(db, db_instance)


@router.post("/{instance_id}/stop", response_model=InstanceResponse)
def stop_instance(instance_id: int, db: Session = Depends(get_db)):
    # TODO: 调用 K8s API 停止资源
    # logger.info(f"[K8s] Stopping instance {instance_id}")
    # from kubernetes import client, config
    # config.load_kube_config()
    # apps_v1 = client.AppsV1Api()
    # scale=0 缩容或删除 Deployment...

    db_instance = instance_service.update_status(db, instance_id, "stopped")
    if not db_instance:
        raise HTTPException(status_code=404, detail="Instance not found")
    logger.info(f"[Instance] Instance {instance_id} stopped")
    return instance_service.to_response(db, db_instance)


@router.delete("/{instance_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_instance(instance_id: int, db: Session = Depends(get_db)):
    # TODO: 调用 K8s API 删除资源
    # logger.info(f"[K8s] Deleting instance {instance_id}")
    # from kubernetes import client, config
    # config.load_kube_config()
    # apps_v1 = client.AppsV1Api()
    # apps_v1.delete_namespaced_deployment(name=..., namespace=...)...

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


@router.get("/template/{template_id}", response_model=InstanceListResponse)
def get_template_instances(template_id: int, db: Session = Depends(get_db)):
    instances = instance_service.get_by_template(db, template_id)
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


# TODO: 后续可添加以下 K8s 监控接口
# @router.get("/{instance_id}/logs") - 获取容器日志
# @router.get("/{instance_id}/events") - 获取事件列表
# @router.post("/{instance_id}/restart") - 重启实例
# @router.get("/{instance_id}/status") - 获取 K8s 中的实时状态
