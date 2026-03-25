from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session

from app.database import get_db
from app.models.role import Role
from app.models.user_role import UserRole
from app.schema.role import (
    PermissionOption,
    RoleCreate,
    RoleListResponse,
    RoleResponse,
    RoleUpdate,
)
from app.service.router import RouterService

router = APIRouter()
router_service = RouterService()


@router.get("/permissions", response_model=list[PermissionOption])
def get_permission_options():
    return router_service.get_permission_options()


@router.post("/", response_model=RoleResponse, status_code=status.HTTP_201_CREATED)
def create_role(payload: RoleCreate, db: Session = Depends(get_db)):
    existing = db.query(Role).filter(Role.name == payload.name).first()
    if existing:
        raise HTTPException(status_code=400, detail="Role name already exists")

    role = Role(name=payload.name, route_keys=payload.route_keys)
    db.add(role)
    db.commit()
    db.refresh(role)
    return role


@router.get("/", response_model=RoleListResponse)
def get_roles(skip: int = 0, limit: int = 100, keyword: str = None, db: Session = Depends(get_db)):
    query = db.query(Role)
    if keyword:
        keyword = keyword.strip()
        if keyword:
            query = query.filter(Role.name.contains(keyword))

    total = query.count()
    roles = query.order_by(Role.created_at.desc()).offset(skip).limit(limit).all()
    return {"total": total, "roles": roles}


@router.get("/{role_id}", response_model=RoleResponse)
def get_role(role_id: str, db: Session = Depends(get_db)):
    role = db.query(Role).filter(Role.id == role_id).first()
    if not role:
        raise HTTPException(status_code=404, detail="Role not found")
    return role


@router.put("/{role_id}", response_model=RoleResponse)
def update_role(role_id: str, payload: RoleUpdate, db: Session = Depends(get_db)):
    role = db.query(Role).filter(Role.id == role_id).first()
    if not role:
        raise HTTPException(status_code=404, detail="Role not found")

    duplicate = (
        db.query(Role)
        .filter(Role.name == payload.name, Role.id != role_id)
        .first()
    )
    if duplicate:
        raise HTTPException(status_code=400, detail="Role name already exists")

    role.name = payload.name
    role.route_keys = payload.route_keys
    db.commit()
    db.refresh(role)
    return role


@router.delete("/{role_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_role(role_id: str, db: Session = Depends(get_db)):
    role = db.query(Role).filter(Role.id == role_id).first()
    if not role:
        raise HTTPException(status_code=404, detail="Role not found")
    db.query(UserRole).filter(UserRole.role_id == role_id).delete()
    db.delete(role)
    db.commit()
