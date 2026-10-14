import json
import logging
import secrets
import socket
import threading
import time
from typing import Optional, Tuple

import paramiko
from kubernetes import client, config as k8s_config
from kubernetes.stream import stream

from app.config import get_settings
from app.database import SessionLocal
from app.models.instance import Instance

logger = logging.getLogger(__name__)


class _GatewayServer(paramiko.ServerInterface):
    def __init__(self):
        self.shell_event = threading.Event()
        self.instance_id: Optional[str] = None
        self.namespace: str = "default"

    def _resolve_instance(self, username: str, password: str) -> Optional[Tuple[str, str]]:
        db = SessionLocal()
        try:
            db_instance = db.query(Instance).filter(Instance.ssh_username == username).first()
            if not db_instance or db_instance.status != "running":
                return None

            config_data = {}
            if db_instance.config:
                try:
                    config_data = (
                        json.loads(db_instance.config)
                        if isinstance(db_instance.config, str)
                        else db_instance.config
                    )
                except Exception:
                    config_data = {}

            expected_password = config_data.get("password")
            if not expected_password:
                return None

            if not secrets.compare_digest(expected_password, password):
                return None

            return db_instance.id, (db_instance.namespace or "default")
        finally:
            db.close()

    def check_auth_password(self, username: str, password: str):
        resolved = self._resolve_instance(username, password)
        if not resolved:
            return paramiko.AUTH_FAILED

        self.instance_id, self.namespace = resolved
        return paramiko.AUTH_SUCCESSFUL

    def get_allowed_auths(self, username: str) -> str:
        return "password"

    def check_channel_request(self, kind: str, chanid: int):
        if kind == "session":
            return paramiko.OPEN_SUCCEEDED
        return paramiko.OPEN_FAILED_ADMINISTRATIVELY_PROHIBITED

    def check_channel_pty_request(self, *args, **kwargs):
        return True

    def check_channel_shell_request(self, channel):
        self.shell_event.set()
        return True


class SSHGatewayService:
    def __init__(self):
        self.settings = get_settings()
        self.host = "0.0.0.0"
        self.port = self.settings.K8S_SSH_PORT
        self._socket: Optional[socket.socket] = None
        self._thread: Optional[threading.Thread] = None
        self._running = False
        self._host_key = paramiko.RSAKey.generate(2048)

    def start(self):
        if self._running:
            return

        self._running = True
        self._thread = threading.Thread(target=self._serve, daemon=True)
        self._thread.start()
        logger.info("[SSH Gateway] Started on %s:%s", self.host, self.port)

    def stop(self):
        self._running = False
        if self._socket:
            try:
                self._socket.close()
            except Exception:
                pass
        logger.info("[SSH Gateway] Stopped")

    def _serve(self):
        try:
            k8s_config.load_kube_config(config_file=self.settings.K8S_CONFIG_PATH)
        except Exception as e:
            logger.error("[SSH Gateway] Failed to load kube config: %s", e)
            self._running = False
            return

        self._socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)

        try:
            self._socket.bind((self.host, self.port))
            self._socket.listen(100)
        except Exception as e:
            logger.error("[SSH Gateway] Failed to bind %s:%s: %s", self.host, self.port, e)
            self._running = False
            return

        while self._running:
            try:
                client_sock, addr = self._socket.accept()
            except OSError:
                break
            except Exception as e:
                logger.error("[SSH Gateway] Accept error: %s", e)
                continue

            threading.Thread(
                target=self._handle_client,
                args=(client_sock, addr),
                daemon=True,
            ).start()

    def _find_target_pod(self, instance_id: str, namespace: str) -> Optional[str]:
        core_v1 = client.CoreV1Api()
        label_selector = f"app=deploy-cs-{instance_id}"
        pods = core_v1.list_namespaced_pod(namespace=namespace, label_selector=label_selector)

        for pod in pods.items:
            if pod.status and pod.status.phase == "Running":
                return pod.metadata.name

        return None

    def _bridge_shell(self, channel, ws_client):
        while True:
            if channel.closed:
                break

            if channel.recv_ready():
                data = channel.recv(4096)
                if not data:
                    break
                ws_client.write_stdin(data.decode("utf-8", errors="ignore"))

            if ws_client.peek_stdout():
                out = ws_client.read_stdout()
                if out:
                    channel.send(out.encode("utf-8", errors="ignore"))

            if ws_client.peek_stderr():
                err = ws_client.read_stderr()
                if err:
                    channel.send_stderr(err.encode("utf-8", errors="ignore"))

            if ws_client.is_open() is False:
                break

            time.sleep(0.01)

    def _handle_client(self, client_sock: socket.socket, addr):
        transport = None
        channel = None
        ws_client = None

        try:
            transport = paramiko.Transport(client_sock)
            transport.add_server_key(self._host_key)

            server = _GatewayServer()
            transport.start_server(server=server)

            channel = transport.accept(20)
            if channel is None:
                return

            if not server.shell_event.wait(timeout=20):
                channel.close()
                return

            if not server.instance_id:
                channel.close()
                return

            pod_name = self._find_target_pod(server.instance_id, server.namespace)
            if not pod_name:
                channel.send("Target instance pod is not running.\r\n")
                channel.close()
                return

            core_v1 = client.CoreV1Api()
            ws_client = stream(
                core_v1.connect_get_namespaced_pod_exec,
                pod_name,
                server.namespace,
                command=["/bin/bash", "-il"],
                stderr=True,
                stdin=True,
                stdout=True,
                tty=True,
                _preload_content=False,
            )

            self._bridge_shell(channel, ws_client)
        except Exception as e:
            logger.error("[SSH Gateway] Client %s error: %s", addr, e)
        finally:
            try:
                if ws_client:
                    ws_client.close()
            except Exception:
                pass
            try:
                if channel:
                    channel.close()
            except Exception:
                pass
            try:
                if transport:
                    transport.close()
            except Exception:
                pass
            try:
                client_sock.close()
            except Exception:
                pass
