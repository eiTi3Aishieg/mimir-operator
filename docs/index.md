# Mimir Operator documentation

The Mimir Operator is a Kubernetes operator to control Mimir tenants using CRDs.

## Installing

### Helm

The Helm Chart is published in the OCI format on GitHub.

```
helm install -i mimir-operator oci://ghcr.io/AmiditeX/helm-charts/mimir-operator --version v0.1.2
```

Helm is the easiest way to install the operator. The manifests for the Helm Chart can be found in ```deploy/helm/mimir-operator```.  

### Kustomize
You can optionally clone the repository and use the manifests in ```config/``` to deploy the operator using Kustomize.  
The easiest way to deploy using Kustomize is to simply run the ```make deploy``` command.

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

Token authentication (```tokenSecretRef``` OR ```token```) has precedence over any other authentication method (both schemes can't be used simultaneously).  
User/API key authentication (```keySecretRef``` OR ```key``` and ```user```) must provide a user AND a key.

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
  # See the K8S documentation on selectors for more information (https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/).
  rules:
    selectors:
      matchLabels:
        helm.sh/chart: loki-4.10.1 # Install PrometheusRules from the Loki chart
```

### Installing Prometheus Rules for a Tenant

**PrometheusRules** are selected using selectors to determine what should be installed in the Mimir Ruler for the tenant. Once all the rules have been filtered using the selectors, they are synced with the remote Mimir instance.

In the Mimir Ruler, alerts are grouped in "groups", which are themselves grouped in "namespaces". The name of a namespace is computed by taking the Kubernetes namespace of the PrometheusRule that was used to generate the Mimir rule, and appending the name of the Kubernetes PrometheusRule.

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
      matchLabels:
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
      matchLabels:
        version: v1
      matchExpressions:
      - key: group
        operator: In
        values:
          - kubernetes
          - node
          - watchdog
```
This would match any PrometheusRule with a label ```version=v1``` and a ```group``` label with any of the following values: ```[kubernetes, node, watchdog]```.  
See the official [Kubernetes documentation](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/) on labels and selectors for more examples.  