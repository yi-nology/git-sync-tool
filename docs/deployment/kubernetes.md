# Kubernetes 部署

Kubernetes 部署适合需要高可用、自动扩缩容的生产环境。

## 前提条件

- Kubernetes 集群（1.20+）
- kubectl 已配置并连接到集群
- Helm 3（可选，用于 Helm 部署）

## 快速部署

### 使用 kubectl

```bash
# 克隆仓库
git clone https://github.com/yi-nology/git-manage-service.git
cd git-manage-service

# 创建命名空间
kubectl create namespace git-manage

# 部署
kubectl apply -f deploy/k8s/ -n git-manage
```

### 使用 Helm

```bash
# 添加仓库（如果已发布）
helm repo add git-manage-service https://yi-nology.github.io/git-manage-service

# 安装
helm install git-manage-service git-manage-service/git-manage-service \
  --namespace git-manage \
  --create-namespace
```

## 部署清单

### Namespace

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: git-manage
```

### Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: git-manage-service
  namespace: git-manage
spec:
  replicas: 2
  selector:
    matchLabels:
      app: git-manage-service
  template:
    metadata:
      labels:
        app: git-manage-service
    spec:
      containers:
      - name: git-manage-service
        image: ghcr.io/yi-nology/git-manage-service:latest
        ports:
        - containerPort: 38080
          name: http
        - containerPort: 8888
          name: rpc
        env:
        - name: DB_TYPE
          value: "mysql"
        - name: DB_HOST
          value: "mysql-service"
        - name: DB_PORT
          value: "3306"
        - name: DB_USER
          valueFrom:
            secretKeyRef:
              name: git-manage-secret
              key: db-user
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: git-manage-secret
              key: db-password
        - name: DB_NAME
          value: "git_manage"
        resources:
          requests:
            cpu: "100m"
            memory: "256Mi"
          limits:
            cpu: "1000m"
            memory: "1Gi"
        livenessProbe:
          httpGet:
            path: /api/health
            port: 38080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /api/health
            port: 38080
          initialDelaySeconds: 5
          periodSeconds: 5
        volumeMounts:
        - name: repos
          mountPath: /app/repos
        - name: data
          mountPath: /app/data
      volumes:
      - name: repos
        persistentVolumeClaim:
          claimName: git-manage-repos-pvc
      - name: data
        persistentVolumeClaim:
          claimName: git-manage-data-pvc
```

### Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: git-manage-service
  namespace: git-manage
spec:
  selector:
    app: git-manage-service
  ports:
  - name: http
    port: 38080
    targetPort: 38080
  - name: rpc
    port: 8888
    targetPort: 8888
  type: ClusterIP
```

### Ingress

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: git-manage-ingress
  namespace: git-manage
  annotations:
    nginx.ingress.kubernetes.io/proxy-body-size: "100m"
spec:
  ingressClassName: nginx
  rules:
  - host: git-manage.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: git-manage-service
            port:
              number: 38080
  tls:
  - hosts:
    - git-manage.example.com
    secretName: git-manage-tls
```

### PVC

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: git-manage-repos-pvc
  namespace: git-manage
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 50Gi
  storageClassName: standard

---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: git-manage-data-pvc
  namespace: git-manage
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
  storageClassName: standard
```

### Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: git-manage-secret
  namespace: git-manage
type: Opaque
stringData:
  db-user: root
  db-password: your_password
```

## 配置说明

### 环境变量配置

使用 ConfigMap 管理非敏感配置：

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: git-manage-config
  namespace: git-manage
data:
  DB_TYPE: "mysql"
  DB_HOST: "mysql-service"
  DB_PORT: "3306"
  DB_NAME: "git_manage"
  REDIS_ADDR: "redis-service:6379"
```

### 资源配置

根据实际负载调整：

| 环境 | CPU | 内存 |
|------|-----|------|
| 开发 | 100m - 500m | 256Mi - 512Mi |
| 测试 | 500m - 1000m | 512Mi - 1Gi |
| 生产 | 1000m - 2000m | 1Gi - 2Gi |

### 副本数

```yaml
spec:
  replicas: 2  # 生产环境建议 2+
```

## HPA（自动扩缩容）

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: git-manage-hpa
  namespace: git-manage
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: git-manage-service
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

## 常用命令

```bash
# 查看部署状态
kubectl get all -n git-manage

# 查看 Pod 日志
kubectl logs -f deployment/git-manage-service -n git-manage

# 进入 Pod
kubectl exec -it deployment/git-manage-service -n git-manage -- sh

# 扩缩容
kubectl scale deployment/git-manage-service --replicas=3 -n git-manage

# 更新镜像
kubectl set image deployment/git-manage-service \
  git-manage-service=ghcr.io/yi-nology/git-manage-service:vX.X.X \
  -n git-manage

# 查看资源使用
kubectl top pods -n git-manage
```

## 升级

```bash
# 更新镜像版本
kubectl set image deployment/git-manage-service \
  git-manage-service=ghcr.io/yi-nology/git-manage-service:v0.7.2 \
  -n git-manage

# 查看滚动更新状态
kubectl rollout status deployment/git-manage-service -n git-manage

# 回滚
kubectl rollout undo deployment/git-manage-service -n git-manage
```

## 监控和日志

### Prometheus 监控

添加 Prometheus annotations：

```yaml
metadata:
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "38080"
    prometheus.io/path: "/metrics"
```

### 日志采集

使用 Filebeat 或 Fluentd 采集日志。

## 故障排查

### Pod 无法启动

```bash
# 查看 Pod 状态
kubectl describe pod <pod-name> -n git-manage

# 查看事件
kubectl get events -n git-manage --sort-by='.lastTimestamp'
```

### 服务无法访问

```bash
# 检查 Service
kubectl get svc -n git-manage

# 检查 Endpoints
kubectl get endpoints -n git-manage
```

## 下一步

- [配置参考](/configuration) - 完整配置说明
- [API 文档](/api) - HTTP API 参考
