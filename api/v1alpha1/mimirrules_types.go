package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MimirRulesSpec defines the desired state of MimirRules
type MimirRulesSpec struct {
	// ID is the identifier of the tenant in the Mimir Ruler
	ID string `json:"id"`

	// URL is the URL of the remote Mimir Ruler
	URL string `json:"url"`

	// Rules that should be linked to the tenant ID in the Mimir Ruler
	Rules *Rules `json:"rules"`
}

// Rules that are associated to a tenant and that should be synchronized to the Mimir Ruler
// The rules must be defined in CRDs of type "PrometheusRule" and this resource should
// only be used to target those PrometheusRules by referencing them through selectors
type Rules struct {
	Selectors *metav1.LabelSelector `json:"selectors,omitempty"`
}

// MimirRulesStatus defines the observed state of MimirRules
type MimirRulesStatus struct {
	// RulesStatus is the synchronization status of the rules linked to that tenant
	RulesStatus *RulesStatus `json:"rulesStatus,omitempty"`
}

// RulesStatus defines the status of the synchronization of Rules associated with a MimirRules
type RulesStatus struct {
	// Status describes whether the rules are synchronized
	Status string `json:"status,omitempty"`

	// Error describes the last synchronization error
	Error string `json:"error,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

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
