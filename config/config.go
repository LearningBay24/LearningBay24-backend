package config

import (
	"io/ioutil"
	"os"
	"strings"

	"git.sr.ht/~sircmpwn/getopt"
	"github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
)

type DB struct {
	Host     string
	Port     int16
	User     string
	Pass     string
	Database string
}

type Files struct {
	Path             string
	AllowedFileTypes []string
}

type Secrets struct {
	JWTSecret string
}

type Config struct {
	Domain      string
	Secure      bool
	Environment string
	LogLevel    string
	DB          DB
	Files       Files
	Secrets     Secrets
}

var (
	Conf Config
)

func parseCLI() {
	opts, _, err := getopt.Getopts(os.Args, "v")
	if err != nil {
		log.Fatalf("Unable to parse command line arguments: %s\n", err.Error())
	}

	for _, opt := range opts {
		switch opt.Option {
		case 'v':
			Conf.LogLevel = "debug"
		}
	}
}

func InitConfig() {
	data, err := ioutil.ReadFile("./config.toml")
	if err != nil {
		log.Fatalf("Unable to read file: %s\n", err.Error())
	}
	err = toml.Unmarshal(data, &Conf)
	if err != nil {
		log.Fatalf("Unable to unmarshal config file: %s\n", err.Error())
	}
	if Conf.LogLevel == "" {
		Conf.LogLevel = "info"
	}
	parseCLI()
}

func InitLogger() {
	var level log.Level

	switch strings.ToLower(Conf.LogLevel) {
	case "trace":
		level = log.TraceLevel
	case "debug":
		level = log.DebugLevel
	case "info":
		level = log.InfoLevel
	case "warn":
		level = log.WarnLevel
	case "error":
		level = log.ErrorLevel
	case "fatal":
		level = log.FatalLevel
	case "panic":
		level = log.PanicLevel
	}

	log.SetLevel(level)
}
