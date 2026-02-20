# Kubernetes 部署指南

本目录包含 Git Manage Service 在 Kubernetes 集群中的部署配置。

## 架构说明

```
┌─────────────────┐
│    Ingress      │ ← 入口（可选，域名路由）
└─────────────────┘
         ↓
    ┌────────┴────────┐
    ↓                 ↓
┌─────────┐     ┌──────────┐
│  Nginx  │     │ Backend  │
│  (前端)  │     │  (API)   │
└─────────┘     └──────────┘
     ↓               ↓
  Frontend       ┌─────┴─────┐
   Assets        ↓           ↓
              MySQL      Redis
              /Postgres  /MinIO
```

**核心组件**：
- **Nginx Pod**: 提供前端静态资源（2 副本）
- **Backend Pod**: Go API 服务（2 副本）
- **MySQL/PostgreSQL**: 数据库
- **Ingress**: 可选的统一入口（基于域名路由）

## 文件说明

| 文件 | 说明 |
|------|------|
| `nginx-configmap.yaml` | Nginx 配置（反向代理 + 静态资源） |
| `nginx-deployment.yaml` | Nginx Deployment + Service + PVC |
| `deployment.yaml` | 后端 Deployment + PVC |
| `service.yaml` | 后端 Service（ClusterIP） |
| `ingress.yaml` | Ingress 配置（可选） |
| `configmap.yaml` | 应用配置 |
| `secret.yaml` | 敏感信息（密码、密钥） |
| `mysql.yaml` | MySQL 数据库部署 |

## 部署步骤

### 1. 准备前端资源

首先需要构建前端并准备资源：

```bash
# 在项目根目录执行
make build-frontend-integrate

# 将前端资源复制到 PVC（方式一：使用临时 Pod）
kubectl create -f - <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: frontend-uploader
spec:
  containers:
  - name: uploader
    image: busybox
    command: ['sh', '-c', 'sleep 3600']
    volumeMounts:
    - name: frontend
      mountPath: /data
  volumes:
  - name: frontend
    persistentVolumeClaim:
      claimName: git-manage-frontend-pvc
EOF

# 等待 Pod 就绪
kubectl wait --for=condition=Ready pod/frontend-uploader

# 复制前端资源
kubectl cp public/. frontend-uploader:/data/

# 清理临时 Pod
kubectl delete pod frontend-uploader
```

### 2. 创建 Secret 和 ConfigMap

```bash
# 创建 Secret（修改密码为实际值）
kubectl apply -f secret.yaml

# 创建 ConfigMap
kubectl apply -f configmap.yaml

# 创建 Nginx ConfigMap
kubectl apply -f nginx-configmap.yaml
```

### 3. 部署数据库（可选）

如果使用集群内数据库：

```bash
kubectl apply -f mysql.yaml
```

如果使用外部数据库，修改 `configmap.yaml` 中的数据库配置。

### 4. 部署应用

```bash
# 部署后端
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml

# 部署 Nginx（前端）
kubectl apply -f nginx-deployment.yaml

# 部署 Ingress（可选）
kubectl apply -f ingress.yaml
```

### 5. 验证部署

```bash
# 查看所有资源
kubectl get all -l app=git-manage

# 查看 Pod 状态
kubectl get pods -l app=git-manage

# 查看日志
kubectl logs -f deployment/git-manage-backend
kubectl logs -f deployment/git-manage-nginx

# 查看服务
kubectl get svc -l app=git-manage
```

## 访问方式

### 方式一：通过 LoadBalancer（如果支持）

```bash
# 获取 Nginx Service 的外部 IP
kubectl get svc git-manage-nginx

# 访问
curl http://<EXTERNAL-IP>
```

### 方式二：通过 NodePort

修改 `nginx-deployment.yaml` 中的 Service type 为 `NodePort`：

```yaml
spec:
  type: NodePort
```

然后访问：`http://<NODE-IP>:<NODE-PORT>`

### 方式三：通过 Ingress

如果部署了 Ingress：

```bash
# 配置 hosts（如果没有 DNS）
echo "<INGRESS-IP> git-manage.example.com" >> /etc/hosts

# 访问
curl http://git-manage.example.com
```

### 方式四：端口转发（测试）

```bash
# 转发 Nginx 端口
kubectl port-forward svc/git-manage-nginx 8080:80

# 访问
curl http://localhost:8080
```

