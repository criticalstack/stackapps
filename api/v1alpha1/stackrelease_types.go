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

type StackReleaseSpec struct {
	AppName         string          `json:"appname"`
	Version         uint64          `json:"version"`
	AppRevision     AppRevisionSpec `json:"apprevision"`
	Config          ReleaseConfig   `json:"releaseconfig"`
	RollBackService string          `json:"rollBackService,omitempty"`
}

type ReleaseStage struct {
	CanaryWeight uint8           `json:"canaryWeight"`
	NextStep     *metav1.Time    `json:"nextStep,omitempty"`
	StepDuration metav1.Duration `json:"stepDuration"`
}

const (
	StackReleaseStateCreating  StackReleaseState = "creating"
	StackReleaseStateError     StackReleaseState = "error"
	StackReleaseStateDeploying StackReleaseState = "deploying"
	StackReleaseStateReady     StackReleaseState = "ready"
	StackReleaseStateRollback  StackReleaseState = "rollback"
)

type StackReleaseState string

type StackReleaseConditionType string

const (
	StackReleaseError      StackReleaseConditionType = "error"
	StackReleaseInstalling StackReleaseConditionType = "installing"
	StackReleaseDeploying  StackReleaseConditionType = "deploying"
)

type StackReleaseCondition struct {
	// Type of statefulset condition.
	Type StackReleaseConditionType `json:"type"`

	// Status of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`

	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`

	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`

	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`

	CanaryWeight ReleaseStage `json:"canaryweight"`
}

// StackReleaseStatus defines the observed state of StackRelease
type StackReleaseStatus struct {
	CurrentRevision CurrentRevisionState `json:"currentRevision"`

	State               StackReleaseState       `json:"state"`
	Reason              string                  `json:"reason"`
	Conditions          []StackReleaseCondition `json:"conditions,omitempty"`
	CurrentCanaryWeight ReleaseStage            `json:"currentCanaryWeight,omitempty"`
}

type CurrentRevisionState struct {
	Name      string                 `json:"name"`
	Namespace string                 `json:"namespace"`
	Revision  uint64                 `json:"revision"`
	Healthy   corev1.ConditionStatus `json:"healthy"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Revision",type="integer",JSONPath=".status.currentRevision.revision",description="Revision"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state",description="Deployment Status"
// +kubebuilder:printcolumn:name="Canary Weight",type="integer",JSONPath=".status.currentCanaryWeight.canaryWeight",description="Current Canary Weight"
// +kubebuilder:printcolumn:name="Next Step",type="string",format="date",JSONPath=".status.currentCanaryWeight.nextStep"
// +kubebuilder:printcolumn:name="Healthy",type="string",JSONPath=".status.currentRevision.healthy",description="AppRevision is ready"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// StackRelease is the Schema for the stackreleases API
// +kubebuilder:resource:scope=Cluster
type StackRelease struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              StackReleaseSpec   `json:"spec,omitempty"`
	Status            StackReleaseStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// StackReleaseList contains a list of StackRelease
type StackReleaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StackRelease `json:"items"`
}

func init() {
	SchemeBuilder.Register(&StackRelease{}, &StackReleaseList{})
}
