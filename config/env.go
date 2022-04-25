package config

import (
	"github.com/sirupsen/logrus"
	"os"
)

func InitializeEnvVar() {
	appMode, avail := os.LookupEnv("APP_MODE")
	if !avail {
		logrus.Warn("defaulting app mode to production; results in Info log only")
		AppMode = AppModeProd
	} else {
		AppMode = appMode
	}

	serverPort, avail := os.LookupEnv("LISTEN_PORT")
	if !avail {
		logrus.Warn("LISTEN_PORT is not provided, defaulting to 8080 port")
		ServerPort = DefaultServerPort
	} else {
		ServerPort = serverPort
	}

	appVersion, avail := os.LookupEnv("APP_VERSION")
	if !avail {
		logrus.Warn("APP_VERSION is not provided")
		AppVersion = ""
	} else {
		AppVersion = appVersion
	}

	argocdNamespace, avail := os.LookupEnv("ARGOCD_NAMESPACE")
	if !avail {
		logrus.Warn("defaulting to `argocd` namespace")
		ArgocdNamespace = DefaultArgoCDNamespace
	} else {
		ArgocdNamespace = argocdNamespace
	}
}
