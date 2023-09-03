# Argo APId Helper

[![Docker Repository on Quay](https://quay.io/repository/gkarthics/apid-helper/status "Docker Repository on Quay")](https://quay.io/repository/gkarthics/apid-helper)
![Release](https://img.shields.io/github/tag-date/gkarthiks/argo-apid-helper.svg?color=Orange&label=Latest%20Release)
![language](https://img.shields.io/badge/Language-go-blue.svg)
![License](https://img.shields.io/github/license/gkarthiks/argo-apid-helper.svg)


*API Deprecation Helper* aims to provide an agent less way of listing all the deprecated APIs in the Kubernetes Cluster that is managed by ArgoCD.

This helper service utilizes the *Kubernetes Secrets* created by ArgoCD to connect to the clusters. By which it gains the same privilege to read all the APIs and the workloads that are deployed on the associated deprecated APIs in that cluster. Although using the same privileges, it only reads from the cluster.

## Getting Started

For the helper to access the clusters properly, make sure the helper has access to the argo-cd cluster secrets. These secrets are created in the ArgoCD namespace. When deploying this `helper service` provide the argo-cd namespace in the environment variable `ARGOCD_NAMESPACE`.

The server will be started on `:8080` unless configured otherwise. There are a few environment variables that can be configured as tabulated below.

| S.No | Env Variable | Default Value | Desc |
|--|--|--|--|
| 01| APP_MODE | `production` | When set in `debug` mode, provides the verbosity|
| 02 | LISTEN_PORT | `80` | Default server startup port |
|03|  ARGOCD_NAMESPACE | `argocd` | ArgoCD Namespace where the service can access the cluster-secrets|

### Available APIs
Once deployed, the service exposes the following apis that can be used to query the details.

#### /v1/ping
Responds with the `pong` message and used for bare minimal health check in containers.

#### /v1alpha/clusters
Will utilize the ArgoCD cluster-secrets and list the name and address of the clusters that are managed by ArgoCD; which in-turn are accessible by this helper service

#### /v1alpha/{cluster-name}/deprecations
The `/v1alpha/{cluster-name}/deprecations` is a targeted cluster query to get the list of deprecated APIs and the resources deployed against those deprecated APIs on the provided cluster. 

This validates if the given cluster is managed by ArgoCD and starts the analysis. This API is very much recommended querying a large number of clusters. Since using the `/v1alpha/deprecations` api will takes longer time which might result in request time out error in some cases.

Also, the repetitive query on this api is guaranteed not to query the ArgoCD secrets for every request until the asked cluster name is not found in-memory.

#### /v1alpha/deprecations
Responds back with the array of clusters, its corresponding deprecation api and workloads that are deployed against that corresponding apis.

Note: This might be a time-consuming task especially if your ArgoCD manages numerous clusters.

### Deployment

This service is available as a container image for easy deployment at quay [here](https://quay.io/repository/gkarthics/apid-helper).

The helm chart for this deployment is available in ArtifactHUB, follow the simple steps by clicking [ArtifactHUB ⎈](https://artifacthub.io/packages/helm/gkarthiks/apid-helper?modal=install).