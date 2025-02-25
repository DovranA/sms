package main

import (
	"log"

	"github.com/alexgear/sms/api"
	"github.com/alexgear/sms/config"
	"github.com/alexgear/sms/database"
	"github.com/alexgear/sms/modem"
	"github.com/alexgear/sms/worker"
)

func main() {
	cfg, err := config.New("config.toml")
	if err != nil {
		log.Fatalf("main: Invalid config: %s", err.Error())
	}

	db, err := database.InitDB("database.db")
	defer db.Close()
	if err != nil {
		log.Fatalf("main: Error initializing database: %s", err.Error())
	}

	err = modem.InitModem(cfg.ComPort, cfg.BaudRate)
	if err != nil {
		log.Fatalf("main: error initializing to modem. %s", err)
	}
	err = modem.Reset()
	if err != nil {
		log.Fatalf("main: error reseting modem. %s", err)
	}
	worker.InitWorker(&worker.Config{RabbitMQHost: cfg.RabbitMQHost, RabbitMQPort: cfg.RabbitMQPort, RabbitMQUser: cfg.RabbitMQUser, RabbitMQPassword: cfg.RabbitMQPassword})
	err = api.InitServer(cfg.ServerHost, cfg.ServerPort)
	if err != nil {
		log.Fatalf("main: Error starting server: %s", err.Error())
	}
}
