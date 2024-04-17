// SPDX-License-Identifier: AGPL-3.0-only
// Provenance-includes-location: https://github.com/grafana/cortex-tools/blob/main/pkg/client/rules.go
// Provenance-includes-license: Apache-2.0
// Provenance-includes-copyright: The Cortex Authors.

package mimirapi

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"slices"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/AmiditeX/mimir-operator/internal/controller/mimirapi/rwrulefmt"
)

// ruleElement describes one element returned by Mimir when listing all the rules for a tenant
type RuleElement struct {
	Namespace string `json:"namespace"`
	RuleGroup string `json:"rulegroup"`
}

// CreateRuleGroup creates a new rule group
func (r *MimirClient) CreateRuleGroup(ctx context.Context, namespace string, rg rwrulefmt.RuleGroup) error {
	payload, err := yaml.Marshal(&rg)
	if err != nil {
		return err
	}

	escapedNamespace := url.PathEscape(namespace)
	path := r.apiPath + "/" + escapedNamespace

	res, err := r.doRequest(ctx, path, "POST", bytes.NewBuffer(payload), int64(len(payload)))
	if err != nil {
		return err
	}

	res.Body.Close()

	return nil
}

// CreateRuleGroup creates a new rule group
func (r *MimirClient) CreateRuleGroupStr(ctx context.Context, namespace, rg string) error {
	var rns rwrulefmt.RuleNamespace

	yaml.Unmarshal([]byte(rg), &rns)

	fmt.Printf("%+v", rns)

	for _, group := range rns.Groups {
		r.CreateRuleGroup(ctx, namespace, group)
	}

	return nil
}

// DeleteRuleGroup deletes a rule group
func (r *MimirClient) DeleteRuleGroup(ctx context.Context, namespace, groupName string) error {
	escapedNamespace := url.PathEscape(namespace)
	escapedGroupName := url.PathEscape(groupName)
	path := r.apiPath + "/" + escapedNamespace + "/" + escapedGroupName

	res, err := r.doRequest(ctx, path, "DELETE", nil, -1)
	if err != nil {
		return err
	}

	res.Body.Close()

	return nil
}

// GetRuleGroup retrieves a rule group
func (r *MimirClient) GetRuleGroup(ctx context.Context, namespace, groupName string) (*rwrulefmt.RuleGroup, error) {
	escapedNamespace := url.PathEscape(namespace)
	escapedGroupName := url.PathEscape(groupName)
	path := r.apiPath + "/" + escapedNamespace + "/" + escapedGroupName

	fmt.Println(path)
	res, err := r.doRequest(ctx, path, "GET", nil, -1)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	rg := rwrulefmt.RuleGroup{}
	err = yaml.Unmarshal(body, &rg)
	if err != nil {
		log.WithFields(log.Fields{
			"body": string(body),
		}).Debugln("failed to unmarshal rule group from response")

		return nil, errors.Wrap(err, "unable to unmarshal response")
	}

	return &rg, nil
}

// ListRules retrieves a rule group
func (r *MimirClient) ListRules(ctx context.Context, namespace string) (map[string][]rwrulefmt.RuleGroup, error) {
	path := r.apiPath
	if namespace != "" {
		path = path + "/" + namespace
	}

	res, err := r.doRequest(ctx, path, "GET", nil, -1)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	ruleSet := map[string][]rwrulefmt.RuleGroup{}
	err = yaml.Unmarshal(body, &ruleSet)
	if err != nil {
		return nil, err
	}

	return ruleSet, nil
}

// ListRulesElement retrieves a list of RuleGroup and Namespace (as RuleElement)
func (r *MimirClient) ListRulesElement(ctx context.Context, namespace string) ([]RuleElement, error) {
	ruleSet, err := r.ListRules(ctx, "")

	if err != nil {
		return nil, err
	}

	nsKeys := make([]string, 0, len(ruleSet))
	for k := range ruleSet {
		nsKeys = append(nsKeys, k)
	}
	slices.Sort(nsKeys)
	var rules []RuleElement

	for _, ns := range nsKeys {
		for _, rg := range ruleSet[ns] {
			rules = append(rules, RuleElement{
				Namespace: ns,
				RuleGroup: rg.Name,
			})
		}
	}
	return rules, nil
}

// DeleteNamespace delete all the rule groups in a namespace including the namespace itself
func (r *MimirClient) DeleteNamespace(ctx context.Context, namespace string) error {
	escapedNamespace := url.PathEscape(namespace)
	path := r.apiPath + "/" + escapedNamespace

	res, err := r.doRequest(ctx, path, "DELETE", nil, -1)
	if err != nil {
		return err
	}

	res.Body.Close()

	return nil
}
