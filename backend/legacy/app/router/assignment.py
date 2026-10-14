import os
import uuid
import json
from datetime import datetime
from typing import Optional
import csv
import io

from fastapi import (
    APIRouter,
    Depends,
    HTTPException,
    status,
    UploadFile,
    Request,
)
from fastapi.responses import Response, FileResponse, StreamingResponse
from sqlalchemy.orm import Session
from pydantic import BaseModel

from app.database import get_db
from app.models.assignment import Assignment, Submission
from app.models.user import User
from app.schema.assignment import (
    AssignmentCreate,
    AssignmentUpdate,
    AssignmentResponse,
    SubmissionCreate,
    SubmissionUpdate,
    SubmissionResponse,
    FileUploadResponse,
)
from app.service.assignment import assignment_service, submission_service
from app.service.auth import AuthService

auth_service = AuthService()
router = APIRouter()

get_current_user = auth_service.get_current_user

UPLOAD_DIR = "uploads/submissions"
ALLOWED_ATTACHMENT_EXTENSIONS = {
    ".mp4": "video/mp4",
    ".webm": "video/webm",
    ".ogg": "video/ogg",
    ".mov": "video/quicktime",
    ".m4v": "video/x-m4v",
    ".pdf": "application/pdf",
    ".doc": "application/msword",
    ".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
    ".ppt": "application/vnd.ms-powerpoint",
    ".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
    ".txt": "text/plain; charset=utf-8",
    ".md": "text/markdown; charset=utf-8",
}
INLINE_ATTACHMENT_EXTENSIONS = {
    ".mp4",
    ".webm",
    ".ogg",
    ".mov",
    ".m4v",
    ".pdf",
    ".txt",
    ".md",
}


class AssignmentListResponse(BaseModel):
    total: int
    assignments: list[AssignmentResponse]


class SubmissionListResponse(BaseModel):
    total: int
    submissions: list[SubmissionResponse]


def ensure_upload_dir():
    os.makedirs(UPLOAD_DIR, exist_ok=True)


def build_file_url(request: Optional[Request], file_path: Optional[str]) -> Optional[str]:
    if not request or not file_path:
        return None

    base_url = str(request.base_url).rstrip("/")
    storage_filename = os.path.basename(file_path)
    return f"{base_url}/api/assignments/file/{storage_filename}"


def parse_assignment_attachments(
    attachments_raw: Optional[str], request: Optional[Request] = None
) -> list[dict]:
    if not attachments_raw:
        return []

    try:
        attachments = json.loads(attachments_raw)
    except (TypeError, json.JSONDecodeError):
        return []

    normalized = []
    for item in attachments:
        if not isinstance(item, dict):
            continue
        name = (item.get("name") or "").strip()
        if not name:
            continue

        file_path = item.get("file_path")
        normalized.append(
            {
                "name": name,
                "file_path": file_path,
                "file_name": item.get("file_name"),
                "file_size": item.get("file_size"),
                "file_url": build_file_url(request, file_path),
            }
        )

    return normalized


def assignment_to_response(assignment: Assignment, request: Optional[Request] = None) -> dict:
    return {
        "id": assignment.id,
        "title": assignment.title,
        "description": assignment.description,
        "attachments": parse_assignment_attachments(assignment.attachments, request),
        "content_id": assignment.content_id,
        "deadline": assignment.deadline,
        "status": assignment.status,
        "visible_roles": [role.id for role in assignment.visible_roles]
        if assignment.visible_roles
        else [],
        "created_by": assignment.created_by,
        "created_at": assignment.created_at,
        "updated_at": assignment.updated_at,
        "submission_count": getattr(assignment, "submission_count", 0),
    }


