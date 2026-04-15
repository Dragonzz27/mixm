import os
import uuid
from datetime import datetime
from typing import Optional

from fastapi import (
    APIRouter,
    Depends,
    HTTPException,
    status,
    UploadFile,
    Request,
)
from fastapi.responses import Response, FileResponse
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


class AssignmentListResponse(BaseModel):
    total: int
    assignments: list[AssignmentResponse]


class SubmissionListResponse(BaseModel):
    total: int
    submissions: list[SubmissionResponse]


def ensure_upload_dir():
    os.makedirs(UPLOAD_DIR, exist_ok=True)


@router.post(
    "/", response_model=AssignmentResponse, status_code=status.HTTP_201_CREATED
)
def create_assignment(
    assignment: AssignmentCreate,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    db_assignment = assignment_service.create(db, assignment, current_user.id)
    return db_assignment


@router.get("/", response_model=AssignmentListResponse)
def get_assignments(
    skip: int = 0,
    limit: int = 20,
    status: Optional[str] = None,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    assignments, total = assignment_service.get_list(db, skip, limit, status)
    return {"total": total, "assignments": assignments}


@router.get("/active", response_model=list[AssignmentResponse])
def get_active_assignments(
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    assignments = assignment_service.get_all_active(db)
    return assignments


@router.get("/my/submissions", response_model=list[SubmissionResponse])
def get_my_submissions(
    assignment_id: Optional[int] = None,
    request: Request = None,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    submissions = submission_service.get_by_user(db, current_user.id, assignment_id)

    if request:
        base_url = str(request.base_url).rstrip("/")
        for sub in submissions:
            if sub.file_path:
                storage_filename = os.path.basename(sub.file_path)
                sub.file_url = f"{base_url}/api/assignments/file/{storage_filename}"

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

    media_type = "application/octet-stream"
    if filename.endswith(".doc"):
        media_type = "application/msword"
    elif filename.endswith(".docx"):
        media_type = (
            "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
        )
    elif filename.endswith(".pdf"):
        media_type = "application/pdf"
    elif filename.endswith(".ppt"):
        media_type = "application/vnd.ms-powerpoint"
    elif filename.endswith(".pptx"):
        media_type = (
            "application/vnd.openxmlformats-officedocument.presentationml.presentation"
        )
    elif filename.endswith(".txt"):
        media_type = "text/plain"

    return FileResponse(
        file_path,
        media_type=media_type,
        filename=filename.split("._.")[-1] if "._." in filename else filename,
    )


@router.get("/{assignment_id}", response_model=AssignmentResponse)
def get_assignment(
    assignment_id: int,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    assignment = assignment_service.get(db, assignment_id)
    if not assignment:
        raise HTTPException(status_code=404, detail="Assignment not found")
    return assignment


@router.put("/{assignment_id}", response_model=AssignmentResponse)
def update_assignment(
    assignment_id: int,
    assignment: AssignmentUpdate,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    db_assignment = assignment_service.update(db, assignment_id, assignment)
    if not db_assignment:
        raise HTTPException(status_code=404, detail="Assignment not found")
    return db_assignment


@router.delete("/{assignment_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_assignment(
    assignment_id: int,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    success = assignment_service.delete(db, assignment_id)
    if not success:
        raise HTTPException(status_code=404, detail="Assignment not found")


@router.get("/{assignment_id}/submissions", response_model=SubmissionListResponse)
def get_submissions(
    assignment_id: int,
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

    base_url = str(request.base_url).rstrip("/")
    for sub in submissions:
        if sub.file_path:
            storage_filename = os.path.basename(sub.file_path)
            sub.file_url = f"{base_url}/api/assignments/file/{storage_filename}"

    return {"total": total, "submissions": submissions}


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

    base_url = str(request.base_url).rstrip("/")
    if db_submission.file_path:
        storage_filename = os.path.basename(db_submission.file_path)
        db_submission.file_url = f"{base_url}/api/assignments/file/{storage_filename}"

    return db_submission


@router.put("/submissions/{submission_id}", response_model=SubmissionResponse)
def update_submission(
    submission_id: int,
    submission: SubmissionUpdate,
    request: Request,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    db_submission = submission_service.get(db, submission_id)
    if not db_submission:
        raise HTTPException(status_code=404, detail="Submission not found")

    if db_submission.user_id != current_user.id:
        raise HTTPException(status_code=403, detail="Forbidden")

    db_submission = submission_service.update(db, submission_id, submission)

    base_url = str(request.base_url).rstrip("/")
    if db_submission.file_path:
        storage_filename = os.path.basename(db_submission.file_path)
        db_submission.file_url = f"{base_url}/api/assignments/file/{storage_filename}"

    return db_submission


@router.delete("/submissions/{submission_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_submission(
    submission_id: int,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user),
):
    success = submission_service.delete(db, submission_id)
    if not success:
        raise HTTPException(status_code=404, detail="Submission not found")
