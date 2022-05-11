package config

import (
	"io/ioutil"
	"log"

	"github.com/pelletier/go-toml"
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
	DB      DB
	Files   Files
	Secrets Secrets
}

var (
	Conf Config
)

func InitConfig() {
	data, err := ioutil.ReadFile("./config.toml")
	if err != nil {
		log.Fatal(err)
	}
	err = toml.Unmarshal(data, &Conf)
	if err != nil {
		log.Fatal(err)
	}
}
