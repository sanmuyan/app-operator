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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
//+kubebuilder:rbac:groups=*,resources=configmaps,verbs=*

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
		acLog.Info("failed to get appConfig", "namespace", req.Namespace, "name", req.Name, "error", err)
		return ctrl.Result{}, ignoreError(err)
	}

	// 添加 finalizer
	if err := r.updateFinalizer(ctx, ac); err != nil {
		acLog.Info("failed to update finalizer", "namespace", req.Namespace, "name", req.Name, "error", err)
		return ctrl.Result{}, ignoreError(err)
	}
	if !ac.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	// 判断是否处于暂停状态
	if ac.Spec.Paused {
		acLog.Info("appConfig is paused, skip update", "namespace", req.Namespace, "name", req.Name)
		return ctrl.Result{}, nil
	}

	// 获取所有的 Deployment
	dmMap, err := r.listDeployment(ctx, ac)
	if err != nil {
		acLog.Info("failed to list deployment", "namespace", req.Namespace, "name", req.Name, "error", err)
		return ctrl.Result{}, ignoreError(err)
	}

	// 更新 AppConfig 的状态
	if err := r.updateStatus(ctx, ac, dmMap); err != nil {
		acLog.Info("failed to update status", "namespace", req.Namespace, "name", req.Name, "error", err)
		return ctrl.Result{}, ignoreError(err)
	}

	// 创建或更新 AppConfig 所属资源、
	if err := r.updateDeploy(ctx, req, ac, dmMap); err != nil {
		acLog.Info("failed to update deploy", "namespace", req.Namespace, "name", req.Name, "error", err)
		return ctrl.Result{}, ignoreError(err)
	}
	return ctrl.Result{}, nil
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
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &corev1.ConfigMap{}, ownerKey, func(rawObj client.Object) []string {
		r.setTemplate(rawObj.(*corev1.ConfigMap))
		return nil
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
