/*
Copyright 2020 Critical Stack.

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
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// AppRevisionSpec defines the desired state of AppRevision
type AppRevisionSpec struct {
	// Revision number of this version of the application
	Revision uint64 `json:"revision,omitempty"`

	// Manifests represents the name of a ConfigMap in the app namespace containing manifests to be deployed
	Manifests string `json:"manifests"`

	// Signatures is an optional map of VerificationKey names to signatures of the manifest data
	Signatures map[string][]byte `json:"signatures,omitempty"`

	HealthChecks []HealthCheck `json:"healthChecks,omitempty"`

	Config AppRevisionConfig `json:"appRevisionConfig,omitempty"`
}

type HealthCheckType string

const (
	HealthCheckTypeJSONPath   HealthCheckType = "jsonpath"
	HealthCheckTypeGoTemplate HealthCheckType = "go-template"
)

type HealthCheck struct {
	// +kubebuilder:validation:Enum=jsonpath;go-template
	Type  HealthCheckType `json:"type"`
	Value string          `json:"value"`
	Name  string          `json:"name,omitempty"`
}

type AppRevisionState string

const (
	AppRevisionStateCreating AppRevisionState = "creating"
	AppRevisionStateUpdating AppRevisionState = "updating"
	AppRevisionStateError    AppRevisionState = "error"
	AppRevisionStateReady    AppRevisionState = "ready"
)

// AppRevisionStatus defines the observed state of AppRevision
type AppRevisionStatus struct {
	Resources          []AppRevisionResource  `json:"resources"`
	OriginalResources  []AppRevisionResource  `json:"originalResources"`
	Conditions         []AppRevisionCondition `json:"conditions"`
	ResourceConditions []ResourceCondition    `json:"resourceConditions"`
}

// Default sets non-nil values for the status
func (s *AppRevisionStatus) Default() {
	if s.Resources == nil {
		s.Resources = make([]AppRevisionResource, 0)
	}
	if s.OriginalResources == nil {
		s.OriginalResources = make([]AppRevisionResource, 0)
	}
	if s.Conditions == nil {
		s.Conditions = make([]AppRevisionCondition, 0)
	}
	if s.ResourceConditions == nil {
		s.ResourceConditions = make([]ResourceCondition, 0)
	}
}

// ResourceReference represents a local (to the namespace) resource as known at time of creation
type ResourceReference struct {
	APIVersion      string `json:"apiVersion"`
	Kind            string `json:"kind"`
	Name            string `json:"name"`
	UID             string `json:"uid,omitempty"`
	ResourceVersion string `json:"resourceVersion"`
}

type ResourceCondition struct {
	Type   ResourceConditionType  `json:"type"`
	Status corev1.ConditionStatus `json:"status"`

	Instances []ResourceConditionInstance `json:"instances"`
}

type ResourceConditionType string

type ResourceConditionInstance struct {
	Status   corev1.ConditionStatus           `json:"status"`
	Reason   string                           `json:"reason,omitempty"`
	Resource corev1.TypedLocalObjectReference `json:"resource"`
}

type AppRevisionResource struct {
	unstructured.Unstructured `json:"-"`
}

func (r AppRevisionResource) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.Object)
}

func (r *AppRevisionResource) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &r.Object)
}

type AppRevisionConditionType string

const (
	AppRevisionReady            AppRevisionConditionType = "Ready"
	AppRevisionHealthy          AppRevisionConditionType = "Healthy"
	AppRevisionDeploymentFailed AppRevisionConditionType = "DeploymentFailed"
	AppRevisionUpdating         AppRevisionConditionType = "Updating"
)

// AppRevisionCondition describes the state of an apprevision at a certain point.
type AppRevisionCondition struct {
	// Type of apprevision condition.
	Type AppRevisionConditionType `json:"type"`

	// Status of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`

	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`

	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`

	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Revision",type="integer",JSONPath=".spec.revision",description="Revision"
// +kubebuilder:printcolumn:name="Manifests",type="string",JSONPath=".spec.manifests",description="Current Manifest ConfigMap"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type == 'Ready')].status",description="AppRevision is ready"
// +kubebuilder:printcolumn:name="Healthy",type="string",JSONPath=".status.conditions[?(@.type == 'Healthy')].status",description="AppRevision is ready"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// AppRevision is the Schema for the stackapps API
type AppRevision struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AppRevisionSpec   `json:"spec,omitempty"`
	Status AppRevisionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// AppRevisionList contains a list of AppRevision
type AppRevisionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AppRevision `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AppRevision{}, &AppRevisionList{})
}
