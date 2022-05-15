package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	argoAppV1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/doitintl/kube-no-trouble/pkg/judge"
	"github.com/doitintl/kube-no-trouble/pkg/printer"
	"github.com/doitintl/kube-no-trouble/pkg/rules"
	"github.com/gin-gonic/gin"
	"github.com/gkarthiks/argo-apid-helper/collector"
	"github.com/gkarthiks/argo-apid-helper/config"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/pointer"
	"net/http"
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

// GetArgoClusters will list the name of all the Kubernetes Clusters
// that are managed by ArgoCD GitOps engine
func GetArgoClusters(c *gin.Context) {
	logrus.Info("listing the clusters managed by ArgoCD")
	clusterNamesList, err := PopulateArgoClusterNames(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("error occured while populating the list: %v", err.Error()),
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"totalClusters": clusterNamesList.Len(),
		"clusters":      clusterNamesList.List(),
	})
}

// ListAPIDeprecations lists the api deprecations for all the clusters that are managed
// by the ArgoCD
func ListAPIDeprecations(c *gin.Context) {
	logrus.Info("listing the clusters managed by ArgoCD")

	clusterSecrets, err := PopulateArgoClusters(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("error occured while populating the list of argo clusters: %v", err.Error()),
		})
	}

	logrus.Debugf("total number of clusters found that are managed by ArgoCD: %d", len(clusterSecrets))
	if config.AppMode != config.AppModeProd {
		logrus.Debugln("Listing the secrets that are found as cluster secrets")
		for _, sec := range clusterSecrets {
			logrus.Debugf("Secret Name: %v", sec.Name)
		}
	}
	clusterList := argoAppV1.ClusterList{
		Items: make([]argoAppV1.Cluster, len(clusterSecrets)),
	}

	hasInClusterCredentials := false
	for i, clusterSecret := range clusterSecrets {
		cluster, err := secretToCluster(&clusterSecret)
		if err != nil || cluster == nil {
			logrus.Errorf("unable to convert cluster secret to cluster object '%s': %v", clusterSecret.Name, err)
		}

		clusterList.Items[i] = *cluster
		if cluster.Server == argoAppV1.KubernetesInternalAPIServerAddr {
			hasInClusterCredentials = true
		}
	}
	if !hasInClusterCredentials {
		localCluster := getLocalCluster(config.KubeClient.Clientset)
		if localCluster != nil {
			clusterList.Items = append(clusterList.Items, *localCluster)
		}
	}

	if config.AppMode != config.AppModeProd {
		logrus.Debugln("listing all the cluster names")
		for idx, clusterName := range clusterList.Items {
			logrus.Debugf("%d ) \t %v", idx, clusterName.Name)
		}
	}

	var deprecationResults []config.DeprecationResults
	for i := 0; i < len(clusterList.Items); i++ {
		deprecationResult := getDeprecationForCluster(c, clusterList.Items[i])
		deprecationResults = append(deprecationResults, *deprecationResult)
	}
	c.JSON(200, gin.H{
		"deprecationResults": deprecationResults,
	})
}

