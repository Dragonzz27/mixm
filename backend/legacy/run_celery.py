import os
import subprocess
import sys

def _build_celery_args() -> list[str]:
    concurrency = os.getenv("CELERY_WORKER_CONCURRENCY", "4")
    pool = os.getenv("CELERY_WORKER_POOL", "solo")
    loglevel = os.getenv("CELERY_WORKER_LOGLEVEL", "info")
    queue_name = os.getenv("CELERY_WORKER_QUEUE", "instance_create")

    return [
        "celery",
        "-A",
        "app.celery_app:celery_app",
        "worker",
        "-Q",
        queue_name,
        "--concurrency",
        str(concurrency),
        "--loglevel",
        loglevel,
        "-P",
        pool,
    ]


if __name__ == "__main__":
    args = _build_celery_args()
    print("[Celery] Starting worker:", " ".join(args))
    cmd = [sys.executable, "-m"] + args
    sys.exit(subprocess.call(cmd, cwd=os.path.dirname(__file__)))
