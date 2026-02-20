#!/bin/bash

# Git Manage Service K8s 部署脚本
# 功能：部署、卸载、重启、状态查看

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
NAMESPACE="default"
APP_NAME="git-manage"
FRONTEND_PVC="git-manage-frontend-pvc"

# 打印函数
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_banner() {
    echo "========================================"
    echo "  Git Manage Service K8s 部署脚本"
    echo "========================================"
    echo ""
}

# 检查前置条件
check_prerequisites() {
    print_info "检查前置条件..."
    
    # 检查 kubectl
    if ! command -v kubectl &> /dev/null; then
        print_error "kubectl 未安装，请先安装 kubectl"
        exit 1
    fi
    
    # 检查集群连接
    if ! kubectl cluster-info &> /dev/null; then
        print_error "无法连接到 Kubernetes 集群"
        exit 1
    fi
    
    print_success "前置条件检查通过"
}

# 检查前端资源
check_frontend_assets() {
    if [ ! -d "../../public" ]; then
        print_warning "前端资源目录 public/ 不存在"
        print_info "请先运行: make build-frontend-integrate"
        return 1
    fi
    
    if [ ! -f "../../public/index.html" ]; then
        print_warning "前端资源未构建"
        print_info "请先运行: make build-frontend-integrate"
        return 1
    fi
    
    return 0
}

# 上传前端资源到 PVC
upload_frontend_assets() {
    print_info "上传前端资源到 PVC..."
    
    # 检查前端资源
    if ! check_frontend_assets; then
        print_error "前端资源检查失败"
        exit 1
    fi
    
    # 检查 PVC 是否存在
    if ! kubectl get pvc $FRONTEND_PVC -n $NAMESPACE &> /dev/null; then
        print_warning "PVC $FRONTEND_PVC 不存在，将自动创建"
    fi
    
    # 创建临时 Pod 上传资源
    print_info "创建临时 Pod..."
    kubectl run frontend-uploader \
        --image=busybox \
        --restart=Never \
        --namespace=$NAMESPACE \
        --overrides='
{
  "spec": {
    "containers": [{
      "name": "uploader",
      "image": "busybox",
      "command": ["sh", "-c", "sleep 3600"],
      "volumeMounts": [{
        "name": "frontend",
        "mountPath": "/data"
      }]
    }],
    "volumes": [{
      "name": "frontend",
      "persistentVolumeClaim": {
        "claimName": "'$FRONTEND_PVC'"
      }
    }]
  }
}' 2>/dev/null || true
    
    # 等待 Pod 就绪
    print_info "等待 Pod 就绪..."
    kubectl wait --for=condition=Ready pod/frontend-uploader -n $NAMESPACE --timeout=60s
    
    # 清空旧资源
    print_info "清空旧的前端资源..."
    kubectl exec frontend-uploader -n $NAMESPACE -- sh -c "rm -rf /data/*" 2>/dev/null || true
    
    # 复制前端资源
    print_info "复制前端资源..."
    kubectl cp ../../public/. $NAMESPACE/frontend-uploader:/data/
    
    # 验证上传
    print_info "验证上传结果..."
    if kubectl exec frontend-uploader -n $NAMESPACE -- ls -la /data/index.html &> /dev/null; then
        print_success "前端资源上传成功"
    else
        print_error "前端资源上传失败"
        kubectl delete pod frontend-uploader -n $NAMESPACE --force --grace-period=0 2>/dev/null || true
        exit 1
    fi
    
    # 清理临时 Pod
    print_info "清理临时 Pod..."
    kubectl delete pod frontend-uploader -n $NAMESPACE --force --grace-period=0 2>/dev/null || true
    
    print_success "前端资源准备完成"
}

