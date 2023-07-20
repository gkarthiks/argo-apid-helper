# API Deprecation Helmper(with ArgoCD)
[apid-helper](https://github.com/gkarthiks/argo-apid-helper) is a service that helps in finding the deprecated Kubernetes API and the workloads deployed in those deprecated APIs for the clusters that are managed by ArgoCD in an agentless way across all the clusters.


![Version: 0.1.3](https://img.shields.io/badge/Version-0.1.3-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v0.2.2](https://img.shields.io/badge/AppVersion-v0.2.2-informational?style=flat-square)

A Helm chart for Kubernetes API Deprecation Helper for Kubernetes that are managed by ArgoCD

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| autoscaling.enabled | bool | `false` | Enables the HPA |
| autoscaling.maxReplicas | int | `3` | Maximum instances to be deployed whne HPA is enabled and reached the throttling |
| autoscaling.minReplicas | int | `1` |  |
| autoscaling.targetCPUUtilizationPercentage | int | `80` | Target CPU throttling to for HPA |
| fullnameOverride | string | `""` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.repository | string | `"quay.io/gkarthics/apid-helper"` |  |
| image.tag | string | `"v0.2.0"` |  |
| imagePullSecrets | list | `[]` |  |
| ingress.enabled | bool | `false` |  |
| nameOverride | string | `""` |  |
| podAnnotations | object | `{}` |  |
| podSecurityContext | object | `{}` |  |
| replicaCount | int | `1` | Total number of instances to be deployed |
| resources.limits.cpu | string | `"100m"` | Maximum CPU to be reached before throttling |
| resources.limits.memory | string | `"128Mi"` | Maximum memory to be reached before throttling |
| resources.requests.cpu | string | `"100m"` | Minimum CPU that is requested for pod deployment |
| resources.requests.memory | string | `"128Mi"` | Minimum memory that is requested for pod deployment |
| server.appMode | string | `"debug"` | Prints all the logs from Debug level, alternate: `production` |
| server.argocdNamespace | string | `"argocd"` | Namespace where the ArgoCD infra is deployed and the cluster secrets are managed by ArgoCD Server |
| server.listenPort | int | `80` | Server HTTP listening port |
| server.test | bool | `false` | Deployes the test pod as part of the chart and does the health chcek as a one off job |
| service.port | int | `80` | Service HTTP port |
| service.type | string | `"ClusterIP"` | Type of the kubernetes service |
| serviceAccount.annotations | object | `{}` | Annotations for SA |
| serviceAccount.create | bool | `false` | When create is true, creates the service account, cluste role and bindings for accessing the ArgoCD managed secrets |
| serviceAccount.name | string | `""` | The service account name that needs to passed which should have access to the ArgoCD managed secrets for the ClusterSecrets |
