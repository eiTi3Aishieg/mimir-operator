// SPDX-License-Identifier: AGPL-3.0-only
// Provenance-includes-location: https://github.com/grafana/cortex-tools/blob/main/pkg/rules/rwrulefmt/rulefmt.go
// Provenance-includes-license: Apache-2.0
// Provenance-includes-copyright: The Cortex Authors.

package rwrulefmt

import (
	"github.com/AmiditeX/mimir-operator/internal/controller/mimirapi/rwrulefmt/model"

	"gopkg.in/yaml.v3"
)

// Wrapper around Prometheus rulefmt.

// RuleGroup is a list of sequentially evaluated recording and alerting rules.
type RuleGroup struct {
	FmtRuleGroup `yaml:",inline"`
	// RWConfigs is used by the remote write forwarding ruler
	RWConfigs []RemoteWriteConfig `yaml:"remote_write,omitempty"`
}

// RuleGroup is a list of sequentially evaluated recording and alerting rules.
type FmtRuleGroup struct {
	Name     string         `yaml:"name"`
	Interval model.Duration `yaml:"interval,omitempty"`
	Limit    int            `yaml:"limit,omitempty"`
	Rules    []RuleNode     `yaml:"rules"`
}

// Rule describes an alerting or recording rule.
type Rule struct {
	Record        string            `yaml:"record,omitempty"`
	Alert         string            `yaml:"alert,omitempty"`
	Expr          string            `yaml:"expr"`
	For           model.Duration    `yaml:"for,omitempty"`
	KeepFiringFor model.Duration    `yaml:"keep_firing_for,omitempty"`
	Labels        map[string]string `yaml:"labels,omitempty"`
	Annotations   map[string]string `yaml:"annotations,omitempty"`
}

// RuleNode adds yaml.v3 layer to support line and column outputs for invalid rules.
type RuleNode struct {
	Record        yaml.Node         `yaml:"record,omitempty"`
	Alert         yaml.Node         `yaml:"alert,omitempty"`
	Expr          yaml.Node         `yaml:"expr"`
	For           model.Duration    `yaml:"for,omitempty"`
	KeepFiringFor model.Duration    `yaml:"keep_firing_for,omitempty"`
	Labels        map[string]string `yaml:"labels,omitempty"`
	Annotations   map[string]string `yaml:"annotations,omitempty"`
}

// RemoteWriteConfig is used to specify a remote write endpoint
type RemoteWriteConfig struct {
	URL string `json:"url,omitempty"`
}
