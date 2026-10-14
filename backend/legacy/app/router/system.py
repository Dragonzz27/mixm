from fastapi import APIRouter
import psutil
import os
import platform
from datetime import datetime

router = APIRouter()


@router.get("/stats")
def get_system_stats():
    cpu_percent = psutil.cpu_percent(interval=1)
    memory = psutil.virtual_memory()
    disk = psutil.disk_usage('/')
    
    return {
        "cpu": cpu_percent,
        "memory": {
            "total": memory.total,
            "used": memory.used,
            "percent": memory.percent
        },
        "disk": {
            "total": disk.total,
            "used": disk.used,
            "percent": disk.percent
        },
        "processes": len(psutil.pids()),
        "timestamp": datetime.now().isoformat()
    }


@router.get("/load")
def get_load_average():
    if hasattr(os, 'getloadavg') and platform.system() != 'Windows':
        load1, load5, load15 = os.getloadavg()
        return {
            "load_1": load1,
            "load_5": load5,
            "load_15": load15
        }
    return {
        "load_1": None,
        "load_5": None,
        "load_15": None,
        "message": "Load average not available on Windows"
    }
