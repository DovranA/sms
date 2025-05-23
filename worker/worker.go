package worker

import (
	"fmt"
	"log"
	"time"

	"github.com/alexgear/sms/common"
	"github.com/alexgear/sms/database"
	otp "github.com/alexgear/sms/event-schemas"
	"github.com/alexgear/sms/modem"
	amqp "github.com/rabbitmq/amqp091-go"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/protobuf/proto"
)

var (
	err error
)

type Config struct {
	RabbitMQHost     string
	RabbitMQPort     string
	RabbitMQUser     string
	RabbitMQPassword string
}

func InitWorker(cfg *Config) {
	messages := make(chan common.SMS)
	go consumer(messages)

	go func() {
		for {
			log.Println("Reconnect to RabbitMQ...")
			rabbitConn, rabbitCh, rabbitQue := initRabbitMQ(cfg)
			go producer(rabbitCh, rabbitQue, messages)

			time.Sleep(2 * time.Minute)

			log.Println("Connection: Close current connection...")
			if rabbitCh != nil {
				_ = rabbitCh.Close()
			}
			if rabbitConn != nil {
				_ = rabbitConn.Close()
			}
		}
	}()
}

func initRabbitMQ(cfg *Config) (*amqp.Connection, *amqp.Channel, amqp.Queue) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", cfg.RabbitMQUser, cfg.RabbitMQPassword, cfg.RabbitMQHost, cfg.RabbitMQPort)
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatalf("Can't connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Can't open channel: %v", err)
	}

	err = ch.ExchangeDeclare("otp", "topic", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Can't announce exchange: %v", err)
	}

	que, err := ch.QueueDeclare("otp.generated:send.sms", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Can't announce queue: %v", err)
	}

	return conn, ch, que
}

func consumer(messages chan common.SMS) {
	for msg := range messages {
		log.Println("consumer: processing", msg.UUID)
		err = modem.SendMessage(msg.Mobile, msg.Body)
		if err != nil {
			msg.Status = "error"
			log.Println("consumer: error sending", msg.UUID, err)
		} else {
			msg.Status = "sent"
		}
		msg.Retries++
		database.UpdateMessageStatus(msg)
	}
}

func producer(ch *amqp.Channel, que amqp.Queue, messages chan common.SMS) {
	msgs, err := ch.Consume(
		que.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("Error consume: %v", err)
		return
	}

	for d := range msgs {
		var message otp.OtpGenerated
		err = proto.Unmarshal(d.Body, &message)
		if err != nil {
			log.Printf("Error parse Protobuf: %v", err)
			continue
		}

		uuid := uuid.NewV1()
		sms := &common.SMS{
			UUID:   uuid.String(),
			Mobile: message.GetPhone(),
			Body:   message.GetValue(),
			Status: "pending",
		}

		err = database.InsertMessage(sms)
		if err != nil {
			log.Printf("Error insert to DB: %v", err)
			continue
		}

		messages <- *sms
	}
}
