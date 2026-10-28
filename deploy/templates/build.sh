#!/usr/bin/env bash

set -euo pipefail

export DOCKER_BUILDKIT="${DOCKER_BUILDKIT:-1}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${SCRIPT_DIR}"

prompt_with_default() {
    local message="$1"
    local default_value="$2"
    local input
    read -r -p "${message} [${default_value}]: " input || return 1
    printf '%s\n' "${input:-${default_value}}"
}

TEMPLATES=(
    mixm
    mixm.codeserver
    mixm.codeserver.torch280
    mixm.codeserver.torch291
)
DESCRIPTIONS=(
    "基础开发环境 + SSH"
    "基础开发环境 + SSH + code-server"
    "code-server + PyTorch 2.8.0"
    "code-server + PyTorch 2.9.1"
)

echo "mixm 实验镜像构建"
echo "构建目录: ${SCRIPT_DIR}"
echo
for index in "${!TEMPLATES[@]}"; do
    printf '  %d) Dockerfile.%s (%s)\n' "$((index + 1))" "${TEMPLATES[index]}" "${DESCRIPTIONS[index]}"
done

while true; do
    choice="$(prompt_with_default "选择模板 1-4" "1")"
    case "${choice}" in
        [1-4])
            TEMPLATE="${TEMPLATES[choice - 1]}"
            break
            ;;
        *) echo "无效选项，请输入 1、2、3 或 4。" >&2 ;;
    esac
done

DOCKERFILE="Dockerfile.${TEMPLATE}"
BUILD_ARGS=()
DEFAULT_IMAGE_TAG=v0.0.1

if [[ "${TEMPLATE}" == *.torch* ]]; then
    while true; do
        DEVICE="$(prompt_with_default "PyTorch 计算设备 cpu/cuda" "${TORCH_DEVICE:-cpu}")"
        case "${DEVICE}" in
            cpu|cuda) break ;;
            *) echo "计算设备只能是 cpu 或 cuda。" >&2 ;;
        esac
    done
    BUILD_ARGS+=(--build-arg "TORCH_DEVICE=${DEVICE}")
    DEFAULT_IMAGE_TAG="cpu-v0.0.1"
    if [[ "${DEVICE}" == cuda ]]; then
        while true; do
            CUDA_TAG="$(prompt_with_default "CUDA 通道" "${TORCH_CUDA_TAG:-cu128}")"
            if [[ "${CUDA_TAG}" =~ ^cu[0-9]+$ ]]; then
                break
            fi
            echo "CUDA 通道格式应为 cu128 等。" >&2
        done
        BUILD_ARGS+=(--build-arg "TORCH_CUDA_TAG=${CUDA_TAG}")
        DEFAULT_IMAGE_TAG="${CUDA_TAG}-v0.0.1"
    fi
fi

IMAGE_NAME="$(prompt_with_default "镜像名称" "${TEMPLATE//./-}")"
IMAGE_TAG="$(prompt_with_default "镜像 Tag" "${DEFAULT_IMAGE_TAG}")"
REGISTRY_DOMAIN="$(prompt_with_default "Registry 地址，不含 http:// 或 https://" "${REGISTRY_PREFIX:-localhost:55000}")"
REGISTRY_DOMAIN="${REGISTRY_DOMAIN%/}"
case "${REGISTRY_DOMAIN}" in
    *://*)
        echo "Registry 地址不能包含协议，请使用 localhost:55000 等地址。" >&2
        exit 1
        ;;
esac

FULL_IMAGE="${IMAGE_NAME}:${IMAGE_TAG}"
PUSH_IMAGE="${REGISTRY_DOMAIN}/${FULL_IMAGE}"
CMD=(docker build -f "${DOCKERFILE}" -t "${FULL_IMAGE}")
CMD+=("${BUILD_ARGS[@]}" .)

echo
printf '将执行:'
printf ' %q' "${CMD[@]}"
printf '\n推送标签: %s\n' "${PUSH_IMAGE}"

read -r -p "确认执行构建？[y/N]: " confirm
if [[ ! "${confirm}" =~ ^[Yy]$ ]]; then
    echo "已取消构建。"
    exit 0
fi

"${CMD[@]}"
docker tag "${FULL_IMAGE}" "${PUSH_IMAGE}"

echo "构建完成: ${FULL_IMAGE}"
printf '推送命令: docker push %q\n' "${PUSH_IMAGE}"
