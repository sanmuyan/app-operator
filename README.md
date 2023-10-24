# app-operator

基于 `Kubebuilder` 脚手架项目，可用搭建出 `Kubernetes API` 以及它所需要的 `CRD` `controller` `webhook`。  
这个项目偏教程性质，通过本项目可以知道怎么快速搭建一个 `Operator`。

## 开发环境

- `Ubuntu 22.4`
- `Go 1.21`
- `Kubernetes v1.20.6`
- `Kubebuilder 3.12.0`

## 创建项目

### 下载工具

```shell
curl -L -o kubebuilder "https://go.kubebuilder.io/dl/latest/$(go env GOOS)/$(go env GOARCH)"
chmod +x kubebuilder && mv kubebuilder /usr/local/bin/
```

### 创建项目

```shell
cd $GOPATH/src
mkdir app-operator
cd app-operator
kubebuilder init --domain sanmuyan.com --repo sanmuyan.com/app-operator
```

### 添加 API

```shell
kubebuilder create api --group app --version v1 --kind AppConfig
```

### 添加 Webhook

```shell
kubebuilder create webhook --group app --version v1 --kind AppConfig --defaulting --programmatic-validation
```

## 开发部署

### 项目目录

- `api/v1/appconfig_types.go` `CRD` 字段定义
- `api/v1/appconfig_webhook.go` `webhook` 业务逻辑
- `internal/controller/appconfig_controller.go` `controller` 业务逻辑

### 安装 CRD

```shell
make install
```

### 运行 Controller

```shell
make run
```

### 部署项目

```shell
make deploy
```

## 注意事项

### RBAC

#### 注解

增加管理资源要在 `controller` 中加上注解 `//+kubebuilder:rbac:groups=*,resources=deployments,verbs=*`

### Webhook

#### Cert manager

`Webhook` 需要签名证书，需要手动部署一下 `cert-manager`

```shell
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.1/cert-manager.yaml
```

#### 配置

第一次添加 `webhook` 添加后需要到 `config/default/kustomization.yaml` 打开相关的部署配置注释

#### 核心类型 Webhook

`Kubebuilder` 不支持核心类型的 `webhook`，比如 `pod`，需要手工实现

1. 添加一个 `webhook` 逻辑，参考 `api/pod_webhook.go`
2. 在程序入口注册 `webhook`，参考 `cmd/mian.go`

#### 如何在本地调试 Webhook

1. 在 `app-operator-system` 命名空间中起一个 `local-proxy-deployment`
2. `app-operator-webhook-service` 指向 `local-proxy-deployment`

```shell
server {
    listen       9443  default_server ssl;
    server_name  _;
    ssl_certificate /data/cert/tls.crt;
    ssl_certificate_key /data/cert/tls.key;
    location / {
       proxy_pass https://<YOUR_DEV_ENV>:9443;
    }
}    
```