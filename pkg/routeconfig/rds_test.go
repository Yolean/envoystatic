package routeconfig_test

import (
	"fmt"
	"testing"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/yolean/envoystatic/v1/pkg/routeconfig"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"sigs.k8s.io/yaml"
)

func TestRouteYaml(t *testing.T) {
	logger := zaptest.NewLogger(t)
	defer logger.Sync()
	undo := zap.ReplaceGlobals(logger)
	defer undo()

	config := &route.RouteConfiguration{
		Name: "test",
		VirtualHosts: []*route.VirtualHost{{
			Name:    "test",
			Domains: []string{"*"},
			Routes:  []*route.Route{},
		}},
	}

	rds := routeconfig.RdsYaml(config)
	fmt.Printf("rds: %s", rds)

	var result map[string]interface{}
	err := yaml.Unmarshal(rds, &result)
	if err != nil {
		t.Errorf("RDS yaml could not be unmarshalled: %s", rds)
	}
	if result["version_info"] != "0" {
		t.Errorf("Expected version_info: %s", rds)
	}
}
