// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AppRevision) DeepCopyInto(out *AppRevision) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AppRevision.
func (in *AppRevision) DeepCopy() *AppRevision {
	if in == nil {
		return nil
	}
	out := new(AppRevision)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AppRevision) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AppRevisionCondition) DeepCopyInto(out *AppRevisionCondition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AppRevisionCondition.
func (in *AppRevisionCondition) DeepCopy() *AppRevisionCondition {
	if in == nil {
		return nil
	}
	out := new(AppRevisionCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AppRevisionConfig) DeepCopyInto(out *AppRevisionConfig) {
	*out = *in
	out.Signing = in.Signing
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AppRevisionConfig.
func (in *AppRevisionConfig) DeepCopy() *AppRevisionConfig {
	if in == nil {
		return nil
	}
	out := new(AppRevisionConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AppRevisionList) DeepCopyInto(out *AppRevisionList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AppRevision, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AppRevisionList.
func (in *AppRevisionList) DeepCopy() *AppRevisionList {
	if in == nil {
		return nil
	}
	out := new(AppRevisionList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AppRevisionList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AppRevisionResource) DeepCopyInto(out *AppRevisionResource) {
	*out = *in
	in.Unstructured.DeepCopyInto(&out.Unstructured)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AppRevisionResource.
func (in *AppRevisionResource) DeepCopy() *AppRevisionResource {
	if in == nil {
		return nil
	}
	out := new(AppRevisionResource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AppRevisionSpec) DeepCopyInto(out *AppRevisionSpec) {
	*out = *in
	if in.Signatures != nil {
		in, out := &in.Signatures, &out.Signatures
		*out = make(map[string][]byte, len(*in))
		for key, val := range *in {
			var outVal []byte
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = make([]byte, len(*in))
				copy(*out, *in)
			}
			(*out)[key] = outVal
		}
	}
	if in.HealthChecks != nil {
		in, out := &in.HealthChecks, &out.HealthChecks
		*out = make([]HealthCheck, len(*in))
		copy(*out, *in)
	}
	out.Config = in.Config
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AppRevisionSpec.
func (in *AppRevisionSpec) DeepCopy() *AppRevisionSpec {
	if in == nil {
		return nil
	}
	out := new(AppRevisionSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AppRevisionStatus) DeepCopyInto(out *AppRevisionStatus) {
	*out = *in
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = make([]AppRevisionResource, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.OriginalResources != nil {
		in, out := &in.OriginalResources, &out.OriginalResources
		*out = make([]AppRevisionResource, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]AppRevisionCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.ResourceConditions != nil {
		in, out := &in.ResourceConditions, &out.ResourceConditions
		*out = make([]ResourceCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AppRevisionStatus.
func (in *AppRevisionStatus) DeepCopy() *AppRevisionStatus {
	if in == nil {
		return nil
	}
	out := new(AppRevisionStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CurrentReleaseState) DeepCopyInto(out *CurrentReleaseState) {
	*out = *in
	in.StackReleaseStatus.DeepCopyInto(&out.StackReleaseStatus)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CurrentReleaseState.
func (in *CurrentReleaseState) DeepCopy() *CurrentReleaseState {
	if in == nil {
		return nil
	}
	out := new(CurrentReleaseState)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CurrentRevisionState) DeepCopyInto(out *CurrentRevisionState) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CurrentRevisionState.
func (in *CurrentRevisionState) DeepCopy() *CurrentRevisionState {
	if in == nil {
		return nil
	}
	out := new(CurrentRevisionState)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HealthCheck) DeepCopyInto(out *HealthCheck) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HealthCheck.
func (in *HealthCheck) DeepCopy() *HealthCheck {
	if in == nil {
		return nil
	}
	out := new(HealthCheck)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReleaseConfig) DeepCopyInto(out *ReleaseConfig) {
	*out = *in
	if in.ReleaseStages != nil {
		in, out := &in.ReleaseStages, &out.ReleaseStages
		*out = make([]ReleaseStage, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.RollbackRevision.DeepCopyInto(&out.RollbackRevision)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReleaseConfig.
func (in *ReleaseConfig) DeepCopy() *ReleaseConfig {
	if in == nil {
		return nil
	}
	out := new(ReleaseConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReleaseStage) DeepCopyInto(out *ReleaseStage) {
	*out = *in
	if in.NextStep != nil {
		in, out := &in.NextStep, &out.NextStep
		*out = (*in).DeepCopy()
	}
	out.StepDuration = in.StepDuration
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReleaseStage.
func (in *ReleaseStage) DeepCopy() *ReleaseStage {
	if in == nil {
		return nil
	}
	out := new(ReleaseStage)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceCondition) DeepCopyInto(out *ResourceCondition) {
	*out = *in
	if in.Instances != nil {
		in, out := &in.Instances, &out.Instances
		*out = make([]ResourceConditionInstance, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceCondition.
func (in *ResourceCondition) DeepCopy() *ResourceCondition {
	if in == nil {
		return nil
	}
	out := new(ResourceCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceConditionInstance) DeepCopyInto(out *ResourceConditionInstance) {
	*out = *in
	in.Resource.DeepCopyInto(&out.Resource)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceConditionInstance.
func (in *ResourceConditionInstance) DeepCopy() *ResourceConditionInstance {
	if in == nil {
		return nil
	}
	out := new(ResourceConditionInstance)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceReference) DeepCopyInto(out *ResourceReference) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceReference.
func (in *ResourceReference) DeepCopy() *ResourceReference {
	if in == nil {
		return nil
	}
	out := new(ResourceReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SigningConfig) DeepCopyInto(out *SigningConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SigningConfig.
func (in *SigningConfig) DeepCopy() *SigningConfig {
	if in == nil {
		return nil
	}
	out := new(SigningConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StackApp) DeepCopyInto(out *StackApp) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StackApp.
func (in *StackApp) DeepCopy() *StackApp {
	if in == nil {
		return nil
	}
	out := new(StackApp)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *StackApp) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StackAppConfig) DeepCopyInto(out *StackAppConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StackAppConfig.
func (in *StackAppConfig) DeepCopy() *StackAppConfig {
	if in == nil {
		return nil
	}
	out := new(StackAppConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *StackAppConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StackAppConfigList) DeepCopyInto(out *StackAppConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]StackAppConfig, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StackAppConfigList.
func (in *StackAppConfigList) DeepCopy() *StackAppConfigList {
	if in == nil {
		return nil
	}
	out := new(StackAppConfigList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *StackAppConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StackAppConfigSpec) DeepCopyInto(out *StackAppConfigSpec) {
	*out = *in
	in.Releases.DeepCopyInto(&out.Releases)
	in.StackValues.DeepCopyInto(&out.StackValues)
	out.AppRevisions = in.AppRevisions
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StackAppConfigSpec.
func (in *StackAppConfigSpec) DeepCopy() *StackAppConfigSpec {
	if in == nil {
		return nil
	}
	out := new(StackAppConfigSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StackAppList) DeepCopyInto(out *StackAppList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]StackApp, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StackAppList.
func (in *StackAppList) DeepCopy() *StackAppList {
	if in == nil {
		return nil
	}
	out := new(StackAppList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *StackAppList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StackAppSpec) DeepCopyInto(out *StackAppSpec) {
	*out = *in
	in.AppRevision.DeepCopyInto(&out.AppRevision)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StackAppSpec.
func (in *StackAppSpec) DeepCopy() *StackAppSpec {
	if in == nil {
		return nil
	}
	out := new(StackAppSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StackAppStatus) DeepCopyInto(out *StackAppStatus) {
	*out = *in
	in.CurrentRelease.DeepCopyInto(&out.CurrentRelease)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StackAppStatus.
func (in *StackAppStatus) DeepCopy() *StackAppStatus {
	if in == nil {
		return nil
	}
	out := new(StackAppStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StackRelease) DeepCopyInto(out *StackRelease) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StackRelease.
func (in *StackRelease) DeepCopy() *StackRelease {
	if in == nil {
		return nil
	}
	out := new(StackRelease)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *StackRelease) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StackReleaseCondition) DeepCopyInto(out *StackReleaseCondition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
	in.CanaryWeight.DeepCopyInto(&out.CanaryWeight)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StackReleaseCondition.
func (in *StackReleaseCondition) DeepCopy() *StackReleaseCondition {
	if in == nil {
		return nil
	}
	out := new(StackReleaseCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StackReleaseList) DeepCopyInto(out *StackReleaseList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]StackRelease, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StackReleaseList.
func (in *StackReleaseList) DeepCopy() *StackReleaseList {
	if in == nil {
		return nil
	}
	out := new(StackReleaseList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *StackReleaseList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StackReleaseSpec) DeepCopyInto(out *StackReleaseSpec) {
	*out = *in
	in.AppRevision.DeepCopyInto(&out.AppRevision)
	in.Config.DeepCopyInto(&out.Config)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StackReleaseSpec.
func (in *StackReleaseSpec) DeepCopy() *StackReleaseSpec {
	if in == nil {
		return nil
	}
	out := new(StackReleaseSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StackReleaseStatus) DeepCopyInto(out *StackReleaseStatus) {
	*out = *in
	out.CurrentRevision = in.CurrentRevision
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]StackReleaseCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.CurrentCanaryWeight.DeepCopyInto(&out.CurrentCanaryWeight)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StackReleaseStatus.
func (in *StackReleaseStatus) DeepCopy() *StackReleaseStatus {
	if in == nil {
		return nil
	}
	out := new(StackReleaseStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StackValue) DeepCopyInto(out *StackValue) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StackValue.
func (in *StackValue) DeepCopy() *StackValue {
	if in == nil {
		return nil
	}
	out := new(StackValue)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *StackValue) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StackValueCondition) DeepCopyInto(out *StackValueCondition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StackValueCondition.
func (in *StackValueCondition) DeepCopy() *StackValueCondition {
	if in == nil {
		return nil
	}
	out := new(StackValueCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StackValueList) DeepCopyInto(out *StackValueList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]StackValue, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StackValueList.
func (in *StackValueList) DeepCopy() *StackValueList {
	if in == nil {
		return nil
	}
	out := new(StackValueList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *StackValueList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StackValueSource) DeepCopyInto(out *StackValueSource) {
	*out = *in
	if in.Token != nil {
		in, out := &in.Token, &out.Token
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StackValueSource.
func (in *StackValueSource) DeepCopy() *StackValueSource {
	if in == nil {
		return nil
	}
	out := new(StackValueSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StackValueSpec) DeepCopyInto(out *StackValueSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StackValueSpec.
func (in *StackValueSpec) DeepCopy() *StackValueSpec {
	if in == nil {
		return nil
	}
	out := new(StackValueSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StackValueStatus) DeepCopyInto(out *StackValueStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]StackValueCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StackValueStatus.
func (in *StackValueStatus) DeepCopy() *StackValueStatus {
	if in == nil {
		return nil
	}
	out := new(StackValueStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StackValuesConfig) DeepCopyInto(out *StackValuesConfig) {
	*out = *in
	if in.Secret != nil {
		in, out := &in.Secret, &out.Secret
		*out = new(v1.SecretReference)
		**out = **in
	}
	if in.Sources != nil {
		in, out := &in.Sources, &out.Sources
		*out = make([]*StackValueSource, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(StackValueSource)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StackValuesConfig.
func (in *StackValuesConfig) DeepCopy() *StackValuesConfig {
	if in == nil {
		return nil
	}
	out := new(StackValuesConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in Values) DeepCopyInto(out *Values) {
	{
		in := &in
		*out = make(Values, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Values.
func (in Values) DeepCopy() Values {
	if in == nil {
		return nil
	}
	out := new(Values)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VerificationKey) DeepCopyInto(out *VerificationKey) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VerificationKey.
func (in *VerificationKey) DeepCopy() *VerificationKey {
	if in == nil {
		return nil
	}
	out := new(VerificationKey)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VerificationKey) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VerificationKeyList) DeepCopyInto(out *VerificationKeyList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]VerificationKey, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VerificationKeyList.
func (in *VerificationKeyList) DeepCopy() *VerificationKeyList {
	if in == nil {
		return nil
	}
	out := new(VerificationKeyList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VerificationKeyList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}
