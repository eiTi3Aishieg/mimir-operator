/*
Copyright 2024.

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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MimirRulesSpec defines the desired state of MimirRules
type MimirRulesSpec struct {
	// ID is the identifier of the tenant in the Mimir Ruler
	ID string `json:"id"`

	// URL is the URL of the remote Mimir Ruler
	URL string `json:"url"`

	// Authentication configuration if it is required by the remote endpoint
	Auth *Auth `json:"auth,omitempty"`

	// Rules that should be added to the tenant in the Mimir Ruler
	Rules *Rules `json:"rules"`

	// Overrides applied to specific rules in this tenant
	Overrides map[string]Override `json:"overrides,omitempty"`

	// ExternalLabels added to the alerts automatically when they are fired
	ExternalLabels map[string]string `json:"externalLabels,omitempty"`
}

// Rules that are associated to a tenant and that should be synchronized to the Mimir Ruler
// The rules must be defined in CRDs of type "PrometheusRule" and this resource should
// only be used to target those PrometheusRules by referencing them through selectors
type Rules struct {
	Selectors []*metav1.LabelSelector `json:"selectors"`
}

// Override is a structure containing parameters that can be overridden inside
// a PrometheusRule. This is useful to override certain alerts within certain
// alert groups with fine-tuned properties such as the query used to fire the alert.
// This structure can also be used to specify the rule should be outright disabled.
type Override struct {
	Disable     bool              `json:"disable,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Expr        string            `json:"expr,omitempty"`
	For         string            `json:"for,omitempty"`
}

// MimirRulesStatus defines the status of the synchronization of Rules associated with a MimirRules
type MimirRulesStatus struct {
	// Status describes whether the rules are synchronized
	Status string `json:"status,omitempty"`

	// Error describes the last synchronization error
	Error string `json:"error,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`

// MimirRules is the Schema for the MimirRules API
type MimirRules struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MimirRulesSpec   `json:"spec"`
	Status MimirRulesStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MimirRulesList contains a list of MimirRules
type MimirRulesList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MimirRules `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MimirRules{}, &MimirRulesList{})
}
