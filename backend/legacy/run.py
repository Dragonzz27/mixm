import psutil
import uvicorn


def _find_celery_workers() -> list[dict]:
    workers = []
    for proc in psutil.process_iter(attrs=["pid", "cmdline"]):
        try:
            cmdline_list = proc.info.get("cmdline") or []
            cmdline = " ".join(cmdline_list)
            if (
                "celery" in cmdline
                and "app.celery_app:celery_app" in cmdline
                and "worker" in cmdline
            ):
                workers.append({"pid": proc.info["pid"], "cmdline": cmdline})
        except (psutil.NoSuchProcess, psutil.AccessDenied):
            continue
    return workers


def _print_celery_status() -> None:
    workers = _find_celery_workers()
    if not workers:
        print("[Startup] Celery worker status: NOT RUNNING")
        print("[Startup] Start manually: python run_celery.py")
        return

    print(f"[Startup] Celery worker status: RUNNING ({len(workers)})")
    for worker in workers:
        print(f"[Startup] - PID {worker['pid']}: {worker['cmdline']}")

if __name__ == "__main__":
    # _print_celery_status()
    uvicorn.run("app.main:app", host="0.0.0.0", port=18002, reload=True)
