#!/usr/bin/env bash

set -euo pipefail

usage() {
  echo "用法: sudo bash scripts/setup-longhorn-mount-propagation.sh"
  echo "启动 mixm 前，在 K3s 节点上手动设置 Longhorn 所需的共享挂载传播。"
  echo "设置仅对当前运行的系统生效，重启系统或 WSL 后需重新执行。"
}

if [[ "${1:-}" == "--help" || "${1:-}" == "-h" ]]; then
  usage
  exit 0
fi

if (( $# > 0 )); then
  usage >&2
  exit 1
fi

if (( EUID != 0 )); then
  echo "请使用 sudo 运行此脚本。" >&2
  usage >&2
  exit 1
fi

for required_command in mount findmnt; do
  if ! command -v "${required_command}" > /dev/null 2>&1; then
    echo "缺少必要命令: ${required_command}" >&2
    exit 1
  fi
done

mount --make-rshared /
findmnt -T / -o TARGET,SOURCE,FSTYPE,PROPAGATION

echo "共享挂载传播已应用，可以继续启动 mixm。"
echo "重启系统或 WSL 后，请在启动 mixm 前重新执行本脚本。"
echo "查看 Longhorn 状态: kubectl -n longhorn-system get pods"
