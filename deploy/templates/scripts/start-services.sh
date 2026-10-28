#!/usr/bin/env bash

set -euo pipefail

# PVC 挂载会覆盖构建时的目录属主，在启动时授予开发用户工作目录写权限。
mkdir -p /workspace /var/run/sshd
chown "${USERNAME}:${USERNAME}" /workspace
/usr/sbin/sshd -D -e &

export PASSWORD="${CODE_SERVER_PASSWORD}"
exec su -s /bin/bash -c '
    . /etc/profile.d/mixm-venv.sh
    exec code-server \
        --bind-addr "0.0.0.0:${CODE_SERVER_PORT}" \
        --auth password \
        --disable-update-check \
        /workspace
' "${USERNAME}"
