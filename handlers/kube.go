package handlers

import (
	"context"
	"fmt"
	"github.com/argoproj/argo-cd/v2/common"
	"github.com/gkarthiks/argo-apid-helper/config"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"strings"
)

// PopulateArgoClusters will populate cluster secrets that are maintained by ArgoCD
func PopulateArgoClusters(ctx context.Context) ([]v1.Secret, error) {
	logrus.Debugln("getting the argocd managed cluster list via its secrets")
	clusterSecretsList, err := config.KubeClient.Clientset.CoreV1().Secrets(config.ArgocdNamespace).List(ctx, metav1.ListOptions{LabelSelector: common.LabelKeySecretType + "=" + common.LabelValueSecretTypeCluster})
	if err != nil {
		logrus.Errorf("error occured while listing the argocd secrets: %v", err)
		return nil, fmt.Errorf("error occured while listing the argocd secrets: %v", err.Error())
	}
	if clusterSecretsList == nil {
		logrus.Errorln("no cluster secrets found that are managed by ArgoCD")
		return nil, fmt.Errorf("no secrets found under the %s namespace for ArgoCD Clusters", config.ArgocdNamespace)
	}
	logrus.Debugf("total cluster secrets found that are managed by argocd: %v", len(clusterSecretsList.Items))

	// kind of refreshing the list of argocd cluster secrets everytime this function is called
	// in a way renewing the cache in-directly to be up-to-date as much as possible
	config.ArgoManagedClusterSecrets = clusterSecretsList.Items

	return clusterSecretsList.Items, nil
}

// PopulateArgoClusterNames will populate the names of the clusters that are maintained by ArgoCD
func PopulateArgoClusterNames(ctx context.Context) (sets.String, error) {
	logrus.Info("getting the argocd managed cluster names list via its secrets")
	clusterSecretsList, err := PopulateArgoClusters(ctx)
	logrus.Debug("extracting the cluster names from the secret")
	for _, clusterSecret := range clusterSecretsList {
		sanitizedClusterName := strings.TrimSpace(string(clusterSecret.Data["name"]))
		config.ArgoManagedClusterNames.Insert(sanitizedClusterName)
		config.ArgoClusterNameToSecretMap[sanitizedClusterName] = clusterSecret
	}
	if err != nil {
		return nil, err
	}
	return config.ArgoManagedClusterNames, nil
}
