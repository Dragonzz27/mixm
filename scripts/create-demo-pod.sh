#!/bin/bash

# 此脚本用于生成 Code-Server 的 Kubernetes 部署清单
# 使用方法：bash scripts/create-demo-pod.sh <user_id>
# 随后会生成 scripts/code-server-<user_id>.yaml
# 请使用 kubectl apply -f <file_name> 来部署
# 访问地址为 http://<your-domain>:port/<user_id>/   （注意最后一个下划线）

if [ -z "$1" ]; then
  echo "Usage: $0 <user_id>"
  exit 1
fi

USER_ID=$1
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PASSWORD=$(openssl rand -base64 12)
FILE_NAME="${SCRIPT_DIR}/code-server-${USER_ID}.yaml"
MIDDLEWARE_NAME="stripprefix-${USER_ID}"

echo "========================================="
echo "User: ${USER_ID}"
echo "Password: ${PASSWORD}"
echo "========================================="
echo "Generating Kubernetes manifest..."

cat <<EOF > "${FILE_NAME}"
---
# 1. PVC
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: pvc-code-server-${USER_ID}
spec:
  accessModes: [ "ReadWriteOnce" ]
  resources:
    requests:
      storage: 1Gi

---
# 2. Deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: code-server-${USER_ID}
spec:
  replicas: 1
  selector:
    matchLabels:
      app: code-server-${USER_ID}
  template:
    metadata:
      labels:
        app: code-server-${USER_ID}
    spec:
      containers:
      - name: code-server
        image: codercom/code-server:latest
        workingDir: /home/coder/project
        ports:
        - containerPort: 8080
        env:
        - name: PASSWORD
          value: "${PASSWORD}"
        # 核心配置：告诉 Code-Server 它的前缀是什么
        - name: VSCODE_PROXY_URI
          value: /${USER_ID}
        # 核心修复：禁止 Code-Server 校验代理路径，强制它信任当前配置
        - name: CS_DISABLE_PROXY_URI_VALIDATION
          value: "1"
        resources:
          limits:
            cpu: "1"
            memory: "1Gi"
          requests:
            cpu: "0.5"
            memory: "512Mi"
        volumeMounts:
        - name: project-data
          mountPath: /home/coder/project
      volumes:
      - name: project-data
        persistentVolumeClaim:
          claimName: pvc-code-server-${USER_ID}

---
# 3. Service
apiVersion: v1
kind: Service
metadata:
  name: svc-code-server-${USER_ID}
spec:
  type: ClusterIP
  selector:
    app: code-server-${USER_ID}
  ports:
  - port: 80
    targetPort: 8080

---
# 4. Middleware (回归最稳定的 StripPrefix)
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: ${MIDDLEWARE_NAME}
spec:
  stripPrefix:
    prefixes:
      - /${USER_ID}
    # 强制加上斜杠，防止路径格式混乱
    forceSlash: true

---
# 5. Ingress
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ingress-code-server-${USER_ID}
  annotations:
    traefik.ingress.kubernetes.io/router.middlewares: default-${MIDDLEWARE_NAME}@kubernetescrd
spec:
  ingressClassName: traefik
  rules:
  - http:
      paths:
      - path: /${USER_ID}
        pathType: Prefix
        backend:
          service:
            name: svc-code-server-${USER_ID}
            port:
              number: 80
EOF

echo "Success! File generated: ${FILE_NAME}"
echo "Next: kubectl apply -f ${FILE_NAME}"
