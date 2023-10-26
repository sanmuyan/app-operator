package v1

type DeployType string

const (
	StableDeploy DeployType = "stable"
	CanaryDeploy DeployType = "canary"
)

const (
	TureValue    = "true"
	FalseValue   = "false"
	NilValue     = ""
	LabelPrefix  = "app.sanmuyan.com"
	OperatorName = "app-operator"
	ApiKind      = "AppConfig"
	AppName      = "app"
)

const (
	AppConfigFinalizer = LabelPrefix + "/protected"
)

const (
	InjectionLabel = "injection"
	CreatedByLabel = "app.kubernetes.io/created-by"
)

const (
	// ContainersInjectionAnnotation 注入容器数组，值应该是 JSON
	ContainersInjectionAnnotation = "injection-containers"
	// ProtectedAnnotation 删除保护
	ProtectedAnnotation = "protected"
	// StrictUpdateAnnotation 严格更新模式
	StrictUpdateAnnotation = "strict-update"
	// DeploymentConfigAnnotation 每个 appConfig 单独配置，优先级高于全局模板，值应该是 JSON
	DeploymentConfigAnnotation = "deployment-config"
	// StrictReleaseAnnotation 严格发布模式
	StrictReleaseAnnotation = "strict-release"
	// CanaryIngressAnnotation 是否为 canary ingress
	CanaryIngressAnnotation = "canary-ingress"
	// CanaryRollingWeightAnnotation 在发布中实时切换 canary ingress 的权重
	CanaryRollingWeightAnnotation = "canary-rolling-weight"
	// IngressAnnotationsAnnotation ingress 追加的 annotations
	IngressAnnotationsAnnotation = "ingress-annotations"
)

const (
	DeleteProtectedMessage  = "cannot delete protected resources"
	VersionNotLatestMessage = "latest version and try again"
)

// 第三方注解
const (
	NginxIngressAnnotationPrefix = "nginx.ingress.kubernetes.io"
	NginxIngressCanaryAnnotation = NginxIngressAnnotationPrefix + "/canary"
	NginxIngressWeightAnnotation = NginxIngressAnnotationPrefix + "canary-weight"
)
