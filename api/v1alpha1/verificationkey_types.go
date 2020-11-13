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
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"sort"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:resource:scope=Cluster
// +kubebuilder:object:root=true

// VerificationKey is the Schema for the verificationkeys API
type VerificationKey struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Data              string `json:"data"`
}

// +kubebuilder:object:root=true

// VerificationKeyList contains a list of VerificationKey
type VerificationKeyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VerificationKey `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VerificationKey{}, &VerificationKeyList{})
}

func (vk *VerificationKey) PublicKey() (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(vk.Data))
	if block == nil {
		return nil, errors.Errorf("no PEM-encoded data found in verification key %q", vk.Name)
	}
	if block.Type != "RSA PUBLIC KEY" {
		return nil, errors.Errorf("expected PEM block with type %q, got %q in verification key %q", "RSA PUBLIC KEY", block.Type, vk.Name)
	}
	key, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		if _, ok := err.(asn1.StructuralError); ok {
			err = errors.Errorf("structural error in asn1-encoded data")
		}
		return nil, errors.Wrapf(err, "failed parsing verification key %q", vk.Name)
	}
	return key, nil
}

var VerificationHash = crypto.SHA256

func (vk *VerificationKey) VerifyConfigMapSignature(signature []byte, cm *corev1.ConfigMap) error {
	pk, err := vk.PublicKey()
	if err != nil {
		return err
	}
	if err := rsa.VerifyPKCS1v15(pk, VerificationHash, ConfigMapHash(cm), signature); err != nil {
		return errors.Wrapf(err, "signature verification using key %q failed", vk.Name)
	}
	return nil
}

func ConfigMapHash(cm *corev1.ConfigMap) []byte {
	keys := make([]string, 0)
	for k := range cm.Data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	h := VerificationHash.New()
	// add to hash in alphabetical order
	for _, k := range keys {
		h.Write([]byte(cm.Data[k]))
	}
	return h.Sum(nil)
}
