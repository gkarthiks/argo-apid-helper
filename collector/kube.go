package collector

import (
	"fmt"
	"github.com/doitintl/kube-no-trouble/pkg/judge"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
)

type kubeCollector struct {
	discoveryClient discovery.DiscoveryInterface
	restConfig      *rest.Config
}

func newKubeCollector(restConfig *rest.Config, discoveryClient discovery.DiscoveryInterface) (*kubeCollector, error) {
	col := &kubeCollector{}
	if discoveryClient != nil {
		col.discoveryClient = discoveryClient
	} else {
		var err error
		col.restConfig = restConfig

		if col.discoveryClient, err = discovery.NewDiscoveryClientForConfig(col.restConfig); err != nil {
			return nil, fmt.Errorf("failed to create client: %w", err)
		}
	}
	return col, nil
}

func (c *kubeCollector) GetRestConfig() *rest.Config {
	return c.restConfig
}

func (c *kubeCollector) GetServerVersion() (*judge.Version, error) {
	version, err := c.discoveryClient.ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get server version %w", err)
	}

	return judge.NewVersion(version.String())
}
