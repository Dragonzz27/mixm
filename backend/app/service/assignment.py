import os
import uuid
from datetime import datetime
from typing import List, Optional, Tuple
from sqlalchemy.orm import Session, joinedload

from app.models.assignment import Assignment, Submission
from app.schema.assignment import (
    AssignmentCreate,
    AssignmentUpdate,
    SubmissionCreate,
    SubmissionUpdate,
)


UPLOAD_DIR = "uploads/submissions"


class AssignmentService:
    def create(
        self, db: Session, assignment: AssignmentCreate, created_by: str
    ) -> Assignment:
        db_assignment = Assignment(
            title=assignment.title,
            description=assignment.description,
            content_id=assignment.content_id,
            deadline=assignment.deadline,
            created_by=created_by,
        )
        db.add(db_assignment)
        db.commit()
        db.refresh(db_assignment)
        return db_assignment

    def get(self, db: Session, assignment_id: int) -> Optional[Assignment]:
        return db.query(Assignment).filter(Assignment.id == assignment_id).first()

    def get_list(
        self,
        db: Session,
        skip: int = 0,
        limit: int = 20,
        status: Optional[str] = None,
    ) -> Tuple[List[Assignment], int]:
        query = db.query(Assignment)
        if status:
            query = query.filter(Assignment.status == status)
        query = query.order_by(Assignment.created_at.desc())
        total = query.count()
        assignments = query.offset(skip).limit(limit).all()

        for assignment in assignments:
            count = (
                db.query(Submission)
                .filter(Submission.assignment_id == assignment.id)
                .count()
            )
            assignment.submission_count = count

        return assignments, total

    def get_all_active(self, db: Session) -> List[Assignment]:
        return (
            db.query(Assignment)
            .filter(Assignment.status == "active")
            .filter(Assignment.deadline >= datetime.now())
            .order_by(Assignment.deadline.asc())
            .all()
        )

    def update(
        self, db: Session, assignment_id: int, assignment: AssignmentUpdate
    ) -> Optional[Assignment]:
        db_assignment = self.get(db, assignment_id)
        if not db_assignment:
            return None

        update_data = assignment.model_dump(exclude_unset=True)
        for field, value in update_data.items():
            setattr(db_assignment, field, value)

        db.commit()
        db.refresh(db_assignment)
        return db_assignment

    def delete(self, db: Session, assignment_id: int) -> bool:
        db_assignment = self.get(db, assignment_id)
        if not db_assignment:
            return False
        db.delete(db_assignment)
        db.commit()
        return True


class SubmissionService:
    def create(
        self,
        db: Session,
        submission: SubmissionCreate,
        user_id: str,
        file_path: Optional[str] = None,
        file_name: Optional[str] = None,
        file_size: Optional[int] = None,
    ) -> Submission:
        final_file_path = file_path or submission.file_path
        final_file_name = file_name or submission.file_name

        if final_file_path and not final_file_name:
            storage_filename = os.path.basename(final_file_path)
            if "._." in storage_filename:
                parts = storage_filename.split("._.", 1)
                if len(parts) == 2:
                    final_file_name = parts[1]

        db_submission = Submission(
            assignment_id=submission.assignment_id,
            user_id=user_id,
            file_path=final_file_path,
            file_name=final_file_name,
            file_size=file_size or submission.file_size,
            content=submission.content,
        )
        db.add(db_submission)
        db.commit()
        db.refresh(db_submission)
        return db_submission

    def get(self, db: Session, submission_id: int) -> Optional[Submission]:
        return db.query(Submission).filter(Submission.id == submission_id).first()

    def get_by_assignment(
        self,
        db: Session,
        assignment_id: int,
        skip: int = 0,
        limit: int = 20,
    ) -> Tuple[List[Submission], int]:
        query = db.query(Submission).filter(Submission.assignment_id == assignment_id)
        total = query.count()
        submissions = (
            query.options(joinedload(Submission.user))
            .order_by(Submission.submitted_at.desc())
            .offset(skip)
            .limit(limit)
            .all()
        )

        for submission in submissions:
            if submission.user:
                submission.username = submission.user.username

        return submissions, total

    def get_by_user(
        self, db: Session, user_id: str, assignment_id: Optional[int] = None
    ) -> List[Submission]:
        query = db.query(Submission).filter(Submission.user_id == user_id)
        if assignment_id:
            query = query.filter(Submission.assignment_id == assignment_id)
        return query.order_by(Submission.submitted_at.desc()).all()

    def get_user_submission_for_assignment(
        self, db: Session, user_id: str, assignment_id: int
    ) -> Optional[Submission]:
        return (
            db.query(Submission)
            .filter(
                Submission.user_id == user_id,
                Submission.assignment_id == assignment_id,
            )
            .first()
        )

    def update(
        self,
        db: Session,
        submission_id: int,
        submission: SubmissionUpdate,
    ) -> Optional[Submission]:
        db_submission = self.get(db, submission_id)
        if not db_submission:
            return None

        update_data = submission.model_dump(exclude_unset=True)
        if "grade" in update_data or "feedback" in update_data:
            update_data["reviewed_at"] = datetime.now()
        for field, value in update_data.items():
            setattr(db_submission, field, value)

        db.commit()
        db.refresh(db_submission)
        return db_submission

    def delete(self, db: Session, submission_id: int) -> bool:
        db_submission = self.get(db, submission_id)
        if not db_submission:
            return False
        if db_submission.file_path and os.path.exists(db_submission.file_path):
            try:
                os.remove(db_submission.file_path)
            except Exception:
                pass
        db.delete(db_submission)
        db.commit()
        return True


assignment_service = AssignmentService()
submission_service = SubmissionService()
