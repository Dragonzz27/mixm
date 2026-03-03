from pathlib import Path

from pydantic_settings import BaseSettings
from functools import lru_cache


class Settings(BaseSettings):
    SECRET_KEY: str
    ALGORITHM: str
    ACCESS_TOKEN_EXPIRE_MINUTES: int
    REFRESH_TOKEN_EXPIRE_DAYS: int
    DATABASE_URL: str
    K8S_EXPORT_PORT: int = 18800
    K8S_CONFIG_PATH: str = "kubeconfig/k3s.yaml"
    K8S_EXPORT_HOST: str = "localhost"  # # MASTER节点IP

    def __init__(self, **kwargs):
        super().__init__(**kwargs)
        if self.K8S_CONFIG_PATH:
            base_dir = Path(__file__).parent
            self.K8S_CONFIG_PATH = str((base_dir / self.K8S_CONFIG_PATH).resolve())

    class Config:
        env_file = ".env"


@lru_cache
def get_settings():
    return Settings()
