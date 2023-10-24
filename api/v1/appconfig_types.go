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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type AppService struct {
	Enable bool  `json:"enable"`
	Port   int32 `json:"port"`
}

type DeployConfig struct {
	Name     string     `json:"name"`
	Image    string     `json:"image"`
	Replicas *int32     `json:"replicas"`
	Type     DeployType `json:"type"`
}

type AppIngress struct {
	Enable bool   `json:"enable"`
	Host   string `json:"host"`
}

// AppConfigSpec defines the desired state of AppConfig
type AppConfigSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of AppConfig. Edit appconfig_types.go to remove/update
	Ingress       AppIngress     `json:"ingress,omitempty"`
	Service       AppService     `json:"service,omitempty"`
	DeployConfigs []DeployConfig `json:"deployConfigs"`
	Paused        bool           `json:"paused,omitempty"`
}

type DeployStatus struct {
	AvailableStatus   corev1.ConditionStatus `json:"availableStatus"`
	ProgressingStatus corev1.ConditionStatus `json:"progressingStatus"`
	AvailableReplicas int32                  `json:"availableReplicas"`
	Type              DeployType             `json:"type"`
}

// AppConfigStatus defines the observed state of AppConfig
type AppConfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	DeployStatus      []DeployStatus `json:"deployStatus"`
	AvailableReplicas int32          `json:"availableReplicas"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AppConfig is the Schema for the appconfigs API
type AppConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AppConfigSpec   `json:"spec,omitempty"`
	Status AppConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AppConfigList contains a list of AppConfig
type AppConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AppConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AppConfig{}, &AppConfigList{})
}
