package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AlertManagerConfigSpec defines the desired state of AlertManagerConfig
type AlertManagerConfigSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of AlertManagerConfig. Edit alertmanagerconfig_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// AlertManagerConfigStatus defines the observed state of AlertManagerConfig
type AlertManagerConfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AlertManagerConfig is the Schema for the alertmanagerconfigs API
type AlertManagerConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AlertManagerConfigSpec   `json:"spec,omitempty"`
	Status AlertManagerConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AlertManagerConfigList contains a list of AlertManagerConfig
type AlertManagerConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AlertManagerConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AlertManagerConfig{}, &AlertManagerConfigList{})
}
