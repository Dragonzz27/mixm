from pathlib import Path

from pydantic_settings import BaseSettings
from functools import lru_cache


class Settings(BaseSettings):
    SECRET_KEY: str
    ALGORITHM: str
    ACCESS_TOKEN_EXPIRE_MINUTES: int
    REFRESH_TOKEN_EXPIRE_DAYS: int
    DATABASE_URL: str
    K8S_CONFIG_PATH: str = "kubeconfig/k3s.yaml"
    K8S_SSH_GATEWAY_HOST: str = "localhost"
    K8S_CODE_SERVER_HOST: str = "localhost"
    K8S_SSH_PORT: int = 30022
    REGISTRY_URL: str = "http://172.20.137.239:55000"
    REGISTRY_PREFIX: str = "172.20.137.239:5000"
    REGISTRY_USERNAME: str = ""
    REGISTRY_PASSWORD: str = ""
    STORAGE_CLASS_NAME: str = "local-path"
    REGISTRY_DELETE_TIMEOUT_SECONDS: int = 30
    USER_PERSONAL_PVC_CAPACITY: str = "10Gi"
    PVC_ACCESS_MODE: str = "ReadWriteMany"
    RUSTFS_ENDPOINT: str = "http://localhost:19000"
    RUSTFS_BUCKET: str = "content"
    RUSTFS_ACCESS_KEY: str = "rustfsadmin"
    RUSTFS_SECRET_KEY: str = "rustfsadmin"
    CELERY_BROKER_URL: str = "redis://localhost:6379/0"
    CELERY_RESULT_BACKEND: str = "redis://localhost:6379/1"
    QUEUE_REDIS_URL: str = "redis://localhost:6379/2"
    LLM_API_URL: str = "https://api.openai.com/v1"
    LLM_API_KEY: str = ""
    LLM_MODEL: str = "gpt-4o"
    KT_MODEL_BACKEND: str = "kt_service"
    KT_AKT_CHECKPOINT: str = ""
    KT_AKT_DEVICE: str = "cpu"
    KT_SERVICE_URL: str = "http://kt-service:8010"
    KT_SERVICE_TIMEOUT_SECONDS: float = 3.0


    class Config:
        env_file = ".env"


@lru_cache
def get_settings():
    return Settings()
