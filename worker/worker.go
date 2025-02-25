package worker

import (
	"fmt"
	"log"

	"github.com/alexgear/sms/common"
	"github.com/alexgear/sms/database"
	otp "github.com/alexgear/sms/event-schemas"
	"github.com/alexgear/sms/modem"
	amqp "github.com/rabbitmq/amqp091-go"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/protobuf/proto"
)

var (
	err        error
	rabbitConn *amqp.Connection
	rabbitCh   *amqp.Channel
	rabbitQue  amqp.Queue
)

type Config struct {
	RabbitMQHost     string
	RabbitMQPort     string
	RabbitMQUser     string
	RabbitMQPassword string
}

func InitWorker(cfg *Config) {
	initRabbitMQ(cfg)
	messages := make(chan common.SMS)
	go producer(messages)
	go consumer(messages)
}
func initRabbitMQ(cfg *Config) {
	var err error
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", cfg.RabbitMQUser, cfg.RabbitMQPassword, cfg.RabbitMQHost, cfg.RabbitMQPort)
	rabbitConn, err = amqp.Dial(url)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	rabbitCh, err = rabbitConn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	err = rabbitCh.ExchangeDeclare(
		"otp",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare an exchange: %v", err)
	}

	rabbitQue, err = rabbitCh.QueueDeclare(
		"otp.generated:send.sms",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}
}
func consumer(messages chan common.SMS) {
	for {
		message := <-messages
		log.Println("consumer: processing", message.UUID)
		err = modem.SendMessage(message.Mobile, message.Body)
		if err != nil {
			message.Status = "error"
			log.Println("consumer: failed to process", message.UUID, err)
		} else {
			message.Status = "sent"
		}
		message.Retries++
		database.UpdateMessageStatus(message)
	}
}

func producer(messages chan common.SMS) {
	msgs, err_msgs := rabbitCh.Consume(
		rabbitQue.Name,
		"otp.generated",
		true,
		false,
		false,
		false,
		nil,
	)
	if err_msgs != nil {
		log.Panicf("%v: %v", msgs, err_msgs)
	}
	for d := range msgs {
		encodedMessage := d.Body
		var message otp.OtpGenerated

		err = proto.Unmarshal(encodedMessage, &message)
		if err != nil {
			log.Fatalf("Ошибка разбора Protobuf: %v", err)
		}

		Mobile := message.GetPhone()
		Body := message.GetValue()
		uuid := uuid.NewV1()
		sms := &common.SMS{
			UUID:   uuid.String(),
			Mobile: Mobile,
			Body:   Body,
			Status: "pending"}
		err = database.InsertMessage(sms)
		if err != nil {
			log.Fatalf("Ошибка insert to db: %v", err)
		}
		msg := common.SMS{
			UUID:   sms.UUID,
			Mobile: sms.Mobile,
			Body:   sms.Body,
			Status: sms.Status,
		}
		messages <- msg
	}

}
