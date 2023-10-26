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

package v1

import (
	apierr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var acLog = logf.Log.WithName("appconfig-resource")

func (r *AppConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-app-sanmuyan-com-v1-appconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=app.sanmuyan.com,resources=appconfigs,verbs=create;update,versions=v1,name=mappconfig.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &AppConfig{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *AppConfig) Default() {
	acLog.Info("default", "name", r.Name)

	for i := range r.Spec.DeployConfigs {
		r.Spec.DeployConfigs[i].Name = r.Name + "-" + string(r.Spec.DeployConfigs[i].Type)
	}
}

//+kubebuilder:webhook:path=/validate-app-sanmuyan-com-v1-appconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=app.sanmuyan.com,resources=appconfigs,verbs=create;update,versions=v1,name=vappconfig.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &AppConfig{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *AppConfig) ValidateCreate() (admission.Warnings, error) {
	acLog.Info("validate create", "name", r.Name)

	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *AppConfig) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	acLog.Info("validate update", "name", r.Name)

	var errList field.ErrorList
	if !controllerutil.ContainsFinalizer(r, AppConfigFinalizer) {
		if GetAnnotation(r, ProtectedAnnotation) == TureValue {
			errList = append(errList, field.Invalid(field.NewPath("annotations"), ProtectedAnnotation, DeleteProtectedMessage))
			return nil, apierr.NewInvalid(
				schema.GroupKind{Group: "app.sanmuyan.com", Kind: "AppConfig"}, r.Name, errList)
		}
	}

	for _, dc := range r.Spec.DeployConfigs {
		if dc.Type != StableDeploy && dc.Type != CanaryDeploy {
			errList = append(errList, field.Invalid(field.NewPath("spec", "deployConfigs", "type"), dc.Type, "invalid type"))
			return nil, apierr.NewInvalid(
				schema.GroupKind{Group: "app.sanmuyan.com", Kind: "AppConfig"}, r.Name, errList)
		}
	}
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *AppConfig) ValidateDelete() (admission.Warnings, error) {
	acLog.Info("validate delete", "name", r.Name)

	return nil, nil
}
