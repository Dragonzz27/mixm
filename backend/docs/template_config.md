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

CPU 范围 0.5-16，内存范围 512Mi-32Gi，临时存储范围 1-40Gi，GPU 数量范围 0-8。旧配置的 memory_limit_str 和 storage（例如 1Gi）会在批量启动时归一化。

每个用户的持久 workspace 使用 edu-pvc-cs-{user_id}，容量由 USER_PERSONAL_PVC_CAPACITY 设置。storage_limit 控制容器临时存储，不改变已有用户 PVC 容量。PVC_ACCESS_MODE 与 STORAGE_CLASS_NAME 应匹配实际存储供应器。

code-server 镜像使用原入口命令，暴露 8080 和 22 端口；普通 Ubuntu 镜像初始化 sshd。Go SSH 网关通过 Kubernetes pods/exec 连接到运行中的容器。
