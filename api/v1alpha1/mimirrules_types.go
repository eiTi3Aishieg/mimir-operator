package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MimirRulesSpec defines the desired state of MimirRules
type MimirRulesSpec struct {
	// ID is the identifier of the tenant in the Mimir Ruler
	ID string `json:"id"`

	// URL is the URL of the remote Mimir Ruler
	URL string `json:"url"`

	// Authentication configuration if it is required by the remote endpoint
	Auth *Auth `json:"auth,omitempty"`

	// Rules that should be linked to the tenant ID in the Mimir Ruler
	Rules *Rules `json:"rules"`
}

// Rules that are associated to a tenant and that should be synchronized to the Mimir Ruler
// The rules must be defined in CRDs of type "PrometheusRule" and this resource should
// only be used to target those PrometheusRules by referencing them through selectors
type Rules struct {
	Selectors *metav1.LabelSelector `json:"selectors,omitempty"`
}

// Auth contains configuration to set up authentication on the remote Mimir Ruler endpoint
// There are two supported authentication schemes:
//   - User/API key
//   - Token (JWT/bearer)
//
// Token has precedence over any other authentication method
// If the user/API key scheme is selected, the key can either be given directly or through
// a secretRef pointing to a Kubernetes Secret containing the API key under the field "key"
// The Token can also be given using a Secret containing the value under the field "token"
type Auth struct {
	User           string                   `json:"user,omitempty"`
	Key            string                   `json:"key,omitempty"`
	KeySecretRef   *v1.LocalObjectReference `json:"keySecretRef,omitempty"`
	Token          string                   `json:"token,omitempty"`
	TokenSecretRef *v1.LocalObjectReference `json:"tokenSecretRef,omitempty"`
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
