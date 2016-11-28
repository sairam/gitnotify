package main

// Mail helper methods
import (
	"fmt"
	"log"
	"os"
	"time"

	"gopkg.in/gomail.v2"
)

// TODO load host smtp host information from config
var (
	emailCh      = make(chan *gomail.Message)
	smtpHost     = "email-smtp.us-east-1.amazonaws.com"
	smtpUsername = os.Getenv("SMTP_USER")
	smtpPassword = os.Getenv("SMTP_PASS")
)

func mailDaemon() {
	var s gomail.SendCloser
	var err error
	open := false

	d := gomail.NewDialer(smtpHost, 587, smtpUsername, smtpPassword)
	d.LocalName = "localhost"
	for {
		select {
		case m, ok := <-emailCh:
			log.Println("starting to send an email!")
			if !ok {
				return
			}
			if !open {
				if s, err = d.Dial(); err != nil {
					log.Println("going to panic")
					panic(err)
				}
				open = true
			}
			if err := gomail.Send(s, m); err != nil {
				log.Print(err)
			}
			log.Println("done with email")
			// You should close the Amazon SES within 5 seconds of next request. else you it fails with 421.
		case <-time.After(4 * time.Second):
			if open {
				if err := s.Close(); err != nil {
					log.Println("going to panic. well. not really!")
					// panic(err)
				}
				open = false
			}
		}
	}
}

func init() {
	go mailDaemon()
}

type recepient struct {
	Name    string
	Address string
}
type emailCtx struct {
	Subject string
	Body    string
}

// TODO: modify from email address
var from = &recepient{
	Name:    "Git Notify",
	Address: "sairam.kunala@gmail.com",
}

func sendEmail(to *recepient, e *emailCtx) {

	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("%s <%s>", from.Name, from.Address))
	m.SetAddressHeader("To", to.Address, to.Name)
	m.SetHeader("Subject", e.Subject)
	m.SetBody("text/plain", e.Body)
	m.AddAlternative("text/html", "<pre>"+e.Body+"</pre>")

	emailCh <- m
}

// Use the channel in your program to send emails.
// TODO add halt when required
func stop() {
	// Close the channel to stop the mail daemon.
	close(emailCh)
}
