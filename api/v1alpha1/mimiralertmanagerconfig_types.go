package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MimirAlertManagerConfigSpec defines the desired state of MimirAlertManagerConfig
type MimirAlertManagerConfigSpec struct {
	// ID is the identifier of the tenant in the Mimir Ruler
	ID string `json:"id"`

	// URL is the URL of the remote Mimir Ruler
	URL string `json:"url"`

	// Authentication configuration if it is required by the remote endpoint
	Auth *Auth `json:"auth,omitempty"`

	// Config that should be added to the tenant in the Mimir Alert Manager
	Config string `json:"config"`
}

// MimirAlertManagerConfigStatus defines the observed state of MimirAlertManagerConfig
type MimirAlertManagerConfigStatus struct {
	// Status describes whether the rules are synchronized
	Status string `json:"status,omitempty"`

	// Error describes the last synchronization error
	Error string `json:"error,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`

// MimirAlertManagerConfig is the Schema for the mimiralertmanagerconfigs API
type MimirAlertManagerConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MimirAlertManagerConfigSpec   `json:"spec,omitempty"`
	Status MimirAlertManagerConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MimirAlertManagerConfigList contains a list of MimirAlertManagerConfig
type MimirAlertManagerConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MimirAlertManagerConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MimirAlertManagerConfig{}, &MimirAlertManagerConfigList{})
}
