package main

import (
	"context"
	"github.com/gkarthiks/argo-apid-helper/config"
	"github.com/gkarthiks/argo-apid-helper/handlers"
	"github.com/sirupsen/logrus"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func init() {
	config.InitializeEnvVar()
	config.InitializeLogger()
	config.InitializeRouter()
	config.InitializeKubeClient()
	handlers.PopulateArgoClusterNames(context.Background())
}

func main() {

	// v1 api group
	v1 := config.Router.Group("/v1")
	v1.GET("/ping", handlers.HealthZ)

	v1alpha := config.Router.Group("/v1alpha")
	v1alpha.GET("/clusters", handlers.GetArgoClusters)

	v1alpha.GET("/deprecations", handlers.ListAPIDeprecations)
	v1alpha.GET("/:clusterName/deprecations", handlers.GetTargetClusterDeprecations)

	server := &http.Server{
		Addr:    ":" + config.ServerPort,
		Handler: config.Router,
	}

	logrus.Infof("configuring the apid server on %s port", config.ServerPort)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("listen: %s", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logrus.Fatalf("Server Shutdown: %s", err)
	}
	logrus.Info("Server exiting ...")
}
