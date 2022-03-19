package main

import (
	"flag"
	"io/ioutil"
	"os"

	"github.com/yolean/envoystatic/v1/pkg/prepare"
	"github.com/yolean/envoystatic/v1/pkg/routeconfig"
	"go.uber.org/zap"
)

var (
	routeCommand *flag.FlagSet
	in           string
	out          string
	docroot      string
	rdsyaml      string
)

func init() {
	routeCommand = flag.NewFlagSet("route", flag.ExitOnError)
	routeCommand.StringVar(&in, "in", "/workspace", "Root path to content source")
	routeCommand.StringVar(&out, "out", "/var/docroot", "Root path to content destination")
	routeCommand.StringVar(&docroot, "docroot", "/var/docroot", "Runtime container's path to out dir")
	routeCommand.StringVar(&rdsyaml, "rdsyaml", "-", "Destination for RDS yaml")
}

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	undo := zap.ReplaceGlobals(logger)
	defer undo()

	switch os.Args[1] {
	case "route":
		routeCommand.Parse(os.Args[2:])
	default:
		zap.L().Fatal("Unknown subcommand", zap.String("arg", os.Args[1]))
	}

	if in == "" {
		zap.L().Fatal("--in path is required")
	}
	if in == "" {
		zap.L().Fatal("--out path is required")
	}

	var pipeline prepare.Pipeline
	pipeline, err = prepare.NewHashed(in, out)
	if err != nil {
		zap.L().Fatal("Failed initialize pipeline",
			zap.String("in", in),
			zap.String("out", out),
			zap.Error(err),
		)
	}

	content, err := pipeline.Process()
	if err != nil {
		zap.L().Fatal("Failed to process",
			zap.String("in", in),
			zap.String("out", out),
			zap.Error(err),
		)
	}
	routeConfiguration := routeconfig.Generate(docroot, content)

	rds := routeconfig.RdsYaml(routeConfiguration)
	if rdsyaml == "-" {
		os.Stdout.Write(rds)
	} else {
		err = ioutil.WriteFile(rdsyaml, rds, 0644)
		if err != nil {
			zap.L().Fatal("Failed to write RDS yaml", zap.String("path", rdsyaml))
		}
	}
}
