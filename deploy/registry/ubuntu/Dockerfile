# syntax=docker/dockerfile:1
FROM ubuntu:24.04

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# 使用基础镜像中已有的用户。
ARG USERNAME=ubuntu

ENV DEBIAN_FRONTEND=noninteractive

# APT 索引和下载缓存保存在 BuildKit 挂载中。
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt/lists,sharing=locked \
    apt-get update \
    && apt-get install -y --no-install-recommends \
        build-essential \
        ca-certificates \
        curl \
        git \
        less \
        libssl-dev \
        openssh-server \
        pkg-config \
        python3 \
        python3-pip \
        python3-venv \
        sudo \
        vim

RUN curl -fsSL https://astral.sh/uv/install.sh -o /tmp/install-uv.sh \
    && env UV_INSTALL_DIR=/usr/local/bin sh /tmp/install-uv.sh \
    && rm -f /tmp/install-uv.sh

# 配置用户权限、默认 SSH 密码和 SSH 服务。
RUN set -eux; \
    if ! id -u "${USERNAME}" > /dev/null 2>&1; then \
        echo "Error: User ${USERNAME} not found" >&2; \
        exit 1; \
    fi; \
    usermod -aG sudo "${USERNAME}"; \
    printf '%s ALL=(ALL) NOPASSWD:ALL\n' "${USERNAME}" > "/etc/sudoers.d/${USERNAME}"; \
    chmod 0440 "/etc/sudoers.d/${USERNAME}"; \
    printf '%s:ubuntu\n' "${USERNAME}" | chpasswd; \
    mkdir -p /var/run/sshd; \
    ssh-keygen -A; \
    sed -i \
        -e 's/^#\?PermitRootLogin.*/PermitRootLogin no/' \
        -e 's/^#\?PasswordAuthentication.*/PasswordAuthentication yes/' \
        -e 's/^#\?PubkeyAuthentication.*/PubkeyAuthentication no/' \
        -e 's/^#\?KbdInteractiveAuthentication.*/KbdInteractiveAuthentication yes/' \
        -e 's/^#\?X11Forwarding.*/X11Forwarding yes/' \
        -e 's/^#\?UsePAM.*/UsePAM yes/' \
        /etc/ssh/sshd_config; \
    printf 'AllowUsers %s\n' "${USERNAME}" >> /etc/ssh/sshd_config

WORKDIR /workspace

RUN chown "${USERNAME}:${USERNAME}" /workspace

EXPOSE 22

CMD ["/usr/sbin/sshd", "-D", "-e"]
