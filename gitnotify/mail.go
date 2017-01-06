package gitnotify

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/mail"
	"strings"
	"time"

	"github.com/sairam/kinli"
)

// MailContent ..
type MailContent struct {
	WebsiteURL string
	User       string // provider/username
	Name       string
	Data       []*gnDiffData
	SavedFile  string
}

// InitMail initialises mailer if config is setup
func InitMail() {
	if config.isEmailSetup() {
		var smtpConfig = &kinli.EmailSMTPConfig{
			Host: config.SMTPHost,
			Port: config.SMTPPort,
			User: config.SMTPUser,
			Pass: config.SMTPPass,
		}
		kinli.InitMailer(smtpConfig)
	} else {
		log.Println("email is not configured")
	}
}

func processForMail(diff gnDiffDatum, conf *Setting, fileName string) error {
	if config.isEmailSetup() == false || !isValidEmail(conf.usersEmail()) {
		return nil
	}

	mailContent := &MailContent{
		WebsiteURL: config.websiteURL(),
		User:       fmt.Sprintf("%s/%s", conf.Auth.Provider, conf.Auth.UserName),
		Name:       conf.usersName(),
		Data:       diff,
		SavedFile:  fileName,
	}

	htmlBuffer := &bytes.Buffer{}
	kinli.DisplayPage(htmlBuffer, "changes_mail", mailContent)
	html, _ := ioutil.ReadAll(htmlBuffer)

	textBuffer := &bytes.Buffer{}
	kinli.DisplayPage(textBuffer, "changes_mail_text", mailContent)
	text, _ := ioutil.ReadAll(textBuffer)
	plain := strings.Replace(string(text), "\n\n", "\n", -1)
	plain = strings.Replace(plain, "\n\n", "\n", -1)

	loc, _ := time.LoadLocation(conf.User.TimeZoneName)
	t := time.Now().In(loc)
	subject := "[GitNotify] New Updates from your Repositories - " + t.Format("02 Jan 2006 | 15 Hrs")

	fromEmail := &mail.Address{
		Name:    config.FromName,
		Address: config.FromEmail,
	}

	toEmail := &mail.Address{
		Name:    conf.usersName(),
		Address: conf.usersEmail(),
	}

	headers := make(map[string]string)
	if config.SMTPSesConfSet != "" {
		headers["X-SES-CONFIGURATION-SET"] = config.SMTPSesConfSet
	}
	headers["X-SES-MESSAGE-TAGS"] = fmt.Sprintf("%s=%s", conf.Auth.Provider, conf.Auth.UserName)
	// TODO - change constant gitnotify.com to config value
	headers["List-ID"] = fmt.Sprintf("%s/%s <%s.%s.%s>",
		conf.Auth.Provider, conf.Auth.UserName, conf.Auth.Provider, conf.Auth.UserName, "gitnotify.com")
	// m.SetHeader("List-Archive", fmt.Sprintf("")) // resource path like https://github.com/spf13/hugo
	headers["List-Unsubscribe"] = fmt.Sprintf("<mailto:unsub+%s-%s@%s>, <%s>",
		conf.Auth.Provider, conf.Auth.UserName, config.ServerHost, config.ServerProto+config.ServerHost)

	e := &kinli.EmailCtx{
		From:      fromEmail,
		To:        []*mail.Address{toEmail},
		Subject:   subject,
		PlainBody: plain,
		HTMLBody:  string(html),
		Headers:   headers,
	}

	e.SendEmail()
	return nil
}
