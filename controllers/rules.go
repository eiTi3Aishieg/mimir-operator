package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	prometheus "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	domain "mimir-operator/api/v1alpha1"
	"mimir-operator/internal/mimirtool"
	"mimir-operator/internal/utils"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// specFilter is used to deserialize YAML into it and filter out properties different from ".spec"
type specFilter struct {
	Spec spec `yaml:"spec"`
}

// spec holds only the ".groups" property that is of interest to us
type spec struct {
	Groups interface{} `yaml:"groups"`
}

// ruleElement describes one element returned by Mimir when listing all the rules for a tenant
type ruleElement struct {
	Namespace string `json:"namespace"`
	RuleGroup string `json:"rulegroup"`
}

// syncRulesToRuler finds all the PrometheusRules relevant for a MimirRules and sends them to Mimir
func (r *MimirRulesReconciler) syncRulesToRuler(ctx context.Context, auth *mimirtool.Authentication, mr *domain.MimirRules) error {
	rules, err := r.findPrometheusRulesFromLabels(ctx, mr.Spec.Rules.Selectors)
	if err != nil {
		return err
	}

	// Convert the PrometheusRules to a format Mimir understands
	unpackedRules, err := r.unpackRules(rules)
	if err != nil {
		return err
	}

	// Synchronize each Rule on the Mimir Ruler
	for ruleName, rule := range unpackedRules {
		if err := sendRuleToMimir(ctx, auth, mr.Spec.ID, mr.Spec.URL, ruleName, rule); err != nil {
			return err
		}
	}

	// Find the namespaces on Mimir that are NOT in our list of WANTED rules
	// Those namespaces might have been created earlier by the operator, but the MimirRules selectors
	// have changed since then, making those namespaces unwanted and in need of deletion.
	namespaces, err := diffRuleNamespaces(ctx, unpackedRules, auth, mr.Spec.ID, mr.Spec.URL)
	if err != nil {
		return err
	}

	// Synchronize each of those unwanted namespace with empty rule content to trigger a deletion
	for _, namespace := range namespaces {
		if err := sendRuleToMimir(ctx, auth, mr.Spec.ID, mr.Spec.URL, namespace, ""); err != nil {
			return err
		}
	}

	return nil
}

// deleteRulesForTenant deletes all the rules from Mimir for a specific tenant
func (r *MimirRulesReconciler) deleteRulesForTenant(ctx context.Context, auth *mimirtool.Authentication, mr *domain.MimirRules) error {
	// List all the rules and namespaces for the MimirRules in JSON format
	json, err := mimirtool.ListRules(ctx, auth, mr.Spec.ID, mr.Spec.URL)
	if err != nil {
		return err
	}

	// Convert the JSON to rules on which we can iterate easily
	rules, err := convertJsonToRules(json)
	if err != nil {
		return err
	}

	// For each rule, synchronize its namespace with empty rules to trigger a deletion
	for _, rule := range rules {
		if err := sendRuleToMimir(ctx, auth, mr.Spec.ID, mr.Spec.URL, rule.Namespace, ""); err != nil {
			return err
		}
	}

	return nil
}

// sendRuleToMimir synchronizes a PrometheusRule with the remote Mimir
func sendRuleToMimir(ctx context.Context, auth *mimirtool.Authentication, tenantId, url, ruleName, rule string) error {
	// Put the rule on the FS
	fileName, err := dumpRuleToFS(tenantId, ruleName, rule)
	if err != nil {
		return err
	}

	// Send the rule to Mimir for synchronization
	err = mimirtool.SynchronizeRules(ctx, auth, ruleName, fileName, tenantId, url)

	// Cleanup after ourselves
	if err := os.RemoveAll(fileName); err != nil {
		log.FromContext(ctx).
			WithValues("mimirrules", tenantId).
			Error(err, "failed to cleanup fs after sending rules to mimir")
	}

	return err
}

