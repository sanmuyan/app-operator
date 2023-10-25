package v1

type DeployType string

const (
	StableDeploy DeployType = "stable"
	CanaryDeploy DeployType = "canary"
)

const (
	TureValue     = "true"
	SidecarPrefix = "sidecar.sanmuyan.com/injection"
)

const (
	AppConfigFinalizer = "app.sanmuyan.com/protected"
)

const (
	// ContainersInjectionAnnotation 注入容器数组，值应该是 JSON
	ContainersInjectionAnnotation = "sidecar.sanmuyan.com/injection-containers"
	// ProtectedAnnotation 删除保护
	ProtectedAnnotation = "app.sanmuyan.com/protected"
	// StrictUpdateAnnotation 严格更新模式
	StrictUpdateAnnotation = "app.sanmuyan.com/strict-update"
	// DeploymentConfigAnnotation 每个 appConfig 单独配置，优先级高于全局模板，值应该是 JSON
	DeploymentConfigAnnotation = "app.sanmuyan.com/deployment-config"
	// StrictReleaseAnnotation 严格发布模式
	StrictReleaseAnnotation = "app.sanmuyan.com/strict-release"
)

const (
	DeleteProtectedMessage  = "cannot delete protected resources"
	VersionNotLatestMessage = "latest version and try again"
)
