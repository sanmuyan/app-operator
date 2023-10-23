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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var appconfiglog = logf.Log.WithName("appconfig-resource")

func (r *AppConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-app-sanmuyan-com-v1-appconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=app.sanmuyan.com,resources=appconfigs,verbs=create;update,versions=v1,name=mappconfig.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &AppConfig{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *AppConfig) Default() {
	appconfiglog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-app-sanmuyan-com-v1-appconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=app.sanmuyan.com,resources=appconfigs,verbs=create;update,versions=v1,name=vappconfig.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &AppConfig{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *AppConfig) ValidateCreate() (admission.Warnings, error) {
	appconfiglog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *AppConfig) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	appconfiglog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *AppConfig) ValidateDelete() (admission.Warnings, error) {
	appconfiglog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