## 更新前端资源

当前端代码更新后：

```bash
# 1. 重新构建前端
make build-frontend-integrate

# 2. 更新 PVC 中的资源（使用临时 Pod）
kubectl create -f - <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: frontend-updater
spec:
  containers:
  - name: updater
    image: busybox
    command: ['sh', '-c', 'rm -rf /data/* && sleep 3600']
    volumeMounts:
    - name: frontend
      mountPath: /data
  volumes:
  - name: frontend
    persistentVolumeClaim:
      claimName: git-manage-frontend-pvc
EOF

kubectl wait --for=condition=Ready pod/frontend-updater
kubectl cp public/. frontend-updater:/data/
kubectl delete pod frontend-updater

# 3. 重启 Nginx Pod
kubectl rollout restart deployment/git-manage-nginx
```

## 扩缩容

```bash
# 扩展后端副本
kubectl scale deployment/git-manage-backend --replicas=3

# 扩展 Nginx 副本
kubectl scale deployment/git-manage-nginx --replicas=3

# 查看副本状态
kubectl get deployment -l app=git-manage
```

## 健康检查

后端配置了健康检查：
- **Liveness Probe**: `/api/v1/ping`（存活检查）
- **Readiness Probe**: `/api/v1/ping`（就绪检查）

Nginx 配置了健康检查：
- **Health Endpoint**: `/health`

## 资源配置

### 后端资源限制

```yaml
resources:
  requests:
    cpu: 500m
    memory: 512Mi
  limits:
    cpu: 2000m
    memory: 2Gi
```

### Nginx 资源限制

```yaml
resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 512Mi
```

根据实际负载调整资源配置。

## 故障排查

### Pod 启动失败

```bash
# 查看 Pod 详情
kubectl describe pod <pod-name>

# 查看日志
kubectl logs <pod-name>

# 进入容器
kubectl exec -it <pod-name> -- sh
```

### 前端无法访问

```bash
# 检查 Nginx Pod
kubectl logs -f deployment/git-manage-nginx

# 检查前端资源是否存在
kubectl exec -it deployment/git-manage-nginx -- ls -la /usr/share/nginx/html

# 检查 ConfigMap
kubectl get configmap nginx-config -o yaml
```

### API 无法访问

```bash
# 检查后端 Pod
kubectl logs -f deployment/git-manage-backend

# 检查服务
kubectl get svc git-manage-backend

# 测试服务连通性
kubectl run test --rm -it --image=curlimages/curl -- curl http://git-manage-backend:8080/api/v1/ping
```

## 清理资源

```bash
# 删除所有资源
kubectl delete -f .

# 或者逐个删除
kubectl delete ingress git-manage-ingress
kubectl delete deployment git-manage-nginx git-manage-backend
kubectl delete svc git-manage-nginx git-manage-backend
kubectl delete configmap git-manage-config nginx-config
kubectl delete secret git-manage-secret
kubectl delete pvc git-manage-data-pvc git-manage-repos-pvc git-manage-frontend-pvc
```

## 生产环境建议

1. **使用专用 Namespace**
   ```bash
   kubectl create namespace git-manage
   # 在所有 yaml 中修改 namespace
   ```

2. **配置资源限额**
   ```yaml
   apiVersion: v1
   kind: ResourceQuota
   metadata:
     name: git-manage-quota
   spec:
     hard:
       requests.cpu: "4"
       requests.memory: 8Gi
       limits.cpu: "8"
       limits.memory: 16Gi
   ```

3. **配置网络策略**（限制 Pod 间通信）

4. **使用 StatefulSet**（如果需要固定 Pod 名称）

5. **配置 HPA**（自动扩缩容）
   ```yaml
   apiVersion: autoscaling/v2
   kind: HorizontalPodAutoscaler
   metadata:
     name: git-manage-backend-hpa
   spec:
     scaleTargetRef:
       apiVersion: apps/v1
       kind: Deployment
       name: git-manage-backend
     minReplicas: 2
     maxReplicas: 10
     metrics:
     - type: Resource
       resource:
         name: cpu
         target:
           type: Utilization
           averageUtilization: 70
   ```

6. **配置 PodDisruptionBudget**（保证可用性）

7. **使用 Helm Chart**（简化部署）
