#!/usr/bin/env bash

set -euo pipefail

torch_version="${1:?缺少 PyTorch 版本}"
torchvision_version="${2:?缺少 torchvision 版本}"
torch_device="${3:?缺少计算设备}"
cuda_tag="${4:?缺少 CUDA 通道}"

case "${torch_device}" in
    cpu) channel=cpu ;;
    cuda) channel="${cuda_tag}" ;;
    *)
        echo "Invalid TORCH_DEVICE=${torch_device}, expected: cpu or cuda" >&2
        exit 1
        ;;
esac

uv pip install \
    --python "${VIRTUAL_ENV}/bin/python" \
    --index-url "https://download.pytorch.org/whl/${channel}" \
    "torch==${torch_version}" "torchvision==${torchvision_version}"
