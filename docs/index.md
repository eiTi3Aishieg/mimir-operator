# Mimir Operator documentation

The Mimir Operator is a Kubernetes operator to control Mimir tenants using CRDs.

## Table of contents

- [Mimir Operator documentation](#mimir-operator-documentation)
  - [Installing](#installing)
    - [Helm](#helm)
    - [Kustomize](#kustomize)
  - [Authentication](#authentication)
  - [Available CRDs](#available-crds)
    - [MimirRules](#mimirrules)
      - [Installing Prometheus Rules for a Tenant](#installing-prometheus-rules-for-a-tenant)
      - [Overriding/disabling rules for a Tenant](#overriding-disabling-rules-for-a-tenant)
      - [Adding external labels](#adding-external-labels)
    - [MimirAlertManagerConfig](#mimiralertmanagerconfig)

## Installing

### Helm

The Helm Chart is published in the OCI format on GitHub.

```
helm install -i mimir-operator oci://ghcr.io/AmiditeX/helm-charts/mimir-operator --version v0.2.3
```

Helm is the easiest way to install the operator. The manifests for the Helm Chart can be found in `deploy/helm/mimir-operator`.

### Kustomize

You can optionally clone the repository and use the manifests in `config/` to deploy the operator using Kustomize.  
The easiest way to deploy using Kustomize is to simply run the `make deploy` command.

## Authentication

If the Mimir Endpoints are protected by authentication, the CRDs support an `auth` object allowing for various authentication methods:

- User/key
- Token (bearer/JWT)

The auth object has the following format:

```yaml
auth:
  tokenSecretRef: # Get the token from a secret in the namespace where the CR was deployed (secret key must be named "token")
    name: "secret-mimir"
  token: "token" # Plaintext token
  keySecretRef: # Get the key from a secret in the namespace where the CR was deployed (secret key must be named "key")
    name: "secret-mimir"
  key: "key" # Plaintext key
  user: "user" # Plaintext user
```

Token authentication (`tokenSecretRef` OR `token`) has precedence over any other authentication method (both schemes can't be used simultaneously).  
User/API key authentication (`keySecretRef` OR `key` and `user`) must provide a user AND a key.

## Available CRDs

### MimirRules

The MimirRules CRD allows the remote control of Rules for a specific tenant in a Mimir Ruler from Kubernetes.

The general structure of the CRD is as follows:

```yaml
apiVersion: mimir.randgen.xyz/v1alpha1
kind: MimirRules
metadata:
  name: mimirrules-sample
  namespace: default
spec:
  id: "tenant1" # ID of the tenant in the Mimir Ruler (X-ScopeOrg-ID header)
  url: "http://mimir.instance.com" # URL of the Mimir Ruler instance, used by the operator to connect and operate on the tenants

  # Authentication parameters if the endpoint is protected (see the Authentication section for more information)
  # auth:
  #   user: "user"
  #   key: "user-key"

  # The Rules section is used to install alerting rules on a tenant.
  # The rules must be defined on the Kubernetes Cluster using "PrometheusRules" from the PrometheusOperator (https://github.com/prometheus-operator/prometheus-operator)
  # Any PrometheusRule that matches the selectors will be installed for the tenant in Mimir
  # Selectors can be of type "matchelLabels", "matchExpressions" or both (requirements are ANDed).
  # Multiple selector blocks can be specified. The PrometheusRules filtered using each selector block are concatenated.
  # See the K8S documentation on selectors for more information (https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/).
  rules:
    selectors:
      - matchLabels:
          helm.sh/chart: loki-4.10.1 # Install PrometheusRules from the Loki chart
```

### Installing Prometheus Rules for a Tenant

**PrometheusRules** are selected using selectors to determine what should be installed in the Mimir Ruler for the tenant. Once all the rules have been filtered using the selectors, they are synced with the remote Mimir instance.

In the Mimir Ruler, alerts are grouped in "groups", which are themselves grouped in "namespaces". The name of a namespace is computed by taking the Kubernetes namespace of the PrometheusRule that was used to generate the Mimir rule, and appending the name of the Kubernetes PrometheusRule.

In the following example, the **loki-alerts** PrometheusRule will be installed in the Mimir tenant under the namespace _alerts-loki-alerts_, with one group named "loki_alerts".

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
apiVersion: mimir.randgen.xyz/v1alpha1
kind: MimirRules
metadata:
  name: mimirrules-sample
  namespace: default
spec:
  id: "loki-tenant"
  url: "http://mimir.instance.com"
  rules:
    selectors:
      - matchLabels:
          alert-type: loki
          alert-level: "0"
```

To select all the rules available on the cluster:

```yaml
rules:
  selectors: {}
```

More complex selections can be done by combining selectors:

```yaml
rules:
  selectors:
    - matchLabels:
        version: v1
      matchExpressions:
        - key: group
          operator: In
          values:
            - kubernetes
            - node
            - watchdog
    - matchLabels:
        version: v3
```

This would match any PrometheusRule with a label `version=v1` and a `group` label with any of the following values: `[kubernetes, node, watchdog]`
and any PrometheusRule with a label `version=v3`.
See the official [Kubernetes documentation](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/) on labels and selectors for more examples.

### Overriding/disabling rules for a Tenant

If you're actively monitoring a lot of tenants, you might make "rulebooks" using multiple PrometheusRules containing rules that should be applied to all your tenants.  
Those rules would constitute a _catalog_ of rules that you can apply (or not) to your tenants based on their nature (VMs, Kubernetes clusters, Network equipment...)  
This standardization of rules is good in practice to limit the unchecked growth of the amount of PrometheusRules deployed.  
But sometimes, it might occur that a specific tenant should have one of the rules contained inside those _rulesbooks_ overriden.
This may mean changing the query used to trigger the alert, change the "for" directive used to wait before firing, or change labels such as the severity.  
You can't change that rule in the common rulebook, or it would change it for every tenant on which that rule is applied.  
One of the solution would be to copy the PrometheusRule containing that rule, modify the rule and redeploy the PrometheusRule with different labelSelectors.  
This just doesn't scale with dozens of tenants monitored.
To make it easier, the **Mimir Operator** provides a way to override specific Rules for a tenant:

```yaml
apiVersion: mimir.randgen.xyz/v1alpha1
kind: MimirRules
metadata:
  name: mimirrules-sample
  namespace: default
spec:
  id: "tenant"
  url: "http://mimir.instance.com"
  rules:
    selectors:
      - matchLabels:
          version: v1
        matchExpressions:
          - key: group
            operator: In
            values:
              - kubernetes
              - node
              - watchdog
  overrides:
    NoMetricsFromTenant: # Name of the rule we wish to override
      disable: true # Disable the rule completely
      expr: "1" # Change the query evaluated to trigger the Alert
      labels:
        severity: info # Change labels, such as the severity, for this specific rule
        newLabel: "example"
      annotations:
        newAnnotation: "example value"
      for: "10m" # Change the "for" directive of the rule
```

The operator will only override properties that are specified. For example, if specifying an override for the "expr" property, but not the "labels" property, the rule will be deployed on Mimir with the overriden "expr" but will keep the labels inherited from the PrometheusRule.

### Adding external labels

It is possible to add labels to every rule installed in Mimir by a MimirRule using `externalLabels`  
For example, it can be useful to add a label indicating what tenant is emitting the alert.  
The syntax to add labels is the following:

```yaml
apiVersion: mimir.randgen.xyz/v1alpha1
kind: MimirRules
metadata:
  name: mimirrules-sample
  namespace: default
spec:
  id: "tenant"
  url: "http://mimir.instance.com"
  rules:
    selectors:
      - matchLabels:
          version: v1
        matchExpressions:
          - key: group
            operator: In
            values:
              - kubernetes
              - node
              - watchdog
  externalLabels:
    myLabel: myValue
    tenant: myTenant
```

The Rules installed in the Ruler by the operator will all have those labels appended to their list of labels.  
This effectively has the same effect as overriding each individual alert to add a new label.

### MimirAlertManagerConfig

The MimirAlertManagerConfig CRD allows the remote control of the Alertmanager config for a specific tenant in a Mimir instance from Kubernetes.

The general structure of the CRD is as follows:

```yaml
apiVersion: mimir.randgen.xyz/v1alpha1
kind: AlertManagerConfig
metadata:
  name: mimiralertmanagerconfig-sample
  namespace: default
spec:
  id: "tenant1" # ID of the tenant in the Mimir Ruler (X-ScopeOrg-ID header)
  url: "http://mimir.instance.com" # URL of the Mimir Ruler instance, used by the operator to connect and operate on the tenants

  # Authentication parameters if the endpoint is protected (see the Authentication section for more information)
  # auth:
  #   user: "user"
  #   key: "user-key"

  # The config section is used to declare the alert manager configuration on a tenant.
  config: |
    your alert manager configuration in yaml
```

### Installing a MimirAlertManagerConfig for a Tenant

In the following example, the **mimiralertmanagerconfig-sample** MimirAlertManagerConfig will be installed in the Mimir tenant under the namespace _default_.
Then the operator will load the configuration for the tenant **tenant1** in *http://mimir.instance.com*.

This example shows how to install a MimirAlertManagerConfig to configure a Teams webhook into the Alert Manager:

```yaml
apiVersion: mimir.randgen.xyz/v1alpha1
kind: AlertManagerConfig
metadata:
  name: mimiralertmanagerconfig-sample
  namespace: default
spec:
  id: "tenant1"
  url: "http://mimir.instance.com"
  config: |
    receivers:
      - name: teams
        msteams_configs:
          - webhook_url: "webhook_url"

    route:
      receiver: teams # Default is no child route is matched
      group_by: ["alertname", "severity"]
      group_interval: 1m
      repeat_interval: 1h
      routes: # Child routes
        - receiver: teams
          continue: true
```
