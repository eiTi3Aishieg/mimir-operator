package alertmanagerconfig

import (
	"context"
	domain "mimir-operator/api/v1alpha1"
	"mimir-operator/internal/mimirtool"
	"os"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

// deleteAlertManagerConfigForTenant deletes the alert manager configuration from Mimir for a specific tenant
func (r *AlertManagerConfigReconciler) deleteAlertManagerConfigForTenant(ctx context.Context, auth *mimirtool.Authentication, mr *domain.AlertManagerConfig) error {
	// Delete the configuration
	err := mimirtool.DeleteAlertManagerConfig(ctx, auth, mr.Spec.ID, mr.Spec.URL)
	if err != nil {
		return err
	}
	return nil
}

// unpackRules reads a PrometheusRule CRD and keeps only the Groups embedded inside it
// The other fields are irrelevant to Mimir as the API only consumes files following
// the standard Prometheus Alerting Rules format
// func (r *AlertManagerConfigReconciler) unpackRules(config *domain.AlertManagerConfig) (map[string]string, error) {

// 	codec := serializer.NewCodecFactory(r.Scheme).
// 	results := make(map[string]string)

// 	// Encode the Rule to JSON in the "kubectl" format to remove runtime fields
// 	output, err := runtime.Encode(codec, rule)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var filter specFilter

// 	// Unmarshal into a structure containing only the ".spec" and ".spec.groups" properties to filter out everything else
// 	if err := json.Unmarshal(output, &filter); err != nil {
// 		return nil, err
// 	}

// 	// Re-marshal to keep only the ".groups" out of the ".spec"
// 	result, err := yaml.Marshal(&filter.Spec)
// 	if err != nil {
// 		return nil, err
// 	}

// 	results[rule.Namespace+"_"+rule.Name] = string(result)

// 	return results, nil
// }

// sendAMConfigToMimir load an alert manager config with the remote Mimir
func sendAMConfigToMimir(ctx context.Context, auth *mimirtool.Authentication, tenantId, url, config string) error {

	// Put the rule on the FS
	configName := "amc_" + tenantId
	fileName, err := dumpConfigToFS(tenantId, configName, config)
	if err != nil {
		return err
	}

	// Verify alert manager configuration before loading it
	err = mimirtool.VerifyAlertManagerConfig(ctx, auth, fileName)

	if err != nil {
		log.FromContext(ctx).
			WithValues("alertmanagerconfig", tenantId).
			Error(err, "failed to validate configuration")

	} else {
		err = mimirtool.LoadAlertManagerConfig(ctx, auth, fileName, tenantId, url)
	}

	// Cleanup after ourselves
	if err := os.RemoveAll(fileName); err != nil {
		log.FromContext(ctx).
			WithValues("alertmanagerconfig", tenantId).
			Error(err, "failed to cleanup fs after loading alert manager configuration to mimir")
	}

	return err
}

// dumpRuleToFS writes a rule for a specific tenant into the filesystem
func dumpConfigToFS(tenant string, configName, rule string) (string, error) {
	path := temporaryFiles + tenant + "/"

	_ = os.Mkdir(path, os.ModePerm)

	fileName := path + configName
	return fileName, os.WriteFile(fileName, []byte(rule), 0644)
}
