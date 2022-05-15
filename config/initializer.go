package config

import (
	"github.com/gin-gonic/gin"
	discovery "github.com/gkarthiks/k8s-discovery"
	"github.com/sirupsen/logrus"
)

// InitializeRouter initializing router
func InitializeRouter() {
	logrus.Infoln("initializing the router")
	Router = gin.New()
	Router.Use(Logger(logrus.New()), gin.Recovery())
}

// InitializeKubeClient initializing the kubeclient
func InitializeKubeClient() {
	logrus.Infoln("initializing the Kube client")
	KubeClient, _ = discovery.NewK8s()
	version, _ := KubeClient.GetVersion()
	logrus.Infoln("running %v version in the target cluster", version)
}
