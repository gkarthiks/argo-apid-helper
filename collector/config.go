package collector

import (
	"fmt"
	"github.com/doitintl/kube-no-trouble/pkg/judge"
	"github.com/doitintl/kube-no-trouble/pkg/printer"
	"strings"
	"unicode"
)

type Config struct {
	// AdditionalKinds yet to be implemented
	AdditionalKinds []string
	Cluster         bool
	Output          string
	TargetVersion   *judge.Version
}

func NewCollectorConfig() (*Config, error) {
	config := Config{
		TargetVersion: &judge.Version{},
		Cluster:       true,
		Output:        "json",
	}
	if _, err := printer.ParsePrinter(config.Output); err != nil {
		return nil, fmt.Errorf("failed to validate argument output: %w", err)
	}
	if err := validateAdditionalResources(config.AdditionalKinds); err != nil {
		return nil, fmt.Errorf("failed to validate arguments: %w", err)
	}
	if config.TargetVersion.Version == nil {
		config.TargetVersion = nil
	}
	return &config, nil
}

// validateAdditionalResources check that all resources are provided in full form
// resource.version.group.com. E.g. managedcertificate.v1beta1.networking.gke.io
func validateAdditionalResources(resources []string) error {
	for _, r := range resources {
		parts := strings.Split(r, ".")
		if len(parts) < 4 {
			return fmt.Errorf("failed to parse additional Kind, full form Kind.version.group.com is expected, instead got: %s", r)
		}
		if !unicode.IsUpper(rune(parts[0][0])) {
			return fmt.Errorf("failed to parse additional Kind, Kind is expected to be capitalized by convention, instead got: %s", parts[0])
		}
	}
	return nil
}
