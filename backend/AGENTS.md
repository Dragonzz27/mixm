# AGENTS.md

This document provides guidelines for AI agents working in the MutiM codebase.

## Project Overview

MutiM is a full-stack application with a FastAPI Python backend and frontend directory. The backend follows a clean architecture pattern with separate modules for models, routers, schemas, and services.

## Build, Run, and Test Commands

### Backend Commands

```bash
# Install dependencies
pip install -r requirements.txt

# Run the application (development)
python run.py

# Run with uvicorn directly
uvicorn app.main:app --host 0.0.0.0 --port 8002 --reload

# Run tests (if pytest is configured)
pytest -v
pytest app/tests/ -v                    # Run all tests
pytest app/tests/test_file.py -v        # Run specific test file
pytest app/tests/test_file.py::test_func -v  # Run specific test function
pytest -k "test_name" -v                # Run tests matching pattern
```

### Docker Commands

```bash
# Start all services
docker-compose up -d

# Start specific service
docker-compose up -d postgres redis

# View logs
docker-compose logs -f

# Stop all services
docker-compose down

# Stop with volumes
docker-compose down -v
```

## Code Style Guidelines

### Imports

- Use absolute imports with `app.` prefix for local modules
- Group imports in this order: standard library, third-party, local application
- Sort imports alphabetically within each group
- Do NOT add comments unless required by linter

```python
from datetime import datetime, timedelta
from typing import Optional, List

from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session

from app.database import get_db
from app.models.user import User
from app.schema.user import UserCreate, UserResponse
from app.service.auth import AuthService
```

### Formatting

- Use 4 spaces for indentation (no tabs)
- Maximum line length: 120 characters
- Use trailing commas consistently in multi-line constructs
- Add blank lines between function definitions and between major code blocks
- Use spaces around operators: `x = 1 + 2`, not `x=1+2`
- No spaces inside parentheses: `def func(a, b):`, not `def func( a , b ):`

### Type Hints

- Use type hints for all function parameters and return values
- Use `Optional[T]` instead of `Union[T, None]` or `T | None`
- Use `List[T]` instead of `list[T]` for Python 3.9 compatibility
- Include return type annotations on all functions

```python
def get_user(user_id: int, db: Session = Depends(get_db)) -> UserResponse:
    ...

def create_access_token(self, data: dict, expires_delta: Optional[timedelta] = None) -> str:
    ...
```

### Naming Conventions

- **Files**: lowercase with underscores: `user_service.py`, `auth_handler.py`
- **Classes**: PascalCase: `User`, `AuthService`, `ContentResponse`
- **Functions/methods**: snake_case: `get_user()`, `create_access_token()`
- **Variables**: snake_case: `hashed_password`, `access_token`
- **Constants**: UPPER_SNAKE_CASE: `ACCESS_TOKEN_EXPIRE_MINUTES`
- **Database columns**: snake_case with comments in Chinese

```python
class User(Base):
    __tablename__ = "users"

    id = Column(Integer, primary_key=True, index=True, comment="用户ID")
    username = Column(String(50), unique=True, index=True, nullable=False, comment="用户名")
```

### Pydantic Models

- Use BaseModel for all schema definitions
- Use `field_validator` for field-level validation (mode='before' for empty-to-none)
- Use `model_validator` for cross-field validation (mode='after')
- Use `Optional[T] = None` for optional fields with defaults
- Include `class Config: from_attributes = True` for ORM models

```python
class UserCreate(UserBase):
    password: str

    @field_validator('email', mode='before')
    @classmethod
    def empty_to_none(cls, v):
        if v == '':
            return None
        return v

    @model_validator(mode='after')
    def validate_email_or_phone(self):
        if not self.email and not self.phone:
            raise ValueError('Either email or phone must be provided')
        return self
```

### Error Handling

- Use HTTPException for API-level errors with appropriate status codes
- Return meaningful error messages in Chinese or English consistently
- Use proper HTTP status codes: 400 for bad request, 401 for unauthorized, 404 for not found

```python
@router.get("/users/{user_id}", response_model=UserResponse)
def get_user(user_id: int, db: Session = Depends(get_db)):
    db_user = db.query(User).filter(User.id == user_id).first()
    if not db_user:
        raise HTTPException(status_code=404, detail="User not found")
    return db_user
```

### Database Models

- Use SQLAlchemy with declarative base
- Use `func.now()` for server-side timestamp defaults
- Add Chinese comments for all columns
- Use `index=True` for frequently queried columns
- Set `nullable=False` for required fields
- Use `unique=True` for columns requiring uniqueness

```python
class User(Base):
    __tablename__ = "users"

    id = Column(Integer, primary_key=True, index=True, comment="用户ID")
    created_at = Column(DateTime(timezone=True), server_default=func.now(), comment="创建时间")
```

### Router Patterns

- Use APIRouter with prefix and tags
- Include response_model on all endpoints
- Use Depends for dependency injection (database sessions, auth)
- Define service instances at module level when stateless

```python
router = APIRouter()
auth_service = AuthService()

@router.post("/register", response_model=UserResponse)
def register(user: UserCreate, db: Session = Depends(get_db)):
    ...
```

### Service Layer

- Keep business logic in service classes
- Use type annotations for all methods
- Return types should be clear and specific
- Handle exceptions at the service level when appropriate

```python
class AuthService:
    def verify_password(self, plain_password: str, hashed_password: str) -> bool:
        return bcrypt.checkpw(plain_password.encode('utf-8'), hashed_password.encode('utf-8'))
```

## Project Structure

```
/mnt/d/Project/MutiM/
├── backend/
│   ├── app/
│   │   ├── __init__.py
│   │   ├── main.py           # FastAPI app definition, CORS, router inclusion
│   │   ├── config.py         # Settings using pydantic-settings
│   │   ├── database.py       # SQLAlchemy engine, session, get_db dependency
│   │   ├── models/           # SQLAlchemy models (User, Content, Category, etc.)
│   │   ├── router/           # API route handlers
│   │   ├── schema/           # Pydantic schemas for request/response
│   │   └── service/          # Business logic services
│   ├── requirements.txt
│   ├── run.py
│   └── .env
├── frontend/                  # Frontend application
├── docker-compose.yaml        # PostgreSQL and Redis services
└── MutiM.svg
```

## Configuration

- Environment variables loaded via `.env` file
- Use `pydantic-settings` BaseSettings class
- Cache settings with `@lru_cache` for performance

```python
class Settings(BaseSettings):
    SECRET_KEY: str
    ALGORITHM: str
    ACCESS_TOKEN_EXPIRE_MINUTES: int
    DATABASE_URL: str

    class Config:
        env_file = ".env"

@lru_cache
def get_settings():
    return Settings()
```

## Database Setup

- SQLite used for development (app.db)
- Models auto-create tables on startup via `Base.metadata.create_all()`
- Use migrations (alembic) for production schema changes

## Key Dependencies

- FastAPI: Web framework
- SQLAlchemy: ORM
- Pydantic: Data validation
- Pydantic-settings: Configuration management
- Python-jose: JWT handling
- Bcrypt: Password hashing
