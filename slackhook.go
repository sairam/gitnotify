package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// SlackMessage ..
type SlackMessage struct {
	Username    string            `json:"username"`
	Text        string            `json:"text"`
	Attachments []SlackAttachment `json:"attachments"`
}

// SlackAttachment ..
type SlackAttachment struct {
	Fallback       string                 `json:"fallback"`
	Title          string                 `json:"title"`
	Color          string                 `json:"color,omitempty"`
	PreText        string                 `json:"pretext"`
	AuthorName     string                 `json:"author_name"`
	AuthorLink     string                 `json:"author_link"`
	Fields         []SlackAttachmentField `json:"fields"`
	MarkdownFormat []string               `json:"mrkdwn_in"`
	Text           string                 `json:"text"`
	ThumbnailURL   string                 `json:"thumb_url,omitempty"`
}

// SlackAttachmentField ..
type SlackAttachmentField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type slackTypeLink struct {
	Text string
	Href string
}

// <http://www.amazon.com|Amazon>
func (s *slackTypeLink) String() string {
	return fmt.Sprintf("<%s|%s>", s.Href, s.Text)
}

func processForWebhook(diff []*repoDiffData, conf *Setting) error {
	if conf.User.isValidWebhook() {
		log.Print("POSTing on a Slack Hook")
		return processForSlack(diff, conf.User.WebhookURL)
	}
	return nil
}

func processForSlack(diffs []*repoDiffData, slackURL string) error {
	// loop and construct the slack message and send it
	for _, repo := range diffs {
		if repo.Changed == false {
			continue
		}
		attachments := make([]SlackAttachment, 1)
		for _, diff := range repo.Data {

			if diff.Changed == false {
				continue
			}
			if diff.ChangeType == "repoBranchDiff" && len(diff.Changes) > 0 {
				if diff.Error == "" {
					a := diff.Changes[0]
					attachment := SlackAttachment{
						Title:          (&slackTypeLink{diff.Title.Text, diff.Title.Href}).String(),
						Text:           (&slackTypeLink{a.Text, a.Href}).String(),
						MarkdownFormat: []string{"text"},
					}
					attachments = append(attachments, attachment)
				} else {
					attachment := SlackAttachment{
						Title:          diff.Title.Text,
						Text:           diff.Error,
						MarkdownFormat: []string{},
					}
					attachments = append(attachments, attachment)
				}

			} else {
				var links []string
				for _, change := range diff.Changes {
					links = append(links, (&slackTypeLink{change.Text, change.Href}).String())
				}

				attachment := SlackAttachment{
					Title:          fmt.Sprintf(diff.Title.Title),
					Text:           strings.Join(links, "\n"),
					MarkdownFormat: []string{"text"},
				}
				attachments = append(attachments, attachment)

			}
		}

		message := &SlackMessage{
			Username:    "gitnotify",
			Text:        fmt.Sprintf("*Changes for %s*:", &slackTypeLink{repo.Repo.Text, repo.Repo.Href}),
			Attachments: attachments,
		}

		pr, pw := io.Pipe()
		go func() {
			// close the writer, so the reader knows there's no more data
			defer pw.Close()

			// write json data to the PipeReader through the PipeWriter
			if err := json.NewEncoder(pw).Encode(message); err != nil {
				log.Print(err)
			}
		}()

		if _, err := http.Post(slackURL, "application/json", pr); err != nil {
			return err
		}
	}
	return nil
}
