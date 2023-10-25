package controller

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/yaml"
	appv1 "sanmuyan.com/app-operator/api/v1"
	"sanmuyan.com/app-operator/pkg/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strings"
)

func (r *AppConfigReconciler) updateFinalizer(ctx context.Context, ac *appv1.AppConfig) error {
	if ac.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(ac, appv1.AppConfigFinalizer) {
			controllerutil.AddFinalizer(ac, appv1.AppConfigFinalizer)
			return r.Update(ctx, ac)
		}
	} else {
		if controllerutil.ContainsFinalizer(ac, appv1.AppConfigFinalizer) {
			controllerutil.RemoveFinalizer(ac, appv1.AppConfigFinalizer)
			return r.Update(ctx, ac)
		}
	}
	return nil
}

func (r *AppConfigReconciler) listDeployment(ctx context.Context, ac *appv1.AppConfig) (map[string]*appsv1.Deployment, error) {
	dmMap := make(map[string]*appsv1.Deployment)
	dmList := &appsv1.DeploymentList{}
	if err := r.List(ctx, dmList, client.InNamespace(ac.Namespace), client.MatchingFields{ownerKey: ac.Name}); err != nil {
		return dmMap, err
	}
	for _, dm := range dmList.Items {
		dmMap[dm.Name] = &dm
	}
	return dmMap, nil

}

func (r *AppConfigReconciler) updateStatus(ctx context.Context, ac *appv1.AppConfig, dmMap map[string]*appsv1.Deployment) error {
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

	return r.Status().Update(ctx, ac)
}

func (r *AppConfigReconciler) setTemplate(cm *corev1.ConfigMap) {
	if getNamePath(&cm.ObjectMeta) == templatePath {
		acLog.Info("set template config", "namespace", cm.Namespace, "name", cm.Name)
		templateCM = cm
	}
}

func (r *AppConfigReconciler) updateDeploy(ctx context.Context, req ctrl.Request, ac *appv1.AppConfig, dmMap map[string]*appsv1.Deployment) error {
	for _, dc := range ac.Spec.DeployConfigs {
		acLog.V(1).Info("updating deployConfig", "namespace", req.Namespace, "name", dc.Name)
		req.MarshalLog()
		if util.GetAnnotation(ac, appv1.StrictReleaseAnnotation) == appv1.TureValue {
			// 开启严格发布模式后，canary 部署失败时，stable 不允许更新
			if dc.Type == appv1.StableDeploy {
				canaryStatus, ok := getDeployStatus(appv1.CanaryDeploy, ac.Status.DeployStatus)
				if !ok || canaryStatus.ProgressingStatus != corev1.ConditionTrue || canaryStatus.AvailableStatus != corev1.ConditionTrue {
					acLog.V(1).Info("canary deploy failed, skip update", "namespace", req.Namespace, "name", req.Name)
					continue
				}
			}

		}

		dm, ok := dmMap[dc.Name]
		if ok {
			if util.GetAnnotation(ac, appv1.StrictUpdateAnnotation) == appv1.TureValue {
				// 开启严格更新模式后 image replicas 都没有变化的情况下暂停更新
				appContainer, ok := getContainer(appContainerName, dm.Spec.Template.Spec.Containers)
				if ok {
					if appContainer.Image == dc.Image && *dm.Spec.Replicas == *dc.Replicas {
						acLog.V(1).Info("image replicas no changes, skip update", "namespace", req.Namespace, "name", req.Name)
						continue
					}
				}
			}
		} else {
			dm = &appsv1.Deployment{}
		}
		dm.SetNamespace(ac.Namespace)
		dm.SetName(dc.Name)

		res, err := controllerutil.CreateOrUpdate(ctx, r.Client, dm, r.setDeployment(dm, ac, &dc))
		if err != nil {
			return err
		}
		acLog.V(1).Info("deployment updated", "namespace", req.Namespace, "name", dc.Name, "result", res)

		if ac.Spec.Service.Enable {
			svc := &corev1.Service{}
			svc.SetNamespace(ac.Namespace)
			svc.SetName(dc.Name)
			res, err := controllerutil.CreateOrUpdate(ctx, r.Client, svc, r.setSvc(svc, ac, &dc))
			if err != nil {
				return err
			}
			acLog.V(1).Info("service updated", "namespace", ac.Namespace, "name", dc.Name, "result", res)
		}

		if ac.Spec.Ingress.Enable {
			ingress := &networkingv1.Ingress{}
			ingress.SetNamespace(ac.Namespace)
			ingress.SetName(dc.Name)
			res, err := controllerutil.CreateOrUpdate(ctx, r.Client, ingress, r.setIngress(ingress, ac, &dc))
			if err != nil {
				return err
			}
			acLog.V(1).Info("ingress updated", "namespace", ac.Namespace, "name", ac.Name, "result", res)
		}
	}
	return nil
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
		svc.Spec.Selector[appContainerName] = dc.Name
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
		if dmTmpl, ok := templateCM.Data["deployment"]; ok {
			acLog.V(1).Info("loading deployment template", "namespace", dm.Namespace, "name", dm.Name)
			if err := yaml.Unmarshal([]byte(dmTmpl), dm); err != nil {
				acLog.Info("failed to unmarshal deployment template", "namespace", dm.Namespace, "name", dm.Name, "error", err)
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
			if k == appv1.DeploymentConfigAnnotation {
				acLog.V(1).Info("loading deployment annotation", "namespace", dm.Namespace, "name", dm.Name)
				err := json.Unmarshal([]byte(v), dm)
				if err != nil {
					return err
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
