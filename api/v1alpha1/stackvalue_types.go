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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO(ktravis): have this use the deployment-style "template" for the underlying object

// StackValueSpec defines the desired state of StackValue
type StackValueSpec struct {
	AppName string `json:"appName"`

	// ObjectType is the kind of resource to be created
	ObjectType string `json:"objectType"`

	// +kubebuilder:validation:Enum=artifactory;aws_s3;vault
	SourceType StackValueSourceType `json:"sourceType"`
	Path       string               `json:"path"`
}

type StackValueSourceType string

const (
	StackValueSourceArtifactory StackValueSourceType = "artifactory"
	StackValueSourceAWSS3       StackValueSourceType = "aws_s3" // aws.s3 maybe?
	StackValueSourceVault       StackValueSourceType = "vault"
)

// TODO(ktravis): "synchronized" state, or is underlying object modified?

// StackValueStatus defines the observed state of StackValue
type StackValueStatus struct {
	Conditions []StackValueCondition `json:"conditions,omitempty"`
}

type StackValueConditionType string

const (
	StackValueReady  StackValueConditionType = "Ready"
	StackValueFailed StackValueConditionType = "Failed"
)

// AppRevisionCondition describes the state of a stackvalue at a certain point.
type StackValueCondition struct {
	// Type of stackvalue condition.
	Type StackValueConditionType `json:"type"`

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
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.objectType",description="Kind of object to be managed"
// +kubebuilder:printcolumn:name="Source",type="string",JSONPath=".spec.sourceType",description="Source type"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type == 'Ready')].status",description="StackValue is ready"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// StackValue is the Schema for the stackvalues API
type StackValue struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StackValueSpec   `json:"spec,omitempty"`
	Status StackValueStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// StackValueList contains a list of StackValues
type StackValueList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StackValue `json:"items"`
}

func init() {
	SchemeBuilder.Register(&StackValue{}, &StackValueList{})
}

type Values []StackValue
