package config

import (
	argoAppV1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/gin-gonic/gin"
	discovery "github.com/gkarthiks/k8s-discovery"
	"github.com/sirupsen/logrus"
	"sync"
)

var (
	AppMode         string
	ServerPort      string
	AppVersion      string
	ArgocdNamespace string
	Router          *gin.Engine
	Log             *logrus.Logger
	K8s             *discovery.K8s

	LocalCluster = argoAppV1.Cluster{
		Name:            "in-cluster",
		Server:          argoAppV1.KubernetesInternalAPIServerAddr,
		ConnectionState: argoAppV1.ConnectionState{Status: argoAppV1.ConnectionStatusSuccessful},
	}
	InitLocalCluster sync.Once
)

const (
	AppModeProd            = "production"
	DefaultArgoCDNamespace = "argocd"
	DefaultServerPort      = "8080"
	ClusterCollectorName   = "Cluster"
)
