#!/usr/bin/env bash

set -euo pipefail

username="${1:?缺少开发用户名}"

if ! id -u "${username}" >/dev/null 2>&1; then
    echo "Error: User ${username} not found" >&2
    exit 1
fi

usermod -aG sudo "${username}"
printf '%s ALL=(ALL) NOPASSWD:ALL\n' "${username}" > "/etc/sudoers.d/${username}"
chmod 0440 "/etc/sudoers.d/${username}"
printf '%s:ubuntu\n' "${username}" | chpasswd

printf '\n. /etc/profile.d/mixm-venv.sh\n' >> /etc/bash.bashrc
mkdir -p /workspace /var/run/sshd
chown "${username}:${username}" /workspace
chown -R "${username}:${username}" /opt/venv

ssh-keygen -A
sed -i \
    -e 's/^#\?PermitRootLogin.*/PermitRootLogin no/' \
    -e 's/^#\?PasswordAuthentication.*/PasswordAuthentication yes/' \
    -e 's/^#\?PubkeyAuthentication.*/PubkeyAuthentication no/' \
    -e 's/^#\?KbdInteractiveAuthentication.*/KbdInteractiveAuthentication yes/' \
    -e 's/^#\?X11Forwarding.*/X11Forwarding yes/' \
    -e 's/^#\?UsePAM.*/UsePAM yes/' \
    /etc/ssh/sshd_config
printf 'AllowUsers %s\n' "${username}" >> /etc/ssh/sshd_config
