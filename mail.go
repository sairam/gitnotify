package main

// Mail helper methods
import (
	"fmt"
	"log"
	"time"

	"gopkg.in/gomail.v2"
)

// TODO load host smtp host information from config
var (
	emailCh = make(chan *gomail.Message)
)

func mailDaemon() {
	var s gomail.SendCloser
	var err error
	open := false

	d := gomail.NewDialer(config.SMTPHost, 587, config.SMTPUser, config.SMTPPass)
	d.LocalName = "localhost"
	for {
		select {
		case m, ok := <-emailCh:
			log.Println("attempting to send the email!")
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
			log.Println("done sending the email")
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

type recepient struct {
	Name    string
	Address string
}
type emailCtx struct {
	Subject  string
	HTMLBody string
	TextBody string
}

func sendEmail(to *recepient, e *emailCtx) {
	var from = &recepient{
		Name:    config.FromName,
		Address: config.FromEmail,
	}

	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(from.Address, from.Name))
	m.SetAddressHeader("To", to.Address, to.Name)
	m.SetHeader("Subject", e.Subject)
	m.SetBody("text/plain", fmt.Sprintf("Hi %s\n\n%s", to.Name, e.TextBody))
	m.AddAlternative("text/html", fmt.Sprintf("<pre style='font-size: 2em'>Hi %s<br/><br/>%s</pre>", to.Name, e.HTMLBody))

	emailCh <- m
}

// Use the channel in your program to send emails.
// TODO add halt when required
func stop() {
	// Close the channel to stop the mail daemon.
	close(emailCh)
}
