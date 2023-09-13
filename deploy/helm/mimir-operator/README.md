# mimir-operator

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v0.1.6](https://img.shields.io/badge/AppVersion-v0.1.6-informational?style=flat-square)

Helm Chart to deploy the Mimir Operator

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` |  |
| fullnameOverride | string | `""` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.repository | string | `"ghcr.io/amiditex/mimir-operator"` |  |
| image.tag | string | `""` |  |
| imagePullSecrets | list | `[]` |  |
| leaderElect | bool | `false` | Enable leader election to run only multiple replicas of the operator, this is not recommended |
| metricsService.metricsPort | int | `9090` |  |
| metricsService.type | string | `"ClusterIP"` |  |
| nameOverride | string | `""` |  |
| nodeSelector | object | `{}` |  |
| podAnnotations | object | `{}` |  |
| podSecurityContext | object | `{}` |  |
| replicaCount | int | `1` |  |
| resources | object | `{}` |  |
| securityContext | object | `{}` |  |
| serviceAccount.annotations | object | `{}` |  |
| serviceAccount.create | bool | `true` |  |
| serviceAccount.name | string | `""` |  |
| tolerations | list | `[]` |  |

