package main

import (
	"github.com/Seician/bookings/internal/models"
	mail "github.com/xhit/go-simple-mail/v2"
	"log"
	"time"
)

func listenForEmail() {
	go func() {
		for {
			message := <-app.MailChan
			sendMessage(message)
		}
	}()

}

func sendMessage(m models.MailData) {
	server := mail.NewSMTPClient()
	server.Host = "localhost"
	server.Port = 1025
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	client, err := server.Connect()
	if err != nil {
		errorLog.Println(err)
	}
	email := mail.NewMSG()
	email.SetFrom(m.From).AddTo(m.To).SetSubject(m.Subject)
	email.SetBody(mail.TextHTML, m.Content)

	err = email.Send(client)
	if err != nil {
		log.Print(err)
	} else {
		log.Print("Email sent!")
	}
}
