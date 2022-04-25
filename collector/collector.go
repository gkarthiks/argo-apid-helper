package collector

import (
	"github.com/doitintl/kube-no-trouble/pkg/judge"
	"github.com/gkarthiks/argo-apid-helper/config"
	"k8s.io/client-go/rest"
)

type Collector interface {
	Get() ([]map[string]interface{}, error)
	Name() string
}

type VersionCollector interface {
	GetServerVersion() (*judge.Version, error)
}

type commonCollector struct {
	name string
}

func newCommonCollector(name string) *commonCollector {
	return &commonCollector{
		name: name,
	}
}

func (c *commonCollector) Name() string {
	return c.name
}

func InitCollectors(config *Config, restConfig *rest.Config) []Collector {
	collectors := []Collector{}
	if config.Cluster {
		collector, err := NewClusterCollector(restConfig, &ClusterOpts{}, config.AdditionalKinds)
		collectors = storeCollector(collector, err, collectors)
	}
	return collectors
}

func storeCollector(collector Collector, err error, collectors []Collector) []Collector {
	if err != nil {
		config.Log.Errorf("Failed to initialize collector: %v", collector)
	} else {
		collectors = append(collectors, collector)
	}
	return collectors
}