// getDeprecationForCluster works on the given cluster and returns the list of
// API deprectation and associated workloads deployed against it
func getDeprecationForCluster(ctx context.Context, cluster argoAppV1.Cluster) *config.DeprecationResults {
	logrus.Infof("starting to work on the %s cluster", cluster.Name)
	var err error
	collectorConfig, _ := collector.NewCollectorConfig()
	logrus.Infoln("Initializing collectors and retrieving data")
	initCollectors := collector.InitCollectors(collectorConfig, cluster.RawRestConfig())

	collectorConfig.TargetVersion, err = getServerVersion(collectorConfig.TargetVersion, initCollectors)
	// If there's an error in communication with the cluster, return error for results
	// against the cluster name
	if err != nil {
		logrus.Errorf("error occured while getting the deprecation result for %s cluster: %v", cluster.Name, err.Error())
		return &config.DeprecationResults{
			ClusterName: cluster.Name,
			Result:      err.Error(),
		}
	}

	if collectorConfig.TargetVersion != nil {
		logrus.Infof("Target K8s version is %s", collectorConfig.TargetVersion.String())
	}

	collectors := getCollectors(initCollectors)

	var additionalKinds []schema.GroupVersionKind
	for _, ar := range collectorConfig.AdditionalKinds {
		gvr, _ := schema.ParseKindArg(ar)
		additionalKinds = append(additionalKinds, *gvr)
	}

	loadedRules, err := rules.FetchRegoRules(additionalKinds)
	if err != nil {
		logrus.Fatalln("name: Rules; Failed to load rules")
	}

	judge, err := judge.NewRegoJudge(&judge.RegoOpts{}, loadedRules)
	if err != nil {
		logrus.Fatalf("name: Rego; Failed to initialize decision engine: %v", err)
	}

	results, err := judge.Eval(collectors)
	if err != nil {
		logrus.Fatalf("name: Rego; Failed to evaluate input: %v", err)
	}

	results, err = printer.FilterNonRelevantResults(results, collectorConfig.TargetVersion)
	if err != nil {
		logrus.Fatalf("name: Rego; Failed to filter results: %v", err)
	}

	return &config.DeprecationResults{
		ClusterName: cluster.Name,
		Result:      results,
	}
}

// GetTargetClusterDeprecations will get the list of deprecations and the workloads
// against those deprecated workloads on a targeted cluster
func GetTargetClusterDeprecations(c *gin.Context) {
	logrus.Info("processing deprecations for the targeted cluster")
	targetCluster := c.Param("clusterName")
	logrus.Debugf("targeting the cluster: %s and checking if its a cluster managed by argocd ", targetCluster)
	var deprecationResult *config.DeprecationResults
	if config.ArgoManagedClusterNames.Has(targetCluster) {
		logrus.Debugf("%s is a valid argocd managed cluster and proceeding with the deprecation list processing", targetCluster)
		deprecationResult = proccedWithDeprecation(c, targetCluster)
	} else if PopulateArgoClusterNames(c); config.ArgoManagedClusterNames.Has(targetCluster) {
		logrus.Debugf("%s was found after refreshing the list of ArgoCD pre-populated cluster names", targetCluster)
		deprecationResult = proccedWithDeprecation(c, targetCluster)
	} else {
		logrus.Errorf("%s not found from the list cluster managed by ArgoCD; It's not a valid cluster managed by ArgoCD", targetCluster)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("%s not found from the list cluster managed by ArgoCD; It's not a valid cluster managed by ArgoCD", targetCluster),
		})
	}
	logrus.Debugf("returning the resultant data for %s cluster", targetCluster)
	c.JSON(http.StatusOK, gin.H{
		"clusterName": deprecationResult.ClusterName,
		"results":     deprecationResult.Result,
	})

}

func proccedWithDeprecation(ctx context.Context, clusterName string) *config.DeprecationResults {
	logrus.Debugf("proceeding with the deprecation analysis for the target cluster: %s", clusterName)
	targetClusterSecret := config.ArgoClusterNameToSecretMap[clusterName]
	cluster, err := secretToCluster(&targetClusterSecret)
	if err != nil || cluster == nil {
		logrus.Errorf("unable to convert cluster secret to cluster object '%s': %v", targetClusterSecret.Name, err)
		return &config.DeprecationResults{
			ClusterName: clusterName,
			Result:      fmt.Errorf("unable to convert cluster secret to cluster object '%s': %v", targetClusterSecret.Name, err),
		}
	}
	return getDeprecationForCluster(ctx, *cluster)
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
			logrus.Warnf("Error while parsing date in cluster secret '%s': %v", s.Name, err)
		} else {
			refreshRequestedAt = &metav1.Time{Time: requestedAt}
		}
	}
	var shard *int64
	if shardStr := s.Data["shard"]; shardStr != nil {
		if val, err := strconv.Atoi(string(shardStr)); err != nil {
			logrus.Warnf("Error while parsing shard in cluster secret '%s': %v", s.Name, err)
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
			logrus.Errorf("collector name: %v; Failed to retrieve data from collector: %v", c.Name(), err)
		} else {
			inputs = append(inputs, rs...)
			logrus.Infof("collector name: %v; Retrieved %d resources from collector", c.Name(), len(rs))
		}
	}
	return inputs
}
