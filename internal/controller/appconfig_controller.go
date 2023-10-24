/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/yaml"
	"os"
	"sanmuyan.com/app-operator/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appv1 "sanmuyan.com/app-operator/api/v1"
)

// AppConfigReconciler reconciles a AppConfig object
type AppConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=app.sanmuyan.com,resources=appconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.sanmuyan.com,resources=appconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=app.sanmuyan.com,resources=appconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups=*,resources=deployments,verbs=*
//+kubebuilder:rbac:groups=*,resources=services,verbs=*
//+kubebuilder:rbac:groups=*,resources=ingresses,verbs=*

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// the AppConfig object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.0/pkg/reconcile
func (r *AppConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Controller 业务逻辑

	_ = log.FromContext(ctx)

	// 获取 AppConfig 对象
	acLog.Info("reconciling appConfig", "namespace", req.Namespace, "name", req.Name)
	ac := &appv1.AppConfig{}
	if err := r.Get(ctx, req.NamespacedName, ac); err != nil {
		acLog.Info("failed to get appConfigs", "err", err)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 添加 finalizer
	if ac.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(ac, appv1.FinalizerAppConfigs) {
			controllerutil.AddFinalizer(ac, appv1.FinalizerAppConfigs)
			if err := r.Update(ctx, ac); err != nil {
				acLog.Error(err, "failed to add finalizer")
				return ctrl.Result{}, err
			}
		}
	} else {
		if controllerutil.ContainsFinalizer(ac, appv1.FinalizerAppConfigs) {
			controllerutil.RemoveFinalizer(ac, appv1.FinalizerAppConfigs)
			if err := r.Update(ctx, ac); err != nil {
				acLog.Info("failed to remove finalizer", "err", err)
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// 判断是否处于暂停状态
	if ac.Spec.Paused {
		acLog.Info("appConfig is paused, skip update", "namespace", req.Namespace, "name", req.Name)
		return ctrl.Result{}, nil
	}

	dmList := &appsv1.DeploymentList{}
	if err := r.List(ctx, dmList, client.InNamespace(ac.Namespace), client.MatchingFields{ownerKey: req.Name}); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	dmMap := make(map[string]*appsv1.Deployment)
	for i, dm := range dmList.Items {
		dmMap[dm.Name] = &dmList.Items[i]
	}

	// 更新 AppConfigs 的状态
	ac.Status = appv1.AppConfigStatus{
		AvailableReplicas: 0,
		DeployStatus:      []appv1.DeployStatus{},
	}
	for _, dc := range ac.Spec.DeployConfigs {
		status := appv1.DeployStatus{}
		dm, ok := dmMap[dc.Name]
		if !ok {
			continue
		}
		status.AvailableReplicas = dm.Status.AvailableReplicas
		ac.Status.AvailableReplicas += dm.Status.AvailableReplicas
		status.Type = dc.Type
		statusAvailable, ok := getCondition(appsv1.DeploymentAvailable, dm.Status.Conditions)
		if ok {
			status.AvailableStatus = statusAvailable.Status
		}

		statusProgressing, ok := getCondition(appsv1.DeploymentProgressing, dm.Status.Conditions)
		if ok {
			status.ProgressingStatus = statusProgressing.Status
		}

		ac.Status.DeployStatus = append(ac.Status.DeployStatus, status)
	}

	if err := r.Status().Update(ctx, ac); err != nil {
		acLog.Error(err, "failed to update status")
		return ctrl.Result{}, err
	}

	// 加载全局模板
	tcPath := strings.Split(os.Getenv("TEMPLATE_CONFIG"), "/")
	if len(tcPath) == 2 {
		templateConfig.SetNamespace(tcPath[0])
		templateConfig.SetName(tcPath[1])
		err := r.Get(ctx, types.NamespacedName{Namespace: templateConfig.Namespace, Name: templateConfig.Name}, templateConfig)
		if err != nil {
			acLog.Error(err, "failed to get template configmap")
		}
	}

	// 创建或更新 AppConfigs 所属资源
	for _, dc := range ac.Spec.DeployConfigs {
		acLog.Info("updating deployConfig", "namespace", req.Namespace, "name", dc.Name)

		if util.GetAnnotation(ac, appv1.AnnotationStrictUpdate) == appv1.TureValue {
			// 开启严格发布模式后，canary 部署失败时，stable 不允许更新
			if dc.Type == appv1.StableDeploy {
				canaryStatus, ok := getDeployStatus(appv1.CanaryDeploy, ac.Status.DeployStatus)
				if !ok || canaryStatus.ProgressingStatus != corev1.ConditionTrue || canaryStatus.AvailableStatus != corev1.ConditionTrue {
					acLog.Info("canary deploy failed, skip update", "namespace", req.Namespace, "name", req.Name)
					continue
				}
			}

		}

		dm, ok := dmMap[dc.Name]
		if ok {
			if util.GetAnnotation(ac, appv1.AnnotationStrictUpdate) == appv1.TureValue {
				// 开启严格更新模式后 image replicas 都没有变化的情况下暂停更新
				appContainer, ok := getContainer(appContainerName, dm.Spec.Template.Spec.Containers)
				if ok {
					if appContainer.Image == dc.Image && *dm.Spec.Replicas == *dc.Replicas {
						acLog.Info("image replicas no changes, skip update", "namespace", req.Namespace, "name", req.Name)
						continue
					}
				}
			}
		} else {
			dm = &appsv1.Deployment{}
		}
		dm.SetNamespace(ac.Namespace)
		dm.SetName(dc.Name)

		acLog.Info("updating deployment", "namespace", req.Namespace, "name", dc.Name)
		res, err := controllerutil.CreateOrUpdate(ctx, r.Client, dm, r.setDeployment(dm, ac, &dc))
		if err != nil {
			acLog.Error(err, "failed to update deployment")
			return ctrl.Result{}, err
		}
		acLog.Info("deployment updated", "namespace", req.Namespace, "name", dc.Name, "result", res)

		if ac.Spec.Service.Enable {
			acLog.Info("service updated", "namespace", ac.Namespace, "name", dc.Name)
			svc := &corev1.Service{}
			svc.SetNamespace(ac.Namespace)
			svc.SetName(dc.Name)
			res, err := controllerutil.CreateOrUpdate(ctx, r.Client, svc, r.setSvc(svc, ac, &dc))
			if err != nil {
				acLog.Error(err, "failed to update service")
				return ctrl.Result{}, err
			}
			acLog.Info("service updated", "namespace", ac.Namespace, "name", dc.Name, "result", res)
		}

		if ac.Spec.Ingress.Enable {
			ingress := &networkingv1.Ingress{}
			ingress.SetNamespace(ac.Namespace)
			ingress.SetName(dc.Name)
			acLog.Info("updating ingress", "namespace", ac.Namespace, "name", ac.Name)
			res, err := controllerutil.CreateOrUpdate(ctx, r.Client, ingress, r.setIngress(ingress, ac, &dc))
			if err != nil {
				acLog.Error(err, "failed to update ingress")
				return ctrl.Result{}, err
			}
			acLog.Info("ingress updated", "namespace", ac.Namespace, "name", ac.Name, "result", res)
		}
	}
	return ctrl.Result{}, nil
}

func (r *AppConfigReconciler) setIngress(ingress *networkingv1.Ingress, ac *appv1.AppConfig, dc *appv1.DeployConfig) controllerutil.MutateFn {
	return func() error {
		if dc.Type == appv1.CanaryDeploy {
			canaryStatus, ok := getDeployStatus(appv1.CanaryDeploy, ac.Status.DeployStatus)
			if ok {
				weight := float32(canaryStatus.AvailableReplicas) / float32(ac.Status.AvailableReplicas) * 100
				ingress.Annotations = make(map[string]string)
				ingress.Annotations["nginx.ingress.kubernetes.io/canary"] = "true"
				ingress.Annotations["nginx.ingress.kubernetes.io/canary-weight"] = fmt.Sprint(int32(weight))
			}
		}
		ingress.Labels = make(map[string]string)
		addCreatedByLabel(ingress.Labels)
		pathType := new(networkingv1.PathType)
		*pathType = networkingv1.PathTypeImplementationSpecific
		ingress.Spec.Rules = []networkingv1.IngressRule{
			{
				Host: ac.Spec.Ingress.Host,
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: []networkingv1.HTTPIngressPath{
							{
								Path:     "/",
								PathType: pathType,
								Backend: networkingv1.IngressBackend{
									Service: &networkingv1.IngressServiceBackend{
										Name: dc.Name,
										Port: networkingv1.ServiceBackendPort{
											Number: ac.Spec.Service.Port,
										},
									},
								},
							},
						},
					},
				},
			},
		}

		return ctrl.SetControllerReference(ac, ingress, r.Scheme)
	}
}

func (r *AppConfigReconciler) setSvc(svc *corev1.Service, ac *appv1.AppConfig, dc *appv1.DeployConfig) controllerutil.MutateFn {
	return func() error {
		svc.Labels = make(map[string]string)
		addCreatedByLabel(svc.Labels)
		svc.Spec.Selector = make(map[string]string)
		svc.Spec.Selector["app"] = dc.Name
		svc.Spec.Ports = []corev1.ServicePort{
			{
				Name:       appContainerName,
				Port:       ac.Spec.Service.Port,
				TargetPort: intstr.FromInt32(ac.Spec.Service.Port),
			},
		}

		return ctrl.SetControllerReference(ac, svc, r.Scheme)
	}
}

func (r *AppConfigReconciler) setDeployment(dm *appsv1.Deployment, ac *appv1.AppConfig, dc *appv1.DeployConfig) controllerutil.MutateFn {
	return func() error {
		// 加载全局配置
		if dmTmpl, ok := templateConfig.Data["deployment"]; ok {
			acLog.Info("loading deployment template", "namespace", dm.Namespace, "name", dm.Name)
			if err := yaml.Unmarshal([]byte(dmTmpl), dm); err != nil {
				acLog.Error(err, "failed to unmarshal deployment template")
			}
		}

		// 标签设置
		dm.Labels = make(map[string]string)
		addCreatedByLabel(dm.Labels)
		dm.Labels[appContainerName] = dm.Name

		dm.Spec.Template.Labels = make(map[string]string)
		dm.Spec.Template.Labels[appContainerName] = dm.Name
		addCreatedByLabel(dm.Spec.Template.Labels)
		dm.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: dm.Spec.Template.Labels,
		}

		// 设置注解
		dm.Spec.Template.Annotations = make(map[string]string)

		for k, v := range ac.Annotations {
			if strings.HasPrefix(k, appv1.SidecarPrefix) {
				dm.Spec.Template.Annotations[k] = v
				dm.Spec.Template.Labels[appv1.SidecarPrefix] = "enable"
			}
			if k == appv1.AnnotationDeployment {
				acLog.Info("loading deployment annotation", "namespace", dm.Namespace, "name", dm.Name)
				err := json.Unmarshal([]byte(v), dm)
				if err != nil {
					acLog.Error(err, "failed to unmarshal deployment annotation")
				}
			}
		}

		// 设置容器
		dm.Spec.Replicas = dc.Replicas
		if _, ok := getContainer(appContainerName, dm.Spec.Template.Spec.Containers); !ok {
			dm.Spec.Template.Spec.Containers = append(dm.Spec.Template.Spec.Containers, corev1.Container{
				Name:  appContainerName,
				Image: dc.Image,
			})
		}
		setContainerImage(appContainerName, dc.Image, dm.Spec.Template.Spec.Containers)

		dm.ResourceVersion = ""
		dm.SetName(dc.Name)
		dm.SetNamespace(ac.Namespace)
		return ctrl.SetControllerReference(ac, dm, r.Scheme)
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *AppConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// 需要被 controller 管理的资源在这里注册

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &appsv1.Deployment{}, ownerKey, func(rawObj client.Object) []string {
		dm := rawObj.(*appsv1.Deployment)
		owner := metav1.GetControllerOf(dm)
		if owner == nil {
			return nil
		}
		if owner.APIVersion != apiGVStr || owner.Kind != apiKind {
			return nil
		}
		return []string{owner.Name}
	}); err != nil {
		return err
	}
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &corev1.Service{}, ownerKey, func(rawObj client.Object) []string {
		svc := rawObj.(*corev1.Service)
		owner := metav1.GetControllerOf(svc)
		if owner == nil {
			return nil
		}
		if owner.APIVersion != apiGVStr || owner.Kind != apiKind {
			return nil
		}
		return []string{owner.Name}
	}); err != nil {
		return err
	}
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &networkingv1.Ingress{}, ownerKey, func(rawObj client.Object) []string {
		ingress := rawObj.(*networkingv1.Ingress)
		owner := metav1.GetControllerOf(ingress)
		if owner == nil {
			return nil
		}
		if owner.APIVersion != apiGVStr || owner.Kind != apiKind {
			return nil
		}
		return []string{owner.Name}
	}); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1.AppConfig{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&networkingv1.Ingress{}).
		Complete(r)
}
