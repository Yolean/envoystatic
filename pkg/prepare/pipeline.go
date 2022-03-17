package prepare

import "github.com/yolean/envoystatic/v1/pkg/routeconfig"

type Pipeline interface {
	Process() (*routeconfig.RouteContent, error)
}