# 部署应用
deploy() {
    print_banner
    print_info "开始部署 Git Manage Service..."
    
    check_prerequisites
    
    # 1. 创建 Secret 和 ConfigMap
    print_info "创建 Secret 和 ConfigMap..."
    kubectl apply -f secret.yaml -n $NAMESPACE
    kubectl apply -f configmap.yaml -n $NAMESPACE
    kubectl apply -f nginx-configmap.yaml -n $NAMESPACE
    
    # 2. 部署数据库（如果需要）
    if [ -f "mysql.yaml" ]; then
        print_info "部署 MySQL 数据库..."
        kubectl apply -f mysql.yaml -n $NAMESPACE
        
        # 等待数据库就绪
        print_info "等待数据库就绪..."
        sleep 10
    fi
    
    # 3. 部署后端
    print_info "部署后端服务..."
    kubectl apply -f deployment.yaml -n $NAMESPACE
    kubectl apply -f service.yaml -n $NAMESPACE
    
    # 4. 上传前端资源
    upload_frontend_assets
    
    # 5. 部署 Nginx
    print_info "部署 Nginx 前端服务..."
    kubectl apply -f nginx-deployment.yaml -n $NAMESPACE
    
    # 6. 部署 Ingress（可选）
    if [ -f "ingress.yaml" ]; then
        read -p "是否部署 Ingress？(y/n) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            print_info "部署 Ingress..."
            kubectl apply -f ingress.yaml -n $NAMESPACE
        fi
    fi
    
    # 7. 等待部署完成
    print_info "等待部署完成..."
    kubectl rollout status deployment/git-manage-backend -n $NAMESPACE --timeout=300s
    kubectl rollout status deployment/git-manage-nginx -n $NAMESPACE --timeout=300s
    
    print_success "部署完成！"
    echo ""
    show_status
    echo ""
    show_access_info
}

# 卸载应用
uninstall() {
    print_banner
    print_warning "准备卸载 Git Manage Service..."
    
    read -p "确认卸载？这将删除所有资源！(y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "取消卸载"
        exit 0
    fi
    
    print_info "开始卸载..."
    
    # 删除 Ingress
    if kubectl get ingress git-manage-ingress -n $NAMESPACE &> /dev/null; then
        print_info "删除 Ingress..."
        kubectl delete ingress git-manage-ingress -n $NAMESPACE
    fi
    
    # 删除 Deployment 和 Service
    print_info "删除 Deployment 和 Service..."
    kubectl delete deployment git-manage-backend git-manage-nginx -n $NAMESPACE --ignore-not-found=true
    kubectl delete service git-manage-backend git-manage-nginx -n $NAMESPACE --ignore-not-found=true
    
    # 删除数据库
    if kubectl get deployment git-manage-mysql -n $NAMESPACE &> /dev/null; then
        print_info "删除 MySQL..."
        kubectl delete -f mysql.yaml -n $NAMESPACE --ignore-not-found=true
    fi
    
    # 删除 ConfigMap 和 Secret
    print_info "删除 ConfigMap 和 Secret..."
    kubectl delete configmap git-manage-config nginx-config -n $NAMESPACE --ignore-not-found=true
    kubectl delete secret git-manage-secret -n $NAMESPACE --ignore-not-found=true
    
    # 询问是否删除 PVC
    read -p "是否删除 PVC（包括数据）？(y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        print_info "删除 PVC..."
        kubectl delete pvc git-manage-data-pvc git-manage-repos-pvc git-manage-frontend-pvc -n $NAMESPACE --ignore-not-found=true
    else
        print_info "保留 PVC 和数据"
    fi
    
    print_success "卸载完成！"
}

# 重启应用
restart() {
    print_banner
    print_info "重启 Git Manage Service..."
    
    # 重启后端
    print_info "重启后端服务..."
    kubectl rollout restart deployment/git-manage-backend -n $NAMESPACE
    
    # 重启 Nginx
    print_info "重启 Nginx 前端服务..."
    kubectl rollout restart deployment/git-manage-nginx -n $NAMESPACE
    
    # 等待重启完成
    print_info "等待重启完成..."
    kubectl rollout status deployment/git-manage-backend -n $NAMESPACE --timeout=300s
    kubectl rollout status deployment/git-manage-nginx -n $NAMESPACE --timeout=300s
    
    print_success "重启完成！"
    echo ""
    show_status
}

# 查看状态
show_status() {
    print_info "应用状态："
    echo ""
    
    echo "========== Deployments =========="
    kubectl get deployment -l app=$APP_NAME -n $NAMESPACE
    echo ""
    
    echo "========== Pods =========="
    kubectl get pods -l app=$APP_NAME -n $NAMESPACE
    echo ""
    
    echo "========== Services =========="
    kubectl get service -l app=$APP_NAME -n $NAMESPACE
    echo ""
    
    if kubectl get ingress git-manage-ingress -n $NAMESPACE &> /dev/null; then
        echo "========== Ingress =========="
        kubectl get ingress git-manage-ingress -n $NAMESPACE
        echo ""
    fi
    
    echo "========== PVC =========="
    kubectl get pvc -n $NAMESPACE | grep git-manage
}

