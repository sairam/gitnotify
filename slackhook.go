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

func processForWebhook(diff []*gitRepoDiffs, conf *Setting) error {
	if conf.User.WebhookType == "slack" && conf.User.WebhookURL != "" {
		log.Println("sending a slack message")
		return processForSlack(diff, conf.User.WebhookURL)
	}
	return nil
}

// TODO: we are repeating the logic too much. clean it up from text, html emails and this.
// Make a properly formatted struct / json
func processForSlack(diff []*gitRepoDiffs, slackURL string) error {
	// loop and construct the slack message and send it
	for _, repo := range diff {
		attachments := make([]SlackAttachment, 1)
		for branch, reference := range repo.References {
			var commitDiffLink string
			if reference.NewCommit == noneString {
				commitDiffLink = fmt.Sprintf("Branch is not present")
			} else if reference.OldCommit == "" {
				commitDiffLink = fmt.Sprintf("New branch being tracked. Current Commit is %s", &slackTypeLink{shortCommit(reference.NewCommit), TreeLink(repo.Provider, repo.RepoName, reference.NewCommit)})
			} else if reference.OldCommit == reference.NewCommit {
				commitDiffLink = fmt.Sprintf("No recent changes. Last Commit is %s", &slackTypeLink{shortCommit(reference.NewCommit), TreeLink(repo.Provider, repo.RepoName, reference.NewCommit)})
			} else {
				text := fmt.Sprintf("%s..%s", shortCommit(reference.OldCommit), shortCommit(reference.NewCommit))
				href := CompareLink(repo.Provider, repo.RepoName, reference.OldCommit, reference.NewCommit)
				commitDiffLink = fmt.Sprintf("Compare Diff: %s", (&slackTypeLink{text, href}).String())
			}
			branchTitle := fmt.Sprintf("Branch: %s", branch)

			attachment := SlackAttachment{
				Title:          (&slackTypeLink{branchTitle, TreeLink(repo.Provider, repo.RepoName, branch)}).String(),
				MarkdownFormat: []string{"text"},
				Text:           commitDiffLink,
			}
			attachments = append(attachments, attachment)
		}

		for _, r := range repo.Others {
			links := make([]string, 0, 1)
			for _, treeName := range r.References {
				href := TreeLink(repo.Provider, repo.RepoName, treeName)
				links = append(links, (&slackTypeLink{treeName, href}).String())
			}
			if len(r.References) == 0 {
				links = append(links, "None")
			}
			attachment := SlackAttachment{
				Title:          fmt.Sprintf("New %s:", r.Title),
				Text:           strings.Join(links, "\n"),
				MarkdownFormat: []string{"text"},
			}
			attachments = append(attachments, attachment)
		}

		message := &SlackMessage{
			Username:    "gitnotify",
			Text:        fmt.Sprintf("*Changes for %s*:", &slackTypeLink{repo.RepoName, RepoLink(repo.Provider, repo.RepoName)}),
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
