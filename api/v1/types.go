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
	FinalizerAppConfigs = "app.sanmuyan.com/protected"
)

const (
	// AnnotationContainersInjection 注入容器，值应该是 JSON
	AnnotationContainersInjection = "sidecar.sanmuyan.com/injection-containers"
	// AnnotationProtected 删除保护
	AnnotationProtected = "app.sanmuyan.com/protected"
	// AnnotationStrictUpdate 严格更新模式
	AnnotationStrictUpdate = "app.sanmuyan.com/strict-update"
	// AnnotationDeployment 每个 appConfig 单独配置，优先级高于全局模板，值应该是 JSON
	AnnotationDeployment = "app.sanmuyan.com/deployment"
	// AnnotationStrictRelease 严格发布模式
	AnnotationStrictRelease = "app.sanmuyan.com/strict-release"
)
