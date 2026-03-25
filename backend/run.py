import atexit
import os
import subprocess
import sys

import psutil
import uvicorn


celery_process = None


def _is_celery_worker_running() -> bool:
    for proc in psutil.process_iter(attrs=["name", "cmdline"]):
        try:
            cmdline = " ".join(proc.info.get("cmdline") or [])
            if "celery" in cmdline and "app.celery_app:celery_app" in cmdline:
                return True
        except (psutil.NoSuchProcess, psutil.AccessDenied):
            continue
    return False


def _start_celery_worker() -> None:
    global celery_process
    if _is_celery_worker_running():
        return

    concurrency = os.getenv("CELERY_WORKER_CONCURRENCY", "4")
    pool = os.getenv("CELERY_WORKER_POOL", "solo")
    loglevel = os.getenv("CELERY_WORKER_LOGLEVEL", "info")

    cmd = [
        sys.executable,
        "-m",
        "celery",
        "-A",
        "app.celery_app:celery_app",
        "worker",
        "-Q",
        "instance_create",
        "--concurrency",
        str(concurrency),
        "--loglevel",
        loglevel,
        "-P",
        pool,
    ]

    celery_process = subprocess.Popen(cmd, cwd=os.path.dirname(__file__))


def _stop_celery_worker() -> None:
    global celery_process
    if celery_process and celery_process.poll() is None:
        celery_process.terminate()
    celery_process = None


if __name__ == "__main__":
    auto_start_worker = os.getenv("AUTO_START_CELERY_WORKER", "true").lower() == "true"
    if auto_start_worker:
        _start_celery_worker()
        atexit.register(_stop_celery_worker)

    uvicorn.run("app.main:app", host="0.0.0.0", port=8002, reload=True)
