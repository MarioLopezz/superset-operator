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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SupersetSpec defines the desired state of Superset
type SupersetSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Size defines the number of Superset instances
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=3
	// +kubebuilder:validation:ExclusiveMaximum=false
	MasterSize       int32          `json:"masterSize,omitempty"`
	WorkerSize       int32          `json:"workerSize,omitempty"`
	BaseImageVersion string         `json:"baseImageVersion,omitempty"`
	Deployment       DeploymentSpec `json:"deployment,omitempty"`
	Service          ServiceSpec    `json:"service"`
	Admin            AdminSpec      `json:"admin,omitempty"`
	DbConnection     DbConnection   `json:"dbConnection"`
}

// SupersetStatus defines the observed state of Superset
type SupersetStatus struct {
	// Represents the observations of a Superset's current state.
	// Superset.status.conditions.type are: "Available", "Progressing", and "Degraded"
	// Superset.status.conditions.status are one of True, False, Unknown.
	// Superset.status.conditions.reason the value should be a CamelCase string and producers of specific
	// condition types may define expected values and meanings for this field, and whether the values
	// are considered a guaranteed API.
	// Superset.status.conditions.Message is a human readable message indicating details about the transition.
	// For further information see: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

type DeploymentSpec struct {
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type ServiceSpec struct {
	UiNodePort int32 `json:"uiNodePort"`
}

type AdminSpec struct {
	User      string `json:"user,omitempty"`
	Pass      string `json:"pass,omitempty"`
	Firstname string `json:"firstname,omitempty"`
	Lastname  string `json:"lastname,omitempty"`
	Email     string `json:"email,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Superset is the Schema for the supersets API
type Superset struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SupersetSpec   `json:"spec,omitempty"`
	Status SupersetStatus `json:"status,omitempty"`
}

type DbConnection struct {
	RedisHost string `json:"redisHost"`
	RedisPort string `json:"redisPort"`
	DbHost    string `json:"dbHost"`
	DbPort    string `json:"dbPort"`
	DbUser    string `json:"dbUser"`
	DbPass    string `json:"dbPass"`
	DbName    string `json:"dbName"`
}

//+kubebuilder:object:root=true

// SupersetList contains a list of Superset
type SupersetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Superset `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Superset{}, &SupersetList{})
}
