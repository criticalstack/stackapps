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
	"github.com/pkg/errors"
	runtime "k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var (
	verificationkeylog = logf.Log.WithName("verificationkey-resource")

	_ webhook.Validator = &VerificationKey{}
)

func (r *VerificationKey) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

func (vk *VerificationKey) validate() error {
	_, err := vk.PublicKey()
	return err
}

func (vk *VerificationKey) ValidateCreate() error {
	return vk.validate()
}

func (vk *VerificationKey) ValidateUpdate(old runtime.Object) error {
	prev, _ := old.(*VerificationKey)
	if prev.Data != vk.Data {
		return errors.Errorf("cannot change the key data of an existing key")
	}
	return vk.validate()
}

func (vk *VerificationKey) ValidateDelete() error {
	return nil
}
