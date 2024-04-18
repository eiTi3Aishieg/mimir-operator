package utils

import (
	"context"
	"fmt"

	mimirrandgenxyzv1alpha1 "github.com/AmiditeX/mimir-operator/api/v1alpha1"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Authentication struct {
	Username string
	Key      string
	Token    string
}

// RemoveDuplicate removes duplicate values from a slice
func RemoveDuplicate[T string | int](sliceList []T) []T {
	allKeys := make(map[T]bool)
	list := []T{}
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

// FindSecretByRef returns a Kubernetes secret referenced using a secret name and namespace
func FindSecretByRef(ctx context.Context, c client.Client, secretName, secretNamespace string) (*v1.Secret, error) {
	secret := &v1.Secret{}

	objectKey := client.ObjectKey{
		Namespace: secretNamespace,
		Name:      secretName,
	}

	err := c.Get(ctx, objectKey, secret)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve secret %s/%s: %w", secretNamespace, secretName, err)
	}

	return secret, nil
}

// FindValueByKeyInSecret returns the value for a given key in a Secret
func FindValueByKeyInSecret(ctx context.Context, c client.Client, secretName, secretNamespace, key string) (string, error) {
	secret, err := FindSecretByRef(ctx, c, secretName, secretNamespace)
	if err != nil {
		return "", err
	}

	value, ok := secret.Data[key]
	if !ok {
		return "", fmt.Errorf("couldn't find  key '%s' in secret %s/%s", key, secretNamespace, secretName)
	}

	return string(value), nil
}

// ExtractAuth returns an internal authentication structure from a CRD authentication structure
// The returned authentication structure can be used by the package to generate authenticated command calls
// This function is safe to call with the 'auth' parameter set to 'nil' and will return a 'nil' auth structure and no error
// This is often needed if no authentication was provided by the user creating the CRD, as can be
// called without any authentication enabled, thus having no need for a mandatory authentication field in the CRDs
func ExtractAuth(ctx context.Context, client client.Client, auth *mimirrandgenxyzv1alpha1.Auth, namespace string) (*Authentication, error) {
	if auth == nil { // No authentication settings were provided
		return &Authentication{}, nil
	}

	if auth.Token != "" { // Token plaintext value has precedence over everything else
		return &Authentication{
			Token: auth.Token,
		}, nil
	}

	if auth.TokenSecretRef != nil { // Token secret reference has precedence over auth/key scheme
		token, err := FindValueByKeyInSecret(ctx, client, auth.TokenSecretRef.Name, namespace, "token")
		if err != nil {
			return nil, err
		}

		return &Authentication{
			Token: token,
		}, nil
	}

	if auth.Key != "" { // Plaintext key has precedence
		return &Authentication{
			Username: auth.User,
			Key:      auth.Key,
		}, nil
	}

	if auth.KeySecretRef != nil {
		key, err := FindValueByKeyInSecret(ctx, client, auth.KeySecretRef.Name, namespace, "key")
		if err != nil {
			return nil, err
		}

		return &Authentication{
			Username: auth.User,
			Key:      key,
		}, nil
	}

	return &Authentication{}, nil // Auth settings were provided but all uninitialized
}
