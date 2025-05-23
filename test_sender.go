package main

import (
	"fmt"
	"log"
	"time"

	"go.bug.st/serial"
)

func main() {
	// Открываем последовательный порт (укажите правильный ttyUSB)
	port, err := serial.Open("/dev/ttyUSB2", &serial.Mode{BaudRate: 115200})
	if err != nil {
		log.Fatal(err)
	}
	defer port.Close()

	// Функция для отправки AT-команд
	sendAT := func(cmd string) {
		fmt.Println(">>", cmd)
		port.Write([]byte(cmd + "\r")) // Отправка команды
		time.Sleep(1 * time.Second)    // Ожидание ответа
	}

	// Инициализация модема
	sendAT("AT")
	sendAT("AT+CMGF=1")                          // Включаем текстовый режим SMS
	sendAT(`AT+CMGS="+99363361498"`)             // Номер получателя
	port.Write([]byte("Привет, тестовое SMS!" + string([]byte{26}))) // Текст + CTRL+Z
}
