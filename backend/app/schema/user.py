from pydantic import BaseModel, EmailStr, field_validator, model_validator
from typing import Optional, List


class UserBase(BaseModel):
    username: str
    email: Optional[EmailStr] = None
    phone: Optional[str] = None

    @field_validator("email", "phone", mode="before")
    @classmethod
    def empty_to_none(cls, v):
        if v == "":
            return None
        return v

    @model_validator(mode="after")
    def validate_email_or_phone(self):
        if not self.email and not self.phone:
            raise ValueError("Either email or phone must be provided")
        return self


class UserCreate(UserBase):
    password: str


class UserUpdate(BaseModel):
    username: Optional[str] = None
    email: Optional[EmailStr] = None
    phone: Optional[str] = None
    is_active: Optional[bool] = None

    @field_validator("email", "phone", mode="before")
    @classmethod
    def empty_to_none(cls, v):
        if v == "":
            return None
        return v


class UserLogin(BaseModel):
    account: str
    password: str


class PasswordChange(BaseModel):
    new_password: str


class UserResponse(UserBase):
    id: int
    is_active: bool
    is_superuser: bool

    class Config:
        from_attributes = True


class UserListResponse(BaseModel):
    total: int
    users: List[UserResponse]


class Token(BaseModel):
    access_token: str
    refresh_token: str
    token_type: str


class TokenData(BaseModel):
    username: Optional[str] = None


class TokenRefresh(BaseModel):
    refresh_token: str


class ImportResult(BaseModel):
    total: int
    success: int
    failed: int
    errors: List[str]
