package routeconfig

import (
	"path/filepath"
	"regexp"
	"strings"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var ValidateETag = regexp.MustCompile(`^"[0-9a-f]+"$`) // Though we actually don't care what's between the quotes

func Generate(docroot string, content *RouteContent) *route.RouteConfiguration {

	if !strings.HasPrefix(docroot, "/") {
		zap.L().Fatal("root must be absolute", zap.String("docroot", docroot))
	}
	if strings.HasSuffix(docroot, "/") {
		zap.L().Fatal("root should lack traling slash", zap.String("docroot", docroot))
	}

	routes := []*route.Route{}
	maxsize := uint32(10 * 1024 * 1024)

	for _, c := range content.Items {
		if c.ContentPath != "" && c.Content != nil {
			zap.L().Fatal("only one of ContentPath and Content should be set", zap.String("path", c.Path))
		}
		if c.ContentPath == "" && c.Content == nil {
			zap.L().Fatal("either ContentPath or Content must be set", zap.String("path", c.Path))
		}
		if c.Content != nil {
			zap.L().Fatal("inline Content is currently unsupported", zap.String("path", c.Path))
		}
		if !ValidateETag.Match([]byte(c.ETag)) {
			zap.L().Fatal("Invalid ETag", zap.String("path", c.Path), zap.String("etag", c.ETag))
		}

		headers := []*core.HeaderValueOption{}
		if c.ContentType != "" {
			headers = append(headers, &core.HeaderValueOption{
				AppendAction: core.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD,
				Header: &core.HeaderValue{
					Key:   "content-type",
					Value: c.ContentType,
				},
			})
		}

		readpath := filepath.Join(docroot, c.ContentPath)
		r := &route.Route{
			Match: &route.RouteMatch{
				PathSpecifier: &route.RouteMatch_Path{
					Path: "/" + c.Path,
				},
			},
			Action: &route.Route_DirectResponse{
				DirectResponse: &route.DirectResponseAction{
					Status: 200,
					Body: &core.DataSource{
						Specifier: &core.DataSource_Filename{
							Filename: readpath,
						},
					},
				},
			},
			ResponseHeadersToAdd: headers,
		}
		routes = append(routes, r)
		zap.L().Info("Route generated for", zap.String("path", readpath))
	}

	config := &route.RouteConfiguration{
		MaxDirectResponseBodySizeBytes: wrapperspb.UInt32(maxsize),
		Name:                           "",
		VirtualHosts: []*route.VirtualHost{{
			Name:    "static",
			Domains: []string{"*"},
			Routes:  routes,
		}},
	}

	return config
}
