package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MimirTenantSpec defines the desired state of a MimirTenant
type MimirTenantSpec struct {
	// ID is the identifier of the MimirTenant
	ID string `json:"id"`

	// URL is the URL of the remote Mimir
	URL string `json:"url"`

	// Rules that are linked to that tenant
	Rules *Rules `json:"rules,omitempty"`
}

// Rules that are associated to a tenant and that should be synchronized to Mimir
// The rules must be defined in CRDs of type "PrometheusRule" and this resource should
// only be used to target those PrometheusRules by referencing them through selectors
type Rules struct {
	Selectors *metav1.LabelSelector `json:"selectors,omitempty"`
}

// MimirTenantStatus defines the observed state of MimirTenant
type MimirTenantStatus struct {
	// RulesStatus is the synchronization status of the rules linked to that tenant
	RulesStatus *RulesStatus `json:"rulesStatus,omitempty"`
}

// RulesStatus defines the status of the synchronization of Rules associated with a MimirTenant
type RulesStatus struct {
	// Status describes whether the rules are synchronized
	Status string `json:"status,omitempty"`

	// Error describes the last synchronization error
	Error string `json:"error,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MimirTenant is the Schema for the mimirtenants API
type MimirTenant struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MimirTenantSpec   `json:"spec"`
	Status MimirTenantStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MimirTenantList contains a list of MimirTenant
type MimirTenantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MimirTenant `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MimirTenant{}, &MimirTenantList{})
}
