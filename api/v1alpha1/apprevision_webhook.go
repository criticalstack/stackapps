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
	apprevisionlog = logf.Log.WithName("apprevision-resource")

	defaultHealthCheck = HealthCheck{
		Type:  HealthCheckTypeJSONPath,
		Value: "{.status.resourceConditions}",
	}

	_ webhook.Defaulter = &AppRevision{}
	_ webhook.Validator = &AppRevision{}
)

func (r *AppRevision) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *AppRevision) Default() {
	apprevisionlog.Info("default", "name", r.Name)

	if r.Spec.Revision == 0 {
		r.Spec.Revision = 1
	}

	// we intentionally check for nil here, not an explicit empty slice
	if r.Spec.HealthChecks == nil {
		r.Spec.HealthChecks = append(r.Spec.HealthChecks, defaultHealthCheck)
	}
}

func (r *AppRevision) validate() error {
	if r.Spec.Revision == 0 {
		return errors.Errorf("invalid revision: 0")
	}
	if r.Spec.Manifests == "" {
		return errors.Errorf("manifests are required")
	}
	if len(r.Spec.Signatures) == 0 && !r.Spec.Config.Signing.Optional {
		return errors.Errorf("no signatures present and signing is required")
	}
	return nil
}

func (r *AppRevision) ValidateCreate() error {
	return r.validate()
}

func (r *AppRevision) ValidateUpdate(old runtime.Object) error {
	prev, _ := old.(*AppRevision)
	if prev.Spec.Manifests != r.Spec.Manifests && prev.Spec.Revision == r.Spec.Revision {
		return errors.Errorf("manifests changed without incrementing revision")
	}
	return r.validate()
}

func (r *AppRevision) ValidateDelete() error {
	return nil
}
