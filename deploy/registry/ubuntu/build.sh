#!/usr/bin/env bash

set -euo pipefail

export DOCKER_BUILDKIT="${DOCKER_BUILDKIT:-1}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${SCRIPT_DIR}"

prompt_with_default() {
	local message="$1"
	local default_value="$2"
	local input
	read -r -p "${message} [${default_value}]: " input
	if [ -z "${input}" ]; then
		echo "${default_value}"
	else
		echo "${input}"
	fi
}

echo "========================================"
echo " Ubuntu 镜像交互式构建"
echo " 当前目录: ${SCRIPT_DIR}"
echo "========================================"

echo
echo "请选择要构建的镜像模板："
echo "  1) Dockerfile (基础开发镜像)"
echo "  2) Dockerfile.torch291 (含 torch 2.9.1)"
echo "  3) Dockerfile.torch291.codeserver (含 torch 2.9.1 + code-server)"
echo "  4) Dockerfile.torch280.codeserver (含 torch 2.8.0 + code-server)"
echo "  5) Dockerfile.mixm.codeserver (mixm官方 code-server 镜像)"

while true; do
	read -r -p "输入选项 [1-5] (默认 1): " dockerfile_choice
	dockerfile_choice="${dockerfile_choice:-1}"
	case "${dockerfile_choice}" in
		1)
			DOCKERFILE="Dockerfile"
			break
			;;
		2)
			DOCKERFILE="Dockerfile.torch291"
			break
			;;
		3)
			DOCKERFILE="Dockerfile.torch291.codeserver"
			break
			;;
		4)
			DOCKERFILE="Dockerfile.torch280.codeserver"
			break
			;;
		5)
			DOCKERFILE="Dockerfile.mixm.codeserver"
			break
			;;
		*)
			echo "无效选项，请输入 1、2、3、4 或 5。"
			;;
	esac
done

case "${DOCKERFILE}" in
	"Dockerfile")
		DEFAULT_IMAGE_NAME="ubuntu"
		DEFAULT_IMAGE_TAG="24.04.dev"
		;;
	"Dockerfile.torch291")
		DEFAULT_IMAGE_NAME="torch291"
		DEFAULT_IMAGE_TAG="cuda-v0.0.1"
		;;
	"Dockerfile.torch291.codeserver")
		DEFAULT_IMAGE_NAME="torch291-codeserver"
		DEFAULT_IMAGE_TAG="cuda-v0.0.1"
		;;
	"Dockerfile.torch280.codeserver")
		DEFAULT_IMAGE_NAME="torch280-codeserver"
		DEFAULT_IMAGE_TAG="v0.0.1"
		;;
	"Dockerfile.mixm.codeserver")
		DEFAULT_IMAGE_NAME="mixm-codeserver"
		DEFAULT_IMAGE_TAG="v0.0.1"
		;;
	*)
		DEFAULT_IMAGE_NAME="ubuntu"
		DEFAULT_IMAGE_TAG="24.04.dev"
		;;
esac

echo
IMAGE_NAME="$(prompt_with_default "请输入镜像名称" "${DEFAULT_IMAGE_NAME}")"
IMAGE_TAG="$(prompt_with_default "请输入镜像 Tag" "${DEFAULT_IMAGE_TAG}")"

REGISTRY_DOMAIN="$(prompt_with_default "请输入推送域名" "localhost:5000")"

FULL_IMAGE="${IMAGE_NAME}:${IMAGE_TAG}"
PUSH_IMAGE="${REGISTRY_DOMAIN}/${IMAGE_NAME}:${IMAGE_TAG}"

CMD=(docker build -f "${DOCKERFILE}" -t "${FULL_IMAGE}")
CMD+=(.)

echo
echo "将执行以下命令："
printf ' %q' "${CMD[@]}"
echo
echo

read -r -p "确认执行构建？[y/N]: " confirm
if [[ ! "${confirm}" =~ ^[Yy]$ ]]; then
	echo "已取消构建。"
	exit 0
fi

"${CMD[@]}"

docker tag "${FULL_IMAGE}" "${PUSH_IMAGE}"

echo
echo "构建完成：${FULL_IMAGE}"
echo "已打推送 Tag：${PUSH_IMAGE}"
echo "可直接推送：docker push ${PUSH_IMAGE}"