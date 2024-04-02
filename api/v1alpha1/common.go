package v1alpha1

import v1 "k8s.io/api/core/v1"

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
