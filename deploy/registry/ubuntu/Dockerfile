# syntax=docker/dockerfile:1.7
FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive

# 直接使用镜像自带的 ubuntu 用户 (UID/GID 通常为 1000)
ARG USERNAME=ubuntu
ARG USER_UID=1000
ARG USER_GID=1000

RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
	--mount=type=cache,target=/var/lib/apt/lists,sharing=locked \
	apt-get update && apt-get install -y --no-install-recommends \
	ca-certificates \
	curl \
	git \
	less \
	vim \
	sudo \
	openssh-server \
	python3 \
	python3-pip \
	python3-venv \
	build-essential \
	pkg-config \
	libssl-dev

RUN curl -LsSf https://astral.sh/uv/install.sh | env UV_INSTALL_DIR=/usr/local/bin sh

RUN set -eux; \
    if ! id -u "${USERNAME}" > /dev/null 2>&1; then \
        echo "Error: User ${USERNAME} not found"; exit 1; \
    fi; \
    usermod -aG sudo "${USERNAME}"; \
    echo "${USERNAME} ALL=(ALL) NOPASSWD:ALL" > "/etc/sudoers.d/${USERNAME}"; \
    chmod 0440 "/etc/sudoers.d/${USERNAME}"; \
    # 设置密码为 "ubuntu" (用户名和密码相同，方便测试)
    echo "${USERNAME}:ubuntu" | chpasswd

# 准备 sshd 运行目录与主机密钥
RUN mkdir -p /var/run/sshd \
	&& ssh-keygen -A

# 【修改点 2】修改 SSH 配置：允许密码、允许 X11、其余保持安全限制
RUN sed -i 's/^#\?PermitRootLogin.*/PermitRootLogin no/' /etc/ssh/sshd_config \
    && sed -i 's/^#\?PasswordAuthentication.*/PasswordAuthentication yes/' /etc/ssh/sshd_config \
	&& sed -i 's/^#\?PubkeyAuthentication.*/PubkeyAuthentication no/' /etc/ssh/sshd_config \
    && sed -i 's/^#\?KbdInteractiveAuthentication.*/KbdInteractiveAuthentication yes/' /etc/ssh/sshd_config \
    && sed -i 's/^#\?X11Forwarding.*/X11Forwarding yes/' /etc/ssh/sshd_config \
    && sed -i 's/^#\?UsePAM.*/UsePAM yes/' /etc/ssh/sshd_config \
    && echo "AllowUsers ${USERNAME}" >> /etc/ssh/sshd_config

WORKDIR /workspace

# 确保工作目录权限正确
RUN chown ${USERNAME}:${USERNAME} /workspace

EXPOSE 22

CMD ["/usr/sbin/sshd", "-D", "-e"]