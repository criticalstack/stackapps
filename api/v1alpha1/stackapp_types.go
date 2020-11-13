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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:webhook:verbs=create;update;delete,path=/validate-features-criticalstack-com-v1alpha1-stackapp,mutating=false,failurePolicy=fail,groups=*,resources=*,versions=*,name=stackapp-validation.criticalstack.com

// StackAppSpec defines the desired state of StackApp
type StackAppSpec struct {
	AppRevision  AppRevisionSpec `json:"appRevision"`
	MajorVersion uint64          `json:"majorVersion"`
}

type StackAppState string

const (
	StackAppStateCreating StackAppState = "creating"
	StackAppStateError    StackAppState = "error"
	StackAppStateReady    StackAppState = "ready"
)

type StackAppStatus struct {
	CurrentRelease CurrentReleaseState `json:"currentRelease,omitempty"`

	State   StackAppState `json:"state"`
	Reason  string        `json:"reason"`
	Message string        `json:"message"`
}

type CurrentReleaseState struct {
	Name               string `json:"name"`
	Namespace          string `json:"namespace"`
	StackReleaseStatus `json:"status,omitempty"`
}

type StackAppConditionType string

const (
	StackAppReady = "Ready"
)

// +kubebuilder:resource:scope=Cluster
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="App Namespace",type="string",JSONPath=".status.currentRelease.status.currentRevision.namespace",description="App Namespace"
// +kubebuilder:printcolumn:name="Revision",type="integer",JSONPath=".status.currentRelease.status.currentRevision.revision",description="Revision"
// +kubebuilder:printcolumn:name="Manifests",type="string",JSONPath=".spec.appRevision.manifests",description="Current Manifest ConfigMap"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.currentRelease.status.state",description="Deployment Status"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// StackApp is the Schema for the stackapps API
type StackApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StackAppSpec   `json:"spec,omitempty"`
	Status StackAppStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// StackAppList contains a list of StackApp
type StackAppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StackApp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&StackApp{}, &StackAppList{})
}
