package mimirrules

import (
	"context"
	"encoding/json"
	"fmt"

	domain "github.com/AmiditeX/mimir-operator/api/v1alpha1"
	"github.com/AmiditeX/mimir-operator/internal/controller/mimirapi"
	"github.com/AmiditeX/mimir-operator/internal/utils"

	prometheus "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// specFilter is used to deserialize YAML into it and filter out properties different from ".spec"
type specFilter struct {
	Spec spec `yaml:"spec"`
}

// spec holds only the ".groups" property that is of interest to us
type spec struct {
	Groups interface{} `yaml:"groups"`
}

// syncRulesToRuler finds all the PrometheusRules relevant for a MimirRules and sends them to Mimir
func (r *MimirRulesReconciler) syncRulesToRuler(ctx context.Context, mc *mimirapi.MimirClient, mr *domain.MimirRules) error {
	rules, err := r.findPrometheusRulesFromLabels(ctx, mr.Spec.Rules.Selectors)
	if err != nil {
		return err
	}

	// Apply overrides on the PrometheusRules using the properties defined inside the MimirRules
	applyOverrides(mr.Spec.Overrides, rules)

	// Add external labels to the PrometheusRules
	applyExternalLabels(mr.Spec.ExternalLabels, rules)

	// Convert the PrometheusRules to a format Mimir understands
	unpackedRules, err := r.unpackRules(rules)
	if err != nil {
		return err
	}

	// Synchronize each Rule on the Mimir Ruler
	for namespace, ruleGroup := range unpackedRules {
		if err := mc.CreateRuleGroupStr(ctx, namespace, ruleGroup); err != nil {
			return err
		}
	}

	// Find the namespaces on Mimir that are NOT in our list of WANTED rules
	// Those namespaces might have been created earlier by the operator, but the MimirRules selectors
	// have changed since then, making those namespaces unwanted and in need of deletion.
	namespaces, err := r.diffRuleNamespaces(ctx, mc, unpackedRules, mr.Spec.ID)
	if err != nil {
		return err
	}

	// Synchronize each of those unwanted namespace with empty rule content to trigger a deletion
	for _, namespace := range namespaces {
		if err := mc.DeleteNamespace(ctx, namespace); err != nil {
			return err
		}
	}

	return nil
}

// deleteRulesForTenant deletes all the rules from Mimir for a specific tenant
func (r *MimirRulesReconciler) deleteRulesForTenant(ctx context.Context, mr *domain.MimirRules, mc *mimirapi.MimirClient) error {
	// List all the rules and namespaces for the MimirRules
	rules, err := mc.ListRulesElement(ctx, mr.Spec.ID)
	if err != nil {
		return err
	}

	// For each rule, synchronize its namespace with empty rules to trigger a deletion
	for _, rule := range rules {
		if err := mc.DeleteNamespace(ctx, rule.Namespace); err != nil {
			return err
		}
	}

	return nil
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

// removeRule removes one rule from a list of Prometheus Rules
func removeRule(s []prometheus.Rule, index int) []prometheus.Rule {
	s[index] = s[len(s)-1]
	return s[:len(s)-1]
}

// removeGroup removes one group from a list of Prometheus RuleGroups
func removeGroup(s []prometheus.RuleGroup, index int) []prometheus.RuleGroup {
	s[index] = s[len(s)-1]
	return s[:len(s)-1]
}

// applyOverrides applies overrides defined in a MimirRule to the properties of rules contained
// inside a list of PrometheusRules. This allows for fine-tuning of imported bulks such as
// rules from a catalog. Rules can be overridden to change a field such as the alert query or
// the amount of time necessary for the query to be true before it fires an alert.
// This is especially useful to set custom alerting conditions on particular tenants when
// they behave differently from most other tenants.
func applyOverrides(overrides map[string]domain.Override, list *prometheus.PrometheusRuleList) {
	if len(overrides) == 0 {
		return
	}

	for _, item := range list.Items {
		for g, group := range item.Spec.Groups {
			for r, rule := range group.Rules {
				// The content of a "rule" in PrometheusRules can be either:
				// - an Alert rule that triggers on specific conditions
				// - a Recording rule that records metrics to be analyzed by Rules
				// The type of rule is based on which of the two "Alert" and "Record" in non-null
				var ruleName string
				if rule.Alert != "" {
					ruleName = rule.Alert
				} else {
					ruleName = rule.Record
				}

				// Search if there's an override available for the name of that rule
				override, ok := overrides[ruleName]
				if !ok {
					continue
				}

				// Override specifies this rule should not exist, remove it entirely from the list
				if override.Disable {
					item.Spec.Groups[g].Rules = removeRule(item.Spec.Groups[g].Rules, r)

					// We may have deleted the last/only Rule inside the RuleGroup, if that is the case, the group
					// is now completely empty, which is invalid to the eyes of Mimir, so we just remove it.
					if len(item.Spec.Groups[g].Rules) == 0 {
						item.Spec.Groups = removeGroup(item.Spec.Groups, g)
					}

					continue
				}

				// Apply the override for any of the fields in the Rule if we have any specified
				if override.Annotations != nil {
					rule.Annotations = override.Annotations
				}

				if override.Labels != nil {
					rule.Labels = override.Labels
				}

				if override.Expr != "" {
					rule.Expr.StrVal = override.Expr
				}

				if override.For != "" {
					d := prometheus.Duration(override.For)
					rule.For = &d
				}

				item.Spec.Groups[g].Rules[r] = rule // We modified a copy of the rule, put it back in the *Rule
			}
		}
	}
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

// diffRuleNamespaces returns Rule namespaces that are currently in Mimir for the tenant but not in the ruleMap
func (r *MimirRulesReconciler) diffRuleNamespaces(ctx context.Context, mc *mimirapi.MimirClient, ruleMap map[string]string, tenant string) ([]string, error) {
	var namespaces []string

	// List all the rules and namespaces for the tenant
	rules, err := mc.ListRulesElement(ctx, tenant)
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

// applyExternalLabels adds a list of labels to every PrometheusRule in a list
func applyExternalLabels(labels map[string]string, list *prometheus.PrometheusRuleList) {
	if len(labels) == 0 {
		return
	}

	for _, item := range list.Items {
		for g, group := range item.Spec.Groups {
			for r, rule := range group.Rules {
				// No labels on the rule, create the map, so we can insert ours
				if rule.Labels == nil {
					rule.Labels = map[string]string{}
				}

				// Insert our label
				for key, value := range labels {
					rule.Labels[key] = value
				}

				item.Spec.Groups[g].Rules[r] = rule // We modified a copy of the rule, put it back in the *Rule }
			}
		}
	}
}
