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
    K8S_SSH_GATEWAY_HOST: str = "localhost"                # # SSH是FastAPI Gateway启动的，所以是服务的主机地址
    K8S_CODE_SERVER_HOST: str = "localhost"                # code-server访问地址，通常是集群内网地址或负载均衡地址
    K8S_SSH_PORT: int = 30022
    REGISTRY_URL: str = "http://172.20.137.239:55000"      # 镜像仓库的 URL，通常是 registry 的地址
    REGISTRY_PREFIX: str = "172.20.137.239:5000"
    REGISTRY_USERNAME: str = ""
    REGISTRY_PASSWORD: str = ""
    STORAGE_CLASS_NAME: str = "local-path" # 存储类名称，默认为 local-path，适用于大多数环境；如果使用 Longhorn，可以改为 longhorn
    REGISTRY_DELETE_TIMEOUT_SECONDS: int = 30
    USER_PERSONAL_PVC_CAPACITY: str = "10Gi" # 每个用户个人PVC的固定容量，单位Gi，避免每个实例都创建一个独占PVC导致资源浪费
    PVC_ACCESS_MODE: str = "ReadWriteMany" # PVC访问模式，默认为ReadWriteOnce，适用于大多数场景；如果需要多个实例共享同一个PVC，可以改为ReadWriteMany，但需要底层存储支持
    
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
