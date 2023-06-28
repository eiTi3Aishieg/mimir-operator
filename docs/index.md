# Mimir Operator documentation

The Mimir Operator is a Kubernetes operator to control Mimir tenants using CRDs.

## Available CRDs

### MimirTenant

The MimirTenant CRD allows the remote control of a tenant in a Mimir instance from Kubernetes.  


The general structure of the CRD is as follows:

```yaml
apiVersion: mimir.grafana.com/v1alpha1
kind: MimirTenant
metadata:
  name: mimirtenant-sample
  namespace: default
spec:
  id: "tenant1" # ID of the tenant in Mimir (X-ScopeOrg-ID header)
  url: "http://mimir.instance.com" # URL of the Mimir instance, used by the operator to connect and operate on the tenants

  # The Rules section is used to install alerting rules on a tenant. 
  # The rules must be defined on the Kubernetes Cluster using "PrometheusRules" from the PrometheusOperator (https://github.com/prometheus-operator/prometheus-operator)
  # Any PrometheusRule that matches the selectors will be installed for the tenant in Mimir
  # Selectors can be of type "matchelLabels", "matchExpressions" or both (requirements are ANDed).
  # See the K8S documentation on selectors for more information (https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/).
  rules:
    selectors:
      matchLabels:
        helm.sh/chart: loki-4.10.1 # Install PrometheusRules from the Loki chart
```

### Installing Prometheus Rules for a Tenant

**PrometheusRules** are selected using selectors to determine what should be installed in the Mimir ruler for the tenant. Once all the rules have been filtered using the selectors, they are synced with the remote Mimir instance.

In the Mimir ruler, alerts are grouped in "groups", which are themselves grouped in "namespaces". The name of a namespace is computed by taking the Kubernetes namespace of the PrometheusRule that was used to generate the Mimir rule, and appending the name of the Kubernetes PrometheusRule.

In the following example, the **loki-alerts** PrometheusRule will be installed in the Mimir tenant under the namespace *alerts-loki-alerts", with one group named "loki_alerts".

This example shows how to install a PrometheusRule to monitor Loki:
```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  labels:
    alert-type: loki
    alert-level: "0"
  name: loki-alerts
  namespace: alerts
spec:
  groups:
  - name: loki_alerts
    rules:
    - alert: LokiRequestErrors
      annotations:
        message: |
          {{ $labels.job }} {{ $labels.route }} is experiencing {{ printf "%.2f" $value }}% errors.
      expr: |
        100 * sum(rate(loki_request_duration_seconds_count{status_code=~"5"}[2m])) by (namespace, job, route) /
        sum(rate(loki_request_duration_seconds_count[2m])) by (namespace, job, route) > 10
      for: 15m
      labels:
        severity: critical

---
apiVersion: mimir.grafana.com/v1alpha1
kind: MimirTenant
metadata:
  name: mimirtenant-sample
  namespace: default
spec:
  id: "loki-tenant"
  url: "http://mimir.instance.com"
  rules:
    selectors:
      matchLabels:
        alert-type: loki
        alert-level: "0"
```

To select all the rules available on the cluster:
```yaml
 rules:
    selectors: {}
```