// findPrometheusRulesFromLabels lists all the CRs of type "PrometheusRules" based on label selectors
func (r *MimirRulesReconciler) findPrometheusRulesFromLabels(ctx context.Context, selector []*metav1.LabelSelector) (*prometheus.PrometheusRuleList, error) {
	prometheusRuleList := &prometheus.PrometheusRuleList{}

	for _, labelSelector := range selector {
		sel, err := metav1.LabelSelectorAsSelector(labelSelector)
		if err != nil {
			return nil, err
		}

		listOptions := client.ListOptions{
			LabelSelector: sel,
			Namespace:     "",
		}

		promRules := &prometheus.PrometheusRuleList{}
		if err := r.Client.List(ctx, promRules, &listOptions); err != nil {
			return nil, err
		}

		concatenatePrometheusRuleList(prometheusRuleList, promRules)
	}

	return prometheusRuleList, nil
}

// concatenatePrometheusRuleList concatenates every rule present in the src parameter into the dest and removes
// any possible duplicate in the process so that all the items added in dest are unique
func concatenatePrometheusRuleList(dest *prometheus.PrometheusRuleList, src *prometheus.PrometheusRuleList) {
	for _, promRule := range src.Items {
		if !isRuleInSlice(dest.Items, promRule) { // Remove duplicates
			dest.Items = append(dest.Items, promRule)
		}
	}
}

// isRuleInSlice returns true if a rule is contained in the slice passed in the first parameter
func isRuleInSlice(rules []*prometheus.PrometheusRule, rule *prometheus.PrometheusRule) bool {
	for _, r := range rules {
		if r.Namespace == rule.Namespace && r.Name == rule.Name {
			return true
		}
	}

	return false
}

// unpackRules reads a PrometheusRule CRD and keeps only the Groups embedded inside it
// The other fields are irrelevant to Mimir as the API only consumes files following
// the standard Prometheus Alerting Rules format
func (r *MimirRulesReconciler) unpackRules(list *prometheus.PrometheusRuleList) (map[string]string, error) {
	if list == nil {
		return nil, fmt.Errorf("no prometheus rules were passed")
	}

	codec := serializer.NewCodecFactory(r.Scheme).LegacyCodec(prometheus.SchemeGroupVersion)
	results := make(map[string]string)

	for _, rule := range list.Items {
		// Encode the Rule to JSON in the "kubectl" format to remove runtime fields
		output, err := runtime.Encode(codec, rule)
		if err != nil {
			return nil, err
		}

		var filter specFilter

		// Unmarshal into a structure containing only the ".spec" and ".spec.groups" properties to filter out everything else
		if err := json.Unmarshal(output, &filter); err != nil {
			return nil, err
		}

		// Re-marshal to keep only the ".groups" out of the ".spec"
		result, err := yaml.Marshal(&filter.Spec)
		if err != nil {
			return nil, err
		}

		results[rule.Namespace+"_"+rule.Name] = string(result)
	}

	return results, nil
}

// dumpRuleToFS writes a rule for a specific tenant into the filesystem
func dumpRuleToFS(tenant string, ruleName, rule string) (string, error) {
	path := temporaryFiles + tenant + "/"

	_ = os.Mkdir(path, os.ModePerm)

	fileName := path + ruleName
	return fileName, os.WriteFile(fileName, []byte(rule), 0644)
}

// diffRuleNamespaces returns Rule namespaces that are currently in Mimir for the tenant but not in the ruleMap
func diffRuleNamespaces(ctx context.Context, ruleMap map[string]string, auth *mimirtool.Authentication, tenant, url string) ([]string, error) {
	var namespaces []string

	// List all the rules and namespaces for the tenant in JSON format
	json, err := mimirtool.ListRules(ctx, auth, tenant, url)
	if err != nil {
		return nil, err
	}

	// Convert the JSON to rules on which we can iterate easily
	rules, err := convertJsonToRules(json)
	if err != nil {
		return nil, err
	}

	// For each namespace in Mimir, check if it's in the ruleMap
	for _, rule := range rules {
		_, ok := ruleMap[rule.Namespace]
		if !ok {
			namespaces = append(namespaces, rule.Namespace)
		}
	}

	// There might have been multiple rules in the same namespace, remove duplicates
	return utils.RemoveDuplicate(namespaces), nil
}

// convertJsonToRules converts JSON data listing the rules in Mimir for a tenant to a list of structures
func convertJsonToRules(data string) ([]ruleElement, error) {
	var elems []ruleElement

	if err := json.Unmarshal([]byte(data), &elems); err != nil {
		return nil, err
	}

	return elems, nil
}
