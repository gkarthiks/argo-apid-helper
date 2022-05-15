package config

import (
	argoAppV1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/gin-gonic/gin"
	discovery "github.com/gkarthiks/k8s-discovery"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sync"
)

var (
	AppMode         string
	ServerPort      string
	AppVersion      string
	ArgocdNamespace string
	Router          *gin.Engine
	KubeClient      *discovery.K8s

	LocalCluster = argoAppV1.Cluster{
		Name:            "in-cluster",
		Server:          argoAppV1.KubernetesInternalAPIServerAddr,
		ConnectionState: argoAppV1.ConnectionState{Status: argoAppV1.ConnectionStatusSuccessful},
	}
	InitLocalCluster           sync.Once
	ArgoManagedClusterSecrets  []v1.Secret
	ArgoManagedClusterNames    = sets.NewString()
	ArgoClusterNameToSecretMap = make(map[string]v1.Secret)
)

const (
	AppModeProd            = "production"
	DefaultArgoCDNamespace = "argocd"
	DefaultServerPort      = "8080"
	ClusterCollectorName   = "Cluster"
)

type DeprecationResults struct {
	ClusterName string      `json:"clusterName"`
	Result      interface{} `json:"result"`
}
