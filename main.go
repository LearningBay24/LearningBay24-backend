package main

import (
	"learningbay24.de/backend/config"
)

func main() {
	config.InitConfig()
	// TODO: do something with the db handle
	_ = config.SetupDbHandle()
}
