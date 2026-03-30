# 智泉

## TODO
现在在HyperV中进行的部署k3s，但是code-server在目前iframe里可以打开，但输入密码后还是重定向到密码框。cd

## 部署

### backend

1. k3s部署后，提取kubeconfig文件，保存至/backend/app/kubeconfig/k3s.yaml
2. 检查.env中是否为 K8S_CONFIG_PATH=kubeconfig/k3s.yaml
3. 注意cluster.cluster.server 非127.0.0.1

部署docker-compose.yaml采用nerdctl，注意postgresql文件夹需赋予权限