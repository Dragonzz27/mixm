# 模板配置说明

模板配置使用 JSON 格式，支持以下配置项：

## 基础配置

```json
{
  "namespace": "default",
  "replicas": 1,
  "image": "codercom/code-server:latest",
  "container_port": 8080,
  "service_port": 80
}
```

## 资源配置

```json
{
  "cpu_limit": "1",
  "memory_limit_str": "1Gi",
  "cpu_request": "0.5",
  "memory_request": "512Mi",
  "storage": "1Gi"
}
```

## 环境变量

```json
{
  "env": [
    {"name": "MY_VAR", "value": "my_value"},
    {"name": "ANOTHER_VAR", "value": "another_value"}
  ]
}
```

## 完整示例

```json
{
  "namespace": "default",
  "replicas": 1,
  "image": "codercom/code-server:latest",
  "container_port": 8080,
  "service_port": 80,
  "cpu_limit": "1",
  "memory_limit_str": "1Gi",
  "cpu_request": "0.5",
  "memory_request": "512Mi",
  "storage": "1Gi",
  "env": [
    {"name": "CUSTOM_ENV", "value": "custom_value"}
  ]
}
```

## 配置项说明

- `namespace`: Kubernetes 命名空间
- `replicas`: 副本数
- `image`: 容器镜像
- `container_port`: 容器端口
- `service_port`: Service 端口
- `cpu_limit`: CPU 限制
- `memory_limit_str`: 内存限制（字符串格式，如 "1Gi"）
- `cpu_request`: CPU 请求
- `memory_request`: 内存请求
- `storage`: 存储大小
- `env`: 自定义环境变量数组
- `memory_limit`: 内存限制（数字格式，单位 MB，用于数据库记录）

## 优先级

实例创建时，配置的优先级如下：
1. 用户传入的参数（最高优先级）
2. 模板配置
3. 默认值（最低优先级）