# 显示访问信息
show_access_info() {
    print_info "访问方式："
    echo ""
    
    # 获取 Nginx Service 信息
    SERVICE_TYPE=$(kubectl get service git-manage-nginx -n $NAMESPACE -o jsonpath='{.spec.type}')
    
    if [ "$SERVICE_TYPE" == "LoadBalancer" ]; then
        EXTERNAL_IP=$(kubectl get service git-manage-nginx -n $NAMESPACE -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
        if [ -z "$EXTERNAL_IP" ]; then
            EXTERNAL_IP=$(kubectl get service git-manage-nginx -n $NAMESPACE -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
        fi
        
        if [ -n "$EXTERNAL_IP" ]; then
            echo "1. 通过 LoadBalancer 访问："
            echo "   http://$EXTERNAL_IP"
        else
            echo "1. LoadBalancer 正在分配外部 IP，请稍候..."
            echo "   运行查看：kubectl get svc git-manage-nginx -n $NAMESPACE"
        fi
    elif [ "$SERVICE_TYPE" == "NodePort" ]; then
        NODE_PORT=$(kubectl get service git-manage-nginx -n $NAMESPACE -o jsonpath='{.spec.ports[0].nodePort}')
        echo "1. 通过 NodePort 访问："
        echo "   http://<NODE-IP>:$NODE_PORT"
    fi
    
    echo ""
    echo "2. 通过端口转发访问（测试）："
    echo "   kubectl port-forward svc/git-manage-nginx 8080:80 -n $NAMESPACE"
    echo "   然后访问：http://localhost:8080"
    
    if kubectl get ingress git-manage-ingress -n $NAMESPACE &> /dev/null; then
        INGRESS_HOST=$(kubectl get ingress git-manage-ingress -n $NAMESPACE -o jsonpath='{.spec.rules[0].host}')
        echo ""
        echo "3. 通过 Ingress 访问："
        echo "   http://$INGRESS_HOST"
    fi
}

# 查看日志
show_logs() {
    print_info "查看应用日志..."
    
    echo "选择要查看的日志："
    echo "1) Backend"
    echo "2) Nginx"
    echo "3) 全部"
    read -p "请选择 (1-3): " -n 1 -r
    echo
    
    case $REPLY in
        1)
            kubectl logs -f deployment/git-manage-backend -n $NAMESPACE --tail=100
            ;;
        2)
            kubectl logs -f deployment/git-manage-nginx -n $NAMESPACE --tail=100
            ;;
        3)
            kubectl logs -f deployment/git-manage-backend -n $NAMESPACE --tail=50 &
            kubectl logs -f deployment/git-manage-nginx -n $NAMESPACE --tail=50
            ;;
        *)
            print_error "无效选择"
            exit 1
            ;;
    esac
}

# 更新前端
update_frontend() {
    print_info "更新前端资源..."
    upload_frontend_assets
    
    print_info "重启 Nginx..."
    kubectl rollout restart deployment/git-manage-nginx -n $NAMESPACE
    kubectl rollout status deployment/git-manage-nginx -n $NAMESPACE --timeout=300s
    
    print_success "前端更新完成！"
}

# 扩缩容
scale() {
    print_info "扩缩容应用..."
    
    read -p "输入 Backend 副本数: " backend_replicas
    read -p "输入 Nginx 副本数: " nginx_replicas
    
    print_info "设置副本数..."
    kubectl scale deployment/git-manage-backend --replicas=$backend_replicas -n $NAMESPACE
    kubectl scale deployment/git-manage-nginx --replicas=$nginx_replicas -n $NAMESPACE
    
    print_info "等待扩缩容完成..."
    kubectl rollout status deployment/git-manage-backend -n $NAMESPACE --timeout=300s
    kubectl rollout status deployment/git-manage-nginx -n $NAMESPACE --timeout=300s
    
    print_success "扩缩容完成！"
    show_status
}

# 显示帮助
show_help() {
    print_banner
    echo "用法: $0 [命令]"
    echo ""
    echo "命令："
    echo "  deploy          部署应用"
    echo "  uninstall       卸载应用"
    echo "  restart         重启应用"
    echo "  status          查看状态"
    echo "  logs            查看日志"
    echo "  update-frontend 更新前端"
    echo "  scale           扩缩容"
    echo "  help            显示帮助"
    echo ""
    echo "示例："
    echo "  $0 deploy       # 部署应用"
    echo "  $0 restart      # 重启应用"
    echo "  $0 status       # 查看状态"
    echo ""
}

# 主函数
main() {
    case "${1:-}" in
        deploy)
            deploy
            ;;
        uninstall)
            uninstall
            ;;
        restart)
            restart
            ;;
        status)
            show_status
            ;;
        logs)
            show_logs
            ;;
        update-frontend)
            update_frontend
            ;;
        scale)
            scale
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            print_error "未知命令: ${1:-}"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"