@router.post(
    "/", response_model=AssignmentResponse, status_code=status.HTTP_201_CREATED
)
def create_assignment(
    assignment: AssignmentCreate,
    request: Request,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    db_assignment = assignment_service.create(db, assignment, current_user.id)
    return assignment_to_response(db_assignment, request)


@router.get("/", response_model=AssignmentListResponse)
def get_assignments(
    skip: int = 0,
    limit: int = 20,
    status: Optional[str] = None,
    request: Request = None,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    # 获取用户角色
    from app.models.user_role import UserRole

    user_roles = (
        db.query(UserRole.role_id).filter(UserRole.user_id == current_user.id).all()
    )
    role_list = [r[0] for r in user_roles] if user_roles else []

    assignments, total = assignment_service.get_list(
        db,
        skip,
        limit,
        status,
        user_roles=role_list,
        is_superuser=current_user.is_superuser,
    )
    return {
        "total": total,
        "assignments": [assignment_to_response(a, request) for a in assignments],
    }


@router.get("/active", response_model=list[AssignmentResponse])
def get_active_assignments(
    request: Request,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    # 获取用户角色
    from app.models.user_role import UserRole

    user_roles = (
        db.query(UserRole.role_id).filter(UserRole.user_id == current_user.id).all()
    )
    role_list = [r[0] for r in user_roles] if user_roles else []

    assignments = assignment_service.get_all_active(
        db, user_roles=role_list, is_superuser=current_user.is_superuser
    )
    return [assignment_to_response(a, request) for a in assignments]


@router.get("/my/submissions", response_model=list[SubmissionResponse])
def get_my_submissions(
    assignment_id: Optional[str] = None,
    request: Request = None,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    submissions = submission_service.get_by_user(db, current_user.id, assignment_id)

    if request:
        for sub in submissions:
            if sub.file_path:
                sub.file_url = build_file_url(request, sub.file_path)

    return submissions


@router.post("/upload")
async def upload_submission_file(
    file: UploadFile,
    request: Request,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    ensure_upload_dir()

    file_ext = os.path.splitext(file.filename)[1] if file.filename else ""
    file_ext = file_ext.lower()
    if file_ext not in ALLOWED_ATTACHMENT_EXTENSIONS:
        allowed_types = ", ".join(ALLOWED_ATTACHMENT_EXTENSIONS.keys())
        raise HTTPException(
            status_code=400,
            detail=f"Unsupported file type. Allowed types: {allowed_types}",
        )

    original_filename = file.filename or "unnamed"
    unique_id = str(uuid.uuid4())
    storage_filename = f"{unique_id}._.{original_filename}"
    file_path = os.path.join(UPLOAD_DIR, storage_filename)

    content = await file.read()
    file_size = len(content)

    with open(file_path, "wb") as f:
        f.write(content)

    base_url = str(request.base_url).rstrip("/")
    file_url = f"{base_url}/api/assignments/file/{storage_filename}"

    return FileUploadResponse(
        file_path=file_path,
        file_name=original_filename,
        file_size=file_size,
        file_url=file_url,
    )


@router.get("/file/{filename:path}")
async def get_submission_file(filename: str):
    file_path = os.path.join(UPLOAD_DIR, filename)
    if not os.path.exists(file_path):
        raise HTTPException(status_code=404, detail="File not found")

    original_name = filename.split("._.")[-1] if "._." in filename else filename
    file_ext = os.path.splitext(original_name)[1].lower()
    media_type = ALLOWED_ATTACHMENT_EXTENSIONS.get(file_ext, "application/octet-stream")
    content_disposition_type = (
        "inline" if file_ext in INLINE_ATTACHMENT_EXTENSIONS else "attachment"
    )

    return FileResponse(
        file_path,
        media_type=media_type,
        filename=original_name,
        content_disposition_type=content_disposition_type,
    )


@router.get("/{assignment_id}", response_model=AssignmentResponse)
def get_assignment(
    assignment_id: str,
    request: Request,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    assignment = assignment_service.get(db, assignment_id)
    if not assignment:
        raise HTTPException(status_code=404, detail="Assignment not found")
    return assignment_to_response(assignment, request)


@router.put("/{assignment_id}", response_model=AssignmentResponse)
def update_assignment(
    assignment_id: str,
    assignment: AssignmentUpdate,
    request: Request,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    db_assignment = assignment_service.update(db, assignment_id, assignment)
    if not db_assignment:
        raise HTTPException(status_code=404, detail="Assignment not found")
    return assignment_to_response(db_assignment, request)


@router.delete("/{assignment_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_assignment(
    assignment_id: str,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    success = assignment_service.delete(db, assignment_id)
    if not success:
        raise HTTPException(status_code=404, detail="Assignment not found")


@router.get("/{assignment_id}/submissions", response_model=SubmissionListResponse)
def get_submissions(
    assignment_id: str,
    skip: int = 0,
    limit: int = 20,
    request: Request = None,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    assignment = assignment_service.get(db, assignment_id)
    if not assignment:
        raise HTTPException(status_code=404, detail="Assignment not found")
    submissions, total = submission_service.get_by_assignment(
        db, assignment_id, skip, limit
    )

    for sub in submissions:
        if sub.file_path:
            sub.file_url = build_file_url(request, sub.file_path)

    return {"total": total, "submissions": submissions}


@router.get("/{assignment_id}/submissions/export")
def export_submissions(
    assignment_id: str,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    assignment = assignment_service.get(db, assignment_id)
    if not assignment:
        raise HTTPException(status_code=404, detail="Assignment not found")

    submissions, total = submission_service.get_by_assignment(
        db, assignment_id, 0, 1000
    )

    os.makedirs("static/export", exist_ok=True)
    filename = f"assignment_{assignment_id}_submissions_{datetime.now().strftime('%Y%m%d%H%M%S')}.csv"
    filepath = os.path.join("static/export", filename)

    with open(filepath, "w", newline="", encoding="utf-8-sig") as f:
        writer = csv.writer(f)
        writer.writerow(
            ["ID", "学号", "姓名", "文件名", "提交时间", "状态", "评分", "反馈"]
        )

        for sub in submissions:
            writer.writerow(
                [
                    sub.id,
                    sub.user_id,
                    getattr(sub, "username", ""),
                    sub.file_name or "",
                    sub.submitted_at.strftime("%Y-%m-%d %H:%M:%S")
                    if sub.submitted_at
                    else "",
                    sub.status,
                    sub.grade if sub.grade is not None else "",
                    sub.feedback or "",
                ]
            )

    return {"file_url": f"/static/export/{filename}", "filename": filename}



@router.post(
    "/submissions",
    response_model=SubmissionResponse,
    status_code=status.HTTP_201_CREATED,
)
def create_submission(
    submission: SubmissionCreate,
    request: Request,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    assignment = assignment_service.get(db, submission.assignment_id)
    if not assignment:
        raise HTTPException(status_code=404, detail="Assignment not found")

    existing = submission_service.get_user_submission_for_assignment(
        db, current_user.id, submission.assignment_id
    )
    if existing:
        raise HTTPException(
            status_code=400,
            detail="You have already submitted this assignment",
        )

    db_submission = submission_service.create(
        db,
        submission,
        current_user.id,
        submission.file_path,
        submission.file_name,
        submission.file_size,
    )

    if db_submission.file_path:
        db_submission.file_url = build_file_url(request, db_submission.file_path)

    return db_submission


@router.put("/submissions/{submission_id}", response_model=SubmissionResponse)
def update_submission(
    submission_id: str,
    submission: SubmissionUpdate,
    request: Request,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    db_submission = submission_service.get(db, submission_id)
    if not db_submission:
        raise HTTPException(status_code=404, detail="Submission not found")

    db_submission = submission_service.update(db, submission_id, submission)

    if db_submission.file_path:
        db_submission.file_url = build_file_url(request, db_submission.file_path)

    return db_submission


@router.delete("/submissions/{submission_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_submission(
    submission_id: str,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    success = submission_service.delete(db, submission_id)
    if not success:
        raise HTTPException(status_code=404, detail="Submission not found")
