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

type ReleaseBackend string

const (
	TraefikBackend   ReleaseBackend = "traefik"
	NoReleaseBackend ReleaseBackend = ""
)

type StackAppConfigSpec struct {
	AppNamespace string            `json:"appNamespace"`
	Releases     ReleaseConfig     `json:"releases,omitempty"`
	StackValues  StackValuesConfig `json:"stackValues,omitempty"`
	AppRevisions AppRevisionConfig `json:"appRevisions,omitempty"`
}

type SigningConfig struct {
	Optional                 bool `json:"optional,omitempty"`
	InsecureSkipVerification bool `json:"insecureSkipVerification,omitempty"`
}

type AppRevisionConfig struct {
	AppNamespace string        `json:"appNamespace"`
	Signing      SigningConfig `json:"signing,omitempty"`
	DevMode      bool          `json:"devMode,omitempty"`
}

type StackValuesConfig struct {
	// Enabled is true when StackValues are enabled
	Enabled bool `json:"enabled"`

	// Secret is a reference to a secret containing auth data
	Secret  *corev1.SecretReference `json:"secret,omitempty"`
	Sources []*StackValueSource     `json:"sources,omitempty"`
}

func (c StackValuesConfig) Source(t StackValueSourceType) *StackValueSource {
	for _, src := range c.Sources {
		if t == src.Type {
			return src
		}
	}
	return nil
}

type StackValueSource struct {
	Name string `json:"name"`

	// +kubebuilder:validation:Enum=artifactory;vault;aws_s3
	Type  StackValueSourceType `json:"type"`
	Route string               `json:"route"`
	// TODO: this should be specified somewhere else
	Region string `json:"region"`

	Token []byte `json:"token,omitempty"`
}

type ReleaseConfig struct {
	Enabled          bool           `json:"enabled,omitempty"`
	ProxyNamespace   string         `json:"proxyNamespace,omitempty"`
	BackendType      ReleaseBackend `json:"backendType,omitempty"`
	IngressPort      int32          `json:"ingressPort,omitempty"`
	ReleaseStages    []ReleaseStage `json:"releaseStages,omitempty"`
	HostName         string         `json:"host,omitempty"`
	RollbackRevision StackAppSpec   `json:"rollbackRevision,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster

// StackAppConfig is the Schema for the stackappconfigs API
type StackAppConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              StackAppConfigSpec `json:"spec"`
}

// +kubebuilder:object:root=true

// StackAppConfigList contains a list of StackAppConfig
type StackAppConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StackAppConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&StackAppConfig{}, &StackAppConfigList{})
}
