from pydantic import BaseModel, EmailStr, field_validator, model_validator
from typing import Optional, List, Dict, Any


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
    role_id: Optional[str] = None


class UserUpdate(BaseModel):
    username: Optional[str] = None
    email: Optional[EmailStr] = None
    phone: Optional[str] = None
    is_active: Optional[bool] = None
    role_id: Optional[str] = None

    @field_validator("email", "phone", mode="before")
    @classmethod
    def empty_to_none(cls, v):
        if v == "":
            return None
        return v


class UserLogin(BaseModel):
    email: EmailStr
    password: str


class PasswordChange(BaseModel):
    new_password: str


class UserRoleItem(BaseModel):
    id: str
    name: str


class UserResponse(UserBase):
    id: str
    is_active: bool
    is_superuser: bool
    role: Optional[UserRoleItem] = None
    role_id: Optional[str] = None

    class Config:
        from_attributes = True


class UserListResponse(BaseModel):
    total: int
    users: List[UserResponse]


class Token(BaseModel):
    access_token: str
    refresh_token: str
    token_type: str


class LoginUserInfo(BaseModel):
    id: str
    username: str
    email: Optional[EmailStr] = None
    phone: Optional[str] = None
    is_active: bool
    is_superuser: bool
    role: str
    user_role: Optional[UserRoleItem] = None
    role_id: Optional[str] = None


class MenuItem(BaseModel):
    key: str
    label: str
    path: Optional[str] = None
    icon: Optional[str] = None
    type: Optional[str] = None


class LoginResponse(Token):
    user: LoginUserInfo
    routes: List[Dict[str, Any]]
    menu: List[MenuItem]


class TokenData(BaseModel):
    username: Optional[str] = None


class TokenRefresh(BaseModel):
    refresh_token: str


class ImportResult(BaseModel):
    total: int
    success: int
    failed: int
    errors: List[str]
