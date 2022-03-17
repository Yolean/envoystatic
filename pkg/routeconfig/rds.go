package routeconfig

import (
	"bytes"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/gogo/protobuf/jsonpb"
	"go.uber.org/zap"
	"sigs.k8s.io/yaml"
)

func RdsYaml(routeConfiguration *route.RouteConfiguration) []byte {

	m := jsonpb.Marshaler{OrigName: true}

	routeJson := &bytes.Buffer{}
	if err := m.Marshal(routeJson, routeConfiguration); err != nil {
		zap.L().Fatal("Failed to marshal route config",
			zap.Error(err),
		)
	}

	rds := []byte(`{
		"version_info": "0",
		"resources": [
			{
				"@type": "type.googleapis.com/envoy.config.route.v3.RouteConfiguration",
				`)
	rds = append(rds, routeJson.Bytes()[1:]...)
	rds = append(rds, []byte("]}")[:]...)

	rdsYaml, err := yaml.JSONToYAML(rds)
	if err != nil {
		zap.L().Fatal("Failed to marshal JSON to yaml",
			zap.Error(err),
		)
	}
	return rdsYaml
}
