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
	runtime "k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var (
	stackapplog = logf.Log.WithName("stackapp-resource")

	_ webhook.Defaulter = &StackApp{}
	_ webhook.Validator = &StackApp{}
)

func (s *StackApp) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(s).
		Complete()
}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (s *StackApp) Default() {
	stackapplog.Info("default", "name", s.Name)

	ar := &AppRevision{Spec: s.Spec.AppRevision}
	ar.Name = s.Name
	ar.Default()
	s.Spec.AppRevision = ar.Spec
}

func (s *StackApp) validate() error {
	return nil
}

func (s *StackApp) ValidateCreate() error {
	return s.validate()
}

func (s *StackApp) ValidateUpdate(old runtime.Object) error {
	prev, _ := old.(*StackApp)
	if err := (&AppRevision{Spec: s.Spec.AppRevision}).ValidateUpdate(&AppRevision{Spec: prev.Spec.AppRevision}); err != nil {
		return err
	}
	return s.validate()
}

func (s *StackApp) ValidateDelete() error {
	return nil
}
