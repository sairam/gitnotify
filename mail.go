package main

// Mail helper methods
import (
	"fmt"
	"log"
	"time"

	"gopkg.in/gomail.v2"
)

var (
	emailCh = make(chan *gomail.Message)
)

func mailDaemon() {
	var s gomail.SendCloser
	var err error
	open := false

	d := gomail.NewDialer(config.SMTPHost, config.SMTPPort, config.SMTPUser, config.SMTPPass)
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
					log.Println("going to panic. ")
					log.Println(err)
					// panic: dial tcp: lookup email-smtp.us-east-1.amazonaws.com on 8.8.4.4:53: dial udp 8.8.4.4:53: i/o timeout
					// panic(err)
				} else {
					open = true
				}
			}
			if open {
				if err := gomail.Send(s, m); err != nil {
					log.Print(err)
				}
				log.Println("done sending the email")
			} else {
				log.Println("see above error. did not panic")
			}
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
	Name     string
	Address  string
	UserName string
	Provider string
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
	if config.SMTPSesConfSet != "" {
		m.SetHeader("X-SES-CONFIGURATION-SET", config.SMTPSesConfSet)
	}
	m.SetHeader("X-SES-MESSAGE-TAGS", fmt.Sprintf("%s=%s", to.Provider, to.UserName))
	m.SetHeader("List-ID", fmt.Sprintf("%s/%s <%s.%s.%s>", to.Provider, to.UserName, to.Provider, to.UserName, "gitnotify.com"))
	// m.SetHeader("List-Archive", fmt.Sprintf("")) // resource path like https://github.com/spf13/hugo
	m.SetHeader("List-Unsubscribe", fmt.Sprintf("<mailto:unsub+%s-%s@%s>, <%s>", to.Provider, to.UserName, config.ServerHost, config.ServerProto+config.ServerHost))

	m.SetBody("text/plain", e.TextBody)
	m.AddAlternative("text/html", e.HTMLBody)

	emailCh <- m
}

// Use the channel in your program to send emails.
// TODO add halt when required
func stop() {
	// Close the channel to stop the mail daemon.
	close(emailCh)
}
