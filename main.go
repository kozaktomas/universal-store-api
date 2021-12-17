package main

import (
	"fmt"
	"github.com/kozaktomas/universal-store-api/config"
	"github.com/kozaktomas/universal-store-api/storage"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
)

var (
	app                   = kingpin.New("usa", "Universal store API [USA] => Runs HTTP JSON REST API based on YAML configuration.")
	runCommand            = app.Command("run", "Run the application")
	runCommandConfig      = runCommand.Arg("config-file", "Path to configuration file").Required().String()
	runCommandStorageType = runCommand.Arg("storage-type", "Type of the storage").Required().String()
	verbose               = app.Flag("verbose", "Verbose mode.").Short('v').Bool()
)

func main() {
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case runCommand.FullCommand():
		run()
	}
}

func run() {
	cfg, err := config.ParseConfig(*runCommandConfig)
	if err != nil {
		abort(err)
	}

	serviceNames := cfg.GetServiceNames()

	stg, err := storage.CreateStorageByType(*runCommandStorageType, serviceNames)
	if err != nil {
		abort(err)
	}

	endpoints := make(map[string]Service, len(serviceNames))
	for _, serviceConfig := range cfg.ServiceConfigs {
		endpoints[serviceConfig.Name] = Service{
			Cfg:     serviceConfig,
			Storage: stg,
		}
	}

	server, err := createHttpServer(endpoints)
	if err != nil {
		abort(err)
	}
	server.Run(8080)
}

func abort(err error) {
	fmt.Println()
	fmt.Println("Something went wrong:")
	fmt.Println(err)
	fmt.Println()
	if *verbose {
		panic(err)
	}
	os.Exit(1)
}
