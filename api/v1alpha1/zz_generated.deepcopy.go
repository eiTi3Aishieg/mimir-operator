//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Auth) DeepCopyInto(out *Auth) {
	*out = *in
	if in.KeySecretRef != nil {
		in, out := &in.KeySecretRef, &out.KeySecretRef
		*out = new(corev1.LocalObjectReference)
		**out = **in
	}
	if in.TokenSecretRef != nil {
		in, out := &in.TokenSecretRef, &out.TokenSecretRef
		*out = new(corev1.LocalObjectReference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Auth.
func (in *Auth) DeepCopy() *Auth {
	if in == nil {
		return nil
	}
	out := new(Auth)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MimirRules) DeepCopyInto(out *MimirRules) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MimirRules.
func (in *MimirRules) DeepCopy() *MimirRules {
	if in == nil {
		return nil
	}
	out := new(MimirRules)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MimirRules) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MimirRulesList) DeepCopyInto(out *MimirRulesList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]MimirRules, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MimirRulesList.
func (in *MimirRulesList) DeepCopy() *MimirRulesList {
	if in == nil {
		return nil
	}
	out := new(MimirRulesList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MimirRulesList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MimirRulesSpec) DeepCopyInto(out *MimirRulesSpec) {
	*out = *in
	if in.Auth != nil {
		in, out := &in.Auth, &out.Auth
		*out = new(Auth)
		(*in).DeepCopyInto(*out)
	}
	if in.Rules != nil {
		in, out := &in.Rules, &out.Rules
		*out = new(Rules)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MimirRulesSpec.
func (in *MimirRulesSpec) DeepCopy() *MimirRulesSpec {
	if in == nil {
		return nil
	}
	out := new(MimirRulesSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MimirRulesStatus) DeepCopyInto(out *MimirRulesStatus) {
	*out = *in
	if in.RulesStatus != nil {
		in, out := &in.RulesStatus, &out.RulesStatus
		*out = new(RulesStatus)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MimirRulesStatus.
func (in *MimirRulesStatus) DeepCopy() *MimirRulesStatus {
	if in == nil {
		return nil
	}
	out := new(MimirRulesStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Rules) DeepCopyInto(out *Rules) {
	*out = *in
	if in.Selectors != nil {
		in, out := &in.Selectors, &out.Selectors
		*out = new(v1.LabelSelector)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Rules.
func (in *Rules) DeepCopy() *Rules {
	if in == nil {
		return nil
	}
	out := new(Rules)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RulesStatus) DeepCopyInto(out *RulesStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RulesStatus.
func (in *RulesStatus) DeepCopy() *RulesStatus {
	if in == nil {
		return nil
	}
	out := new(RulesStatus)
	in.DeepCopyInto(out)
	return out
}
