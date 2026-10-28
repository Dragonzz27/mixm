# 实例模板配置

templates.config 保存 JSON 字符串。批量启动使用 repo_id 与 repo_tag 指定已有镜像仓库，资源配置示例：

```json
{
  "repo_id": "已有镜像仓库ID",
  "repo_tag": "latest",
  "cpu_limit": 1.0,
  "memory_limit": "1Gi",
  "storage_limit": 1,
  "gpu_limit": 0,
  "gpu_memory_limit": 0
}
```

CPU 范围 0.5-16，内存范围 512Mi-32Gi，临时存储范围 1-40Gi，GPU 数量范围 0-8。前端模板编辑器使用 memory_limit_str 和 storage 字段，批量启动时会转换为实例所需的 memory_limit 和 storage_limit。

每个用户的持久 workspace 使用 edu-pvc-cs-{user_id}，容量由 USER_PERSONAL_PVC_CAPACITY 设置。storage_limit 控制容器临时存储，不改变已有用户 PVC 容量。PVC_ACCESS_MODE 与 STORAGE_CLASS_NAME 应匹配实际存储供应器。

code-server 镜像使用原入口命令，暴露 8080 和 22 端口；普通 Ubuntu 镜像初始化 sshd。Go SSH 网关通过 Kubernetes pods/exec 连接到运行中的容器。

新实例的 SSH 网关用户名为平台用户名加上实际保存的实例 UUID 去除连字符后的前 8 位，例如 `admin0086c788`。该后缀对应实例配置和访问地址中的同一个实例 ID，用于区分同一用户的多个实例；连接时使用实例配置中显示的完整 SSH 用户名。
