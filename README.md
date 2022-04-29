# Argo APId Helper

*API Deprecation Helper* aims to provide an agentless way of listing all the deprecated APIs in the Kubernetes Cluster tha ismanaged by ArgoCD.

This helper service utilizes the *Kubernetes Secrets* created by ArgoCD to connect to the clusters. By which it gains the same privilege to read all the APIs and the workloads that are deployed on the associated deprecated APIs in that cluster. Although using the same privileges, it only reads from the cluster.

## Getting Started

For the helper to access the clusters properly, make sure the helper has access to the argo-cd cluster secrets. These secrets are created in the ArgoCD namespace. When deploying this `helper service` provide the argo-cd namespace in the environment variable `ARGOCD_NAMESPACE`.

The server will be started on `:8080` unless configured otherwise. There are a few environment variables that can be configured as tabulated below.

| S.No | Env Variable | Default Value | Desc |
|--|--|--|--|
| 01| APP_MODE | `production` | When set in `debug` mode, provides the verbosity|
| 02 | LISTEN_PORT | `8080` | Default server startup port |
|03|  ARGOCD_NAMESPACE | `argocd` | ArgoCD Namespace where the service can access the cluster-secrets|

### Available APIs
Once deployed, the service exposes the following apis that can be used to query the details.

#### /v1/ping
Responds with the `pong` message and used for bare minimal health check in containers.

#### /v1alpha/clusters
Will utilize the ArgoCD cluster-secrets and list the name and address of the clusters that are managed by ArgoCD; which in-turn are accessible by this helper service

#### /v1alpha/{cluster-name}/deprecations
This is to be implemented feature.

#### /v1alpha/deprecations
Responds back with the array of clusters, its corresponding deprecation api and workloads that are deployed against that corresponding apis.

Note: This might be a time-consuming task especially if your ArgoCD manages numerous clusters.