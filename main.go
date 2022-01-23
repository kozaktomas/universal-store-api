package main

import (
	"github.com/kozaktomas/universal-store-api/config"
	"github.com/kozaktomas/universal-store-api/storage"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"strings"
)

var (
	app                   = kingpin.New("usa", "Universal store API [USA] => Runs HTTP JSON REST API based on YAML configuration.")
	runCommand            = app.Command("run", "Run the application")
	runCommandConfig      = runCommand.Arg("config-file", "Path to configuration file").Required().String()
	runCommandStorageType = runCommand.Arg("storage-type", "Type of the storage").Required().String()
	verbose               = app.Flag("verbose", "Verbose mode sets log level to trace").Short('v').Bool()
)

func main() {
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case runCommand.FullCommand():
		run()
	}
}

func run() {
	logger := createLogger()
	logger.Info("Program starting...")

	cfg, err := config.ParseConfig(*runCommandConfig, logger)
	if err != nil {
		logger.WithError(err).Fatalf("could not parse configuration")
	}

	if err = cfg.Validate(); err != nil {
		logger.WithError(err).Fatalf("invalid configuration file")
	}

	serviceNames := cfg.GetServiceNames()
	if err = ValidateServiceNames(serviceNames); err != nil {
		logger.WithError(err).Fatalf("service name validation falied")
	}

	stg, err := storage.CreateStorageByType(*runCommandStorageType, serviceNames)
	if err != nil {
		logger.WithError(err).Fatalf("could not create storage")
	}

	logger.Infof("Storage prepared for %d service(s) (%s)", len(serviceNames), strings.Join(serviceNames, ", "))

	endpoints := make(map[string]Service, len(serviceNames))
	for _, serviceConfig := range cfg.ServiceConfigs {
		endpoints[serviceConfig.Name] = Service{
			Cfg:     serviceConfig,
			Storage: stg,
		}
	}

	server, err := createHttpServer(endpoints, logger)
	if err != nil {
		logger.WithError(err).Fatalf("could not create http server")
	}
	server.Run(8080)
}

func createLogger() *logrus.Logger {
	logger := logrus.New()

	logger.SetOutput(os.Stdout)

	levelConfig := os.Getenv("LOG_LEVEL")
	level, err := logrus.ParseLevel(levelConfig)
	if *verbose {
		logger.SetLevel(logrus.TraceLevel)
		logger.Infof("using verbose log_level %q", "trace")
	} else if levelConfig == "" {
		logger.SetLevel(logrus.InfoLevel)
		logger.Errorf("using default log_level %q", "info")
	} else if err != nil {
		logger.SetLevel(logrus.InfoLevel)
		logger.Infof("could not set %q as log_level (invalid value). Using default %q", levelConfig, "info")
	} else {
		logger.SetLevel(level)
		logger.Infof("log_level set to %q", levelConfig)
	}

	return logger
}
