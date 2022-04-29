package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/argoproj/argo-cd/v2/common"
	argoAppV1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/doitintl/kube-no-trouble/pkg/judge"
	"github.com/doitintl/kube-no-trouble/pkg/printer"
	"github.com/doitintl/kube-no-trouble/pkg/rules"
	"github.com/gin-gonic/gin"
	"github.com/gkarthiks/argo-apid-helper/collector"
	"github.com/gkarthiks/argo-apid-helper/config"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/pointer"
	"strconv"
	"strings"
	"time"
)

// HealthZ handler will return http.Response with `200 OK` for
// health pings
func HealthZ(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func GetArgoClusters(c *gin.Context) {
	c.JSON(501, gin.H{
		"message": "Not Implemented",
	})
}

func ListAPIDeprecations(c *gin.Context) {
	config.Log.Info("listing the clusters managed by ArgoCD")

	clusterSecretsList, err := config.K8s.Clientset.CoreV1().Secrets(config.ArgocdNamespace).List(context.Background(),
		metav1.ListOptions{LabelSelector: common.LabelKeySecretType + "=" + common.LabelValueSecretTypeCluster})
	if err != nil {
		config.Log.Errorf("error occured while listing the argocd secrets: %v", err)
	}
	if clusterSecretsList == nil {
		config.Log.Errorln("no cluster secrets found that are managed by ArgoCD")
	}
	config.Log.Debugf("total cluster secrets found that are managed by argocd: %v", len(clusterSecretsList.Items))

	clusterSecrets := clusterSecretsList.Items
	if config.AppMode != config.AppModeProd {
		config.Log.Debugln("Listing the secrets that are found as cluster secrets")
		for _, sec := range clusterSecrets {
			config.Log.Debugf("Secret Name: %v", sec.Name)
		}
	}
	clusterList := argoAppV1.ClusterList{
		Items: make([]argoAppV1.Cluster, len(clusterSecrets)),
	}

	hasInClusterCredentials := false
	for i, clusterSecret := range clusterSecrets {
		cluster, err := secretToCluster(&clusterSecret)
		if err != nil || cluster == nil {
			config.Log.Errorf("unable to convert cluster secret to cluster object '%s': %v", clusterSecret.Name, err)
		}

		clusterList.Items[i] = *cluster
		if cluster.Server == argoAppV1.KubernetesInternalAPIServerAddr {
			hasInClusterCredentials = true
		}
	}
	if !hasInClusterCredentials {
		localCluster := getLocalCluster(config.K8s.Clientset)
		if localCluster != nil {
			clusterList.Items = append(clusterList.Items, *localCluster)
		}
	}

	if config.AppMode != config.AppModeProd {
		config.Log.Debugln("listing all the cluster names")
		for idx, clusterName := range clusterList.Items {
			config.Log.Debugf("%d ) \t %v", idx, clusterName.Name)
		}
	}
	type DeprecationResults struct {
		ClusterName string      `json:"clusterName"`
		Result      interface{} `json:"result"`
	}

	var deprecationResults []DeprecationResults
	for i := 0; i < len(clusterList.Items); i++ {
		config.Log.Infof("starting to work on the %s cluster", clusterList.Items[i].Name)

		collectorConfig, _ := collector.NewCollectorConfig()
		config.Log.Infoln("Initializing collectors and retrieving data")
		initCollectors := collector.InitCollectors(collectorConfig, clusterList.Items[i].RawRestConfig())

		collectorConfig.TargetVersion, err = getServerVersion(collectorConfig.TargetVersion, initCollectors)
		if err != nil {
			deprecationResults = append(deprecationResults, DeprecationResults{
				ClusterName: clusterList.Items[i].Name,
				Result:      err.Error(),
			})
			continue
		}
		if collectorConfig.TargetVersion != nil {
			config.Log.Infof("Target K8s version is %s", collectorConfig.TargetVersion.String())
		}

		collectors := getCollectors(initCollectors)

		var additionalKinds []schema.GroupVersionKind
		for _, ar := range collectorConfig.AdditionalKinds {
			gvr, _ := schema.ParseKindArg(ar)
			additionalKinds = append(additionalKinds, *gvr)
		}

		loadedRules, err := rules.FetchRegoRules(additionalKinds)
		if err != nil {
			config.Log.Fatalln("name: Rules; Failed to load rules")
		}

		judge, err := judge.NewRegoJudge(&judge.RegoOpts{}, loadedRules)
		if err != nil {
			config.Log.Fatalf("name: Rego; Failed to initialize decision engine: %v", err)
		}

		results, err := judge.Eval(collectors)
		if err != nil {
			config.Log.Fatalf("name: Rego; Failed to evaluate input: %v", err)
		}

		results, err = printer.FilterNonRelevantResults(results, collectorConfig.TargetVersion)
		if err != nil {
			config.Log.Fatalf("name: Rego; Failed to filter results: %v", err)
		}

		deprecationResults = append(deprecationResults, DeprecationResults{
			ClusterName: clusterList.Items[i].Name,
			Result:      results,
		})
	}
	c.JSON(200, gin.H{
		"deprecationResults": deprecationResults,
	})
}

// secretToCluster converts a secret into a Cluster object
func secretToCluster(s *corev1.Secret) (*argoAppV1.Cluster, error) {
	var clusterConfig argoAppV1.ClusterConfig
	if len(s.Data["config"]) > 0 {
		if err := json.Unmarshal(s.Data["config"], &clusterConfig); err != nil {
			// This line has changed from the original Argo CD: now returns an error rather than panicing.
			return nil, err
		}
	}

	var namespaces []string
	for _, ns := range strings.Split(string(s.Data["namespaces"]), ",") {
		if ns = strings.TrimSpace(ns); ns != "" {
			namespaces = append(namespaces, ns)
		}
	}
	var refreshRequestedAt *metav1.Time
	if v, found := s.Annotations[argoAppV1.AnnotationKeyRefresh]; found {
		requestedAt, err := time.Parse(time.RFC3339, v)
		if err != nil {
			config.Log.Warnf("Error while parsing date in cluster secret '%s': %v", s.Name, err)
		} else {
			refreshRequestedAt = &metav1.Time{Time: requestedAt}
		}
	}
	var shard *int64
	if shardStr := s.Data["shard"]; shardStr != nil {
		if val, err := strconv.Atoi(string(shardStr)); err != nil {
			config.Log.Warnf("Error while parsing shard in cluster secret '%s': %v", s.Name, err)
		} else {
			shard = pointer.Int64Ptr(int64(val))
		}
	}
	cluster := argoAppV1.Cluster{
		ID:                 string(s.UID),
		Server:             strings.TrimRight(string(s.Data["server"]), "/"),
		Name:               string(s.Data["name"]),
		Namespaces:         namespaces,
		Config:             clusterConfig,
		RefreshRequestedAt: refreshRequestedAt,
		Shard:              shard,
	}
	return &cluster, nil
}

func getLocalCluster(clientset kubernetes.Interface) *argoAppV1.Cluster {
	config.InitLocalCluster.Do(func() {
		info, err := clientset.Discovery().ServerVersion()
		if err == nil {
			config.LocalCluster.ServerVersion = fmt.Sprintf("%s.%s", info.Major, info.Minor)
			config.LocalCluster.ConnectionState = argoAppV1.ConnectionState{Status: argoAppV1.ConnectionStatusSuccessful}
		} else {
			config.LocalCluster.ConnectionState = argoAppV1.ConnectionState{
				Status:  argoAppV1.ConnectionStatusFailed,
				Message: err.Error(),
			}
		}
	})
	cluster := config.LocalCluster.DeepCopy()
	now := metav1.Now()
	cluster.ConnectionState.ModifiedAt = &now
	return cluster
}

func getServerVersion(cv *judge.Version, collectors []collector.Collector) (*judge.Version, error) {
	if cv == nil {
		for _, c := range collectors {
			if versionCol, ok := c.(collector.VersionCollector); ok {
				version, err := versionCol.GetServerVersion()
				if err != nil {
					return nil, fmt.Errorf("failed to detect k8s version: %w", err)
				}
				return version, nil
			}
		}
	}
	return cv, nil
}

func getCollectors(collectors []collector.Collector) []map[string]interface{} {
	var inputs []map[string]interface{}
	for _, c := range collectors {
		rs, err := c.Get()
		if err != nil {
			config.Log.Errorf("collector name: %v; Failed to retrieve data from collector: %v", c.Name(), err)
		} else {
			inputs = append(inputs, rs...)
			config.Log.Infof("collector name: %v; Retrieved %d resources from collector", c.Name(), len(rs))
		}
	}
	return inputs
}
