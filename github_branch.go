package main

import (
	"fmt"
	"strings"
)

/*
Example:
  [{
    "name": "1-2-stable",
    "commit": {
      "sha": "5b3f7563ae1b4a7160fda7fe34240d40c5777dcd",
      "url": "https://api.github.com/repos/rails/rails/commits/5b3f7563ae1b4a7160fda7fe34240d40c5777dcd"
    }
  }]
*/
// TagInfo .
// In `Tag` ignore zipball_url, tarball_url
type TagInfo struct {
	Name   string     `json:"name" yaml:"name"`
	Commit *CommitRef `json:"commit" yaml:"commit"`
}

func (e *TagInfo) String() string {
	return Stringify(e)
}

// CommitRef
type CommitRef struct {
	Sha string `json:"sha" yaml:"sha"`
	URL string `json:"url" yaml:"url"`
}

func (e *CommitRef) String() string {
	return Stringify(e)
}

// DifferView is implemented by LocalRef and BranchDiff
type DifferView interface {
	toText() string
	toHTML() string
}

// LocalDiff has multiple references to DifferView interface
type LocalDiff struct {
	refs []DifferView
}

func (ld *LocalDiff) add(l DifferView) {
	ld.refs = append(ld.refs, l)
}

func (ld *LocalDiff) toHTML() string {
	s := make([]string, len(ld.refs))

	for i, ref := range ld.refs {
		s[i] = ref.toHTML()
	}
	return strings.Join(s, "<br /><hr/>")
}

func (ld *LocalDiff) toText() string {
	s := make([]string, len(ld.refs))

	for i, ref := range ld.refs {
		s[i] = ref.toText()
	}
	delim := strings.Repeat("-", 80)
	return strings.Join(s, "\n\n"+delim+"\n\n")
}

type BranchDiff struct {
	Repo       string
	Title      string
	References []string // old and new
}

func (b *BranchDiff) toText() string {
	return "text"
}
func (b *BranchDiff) toHTML() string {
	return "html"
}

// LocalRef is used tracking Repo and Branch from the email
type LocalRef struct {
	Repo       string
	Title      string
	References []string
}

func (l *LocalRef) urlLink() string {
	return fmt.Sprintf(githubURLEndPoint, l.Repo)
}

func (l *LocalRef) treeLink(ref string) string {
	return fmt.Sprintf(githubTreeURLEndPoint, l.Repo, ref)
}
func (l *LocalRef) commitLink(ref string) string {
	return fmt.Sprintf(githubCommitURLEndPoint, l.Repo, ref)
}

func (l *LocalRef) toHTML() string {
	s := make([]string, len(l.References)+1)
	if len(l.References) != 0 {
		s[0] = fmt.Sprintf("Fetched <strong>%d</strong> recently created <strong>%s</strong> for <a href=\"%s\">%s</a>", len(l.References), l.Title, l.urlLink(), l.Repo)
	} else {
		s[0] = fmt.Sprintf("No new %s were found for <a href=\"%s\">%s</a>", l.Title, l.urlLink(), l.Repo)
	}
	for i, ref := range l.References {
		s[i+1] = fmt.Sprintf("<a href=\"%s\">%s::%s</a> <a href=\"%s\">[commits]</a>", l.treeLink(ref), l.Repo, ref, l.commitLink(ref))
	}
	return strings.Join(s, "<br />")
}

func (l *LocalRef) toText() string {
	s := make([]string, len(l.References)+1)
	if len(l.References) != 0 {
		s[0] = fmt.Sprintf("Fetched *%d recently created %s*: for %s(%s)", len(l.References), l.Title, l.Repo, l.urlLink())
	} else {
		s[0] = fmt.Sprintf("No new %s were found for %s(%s)", l.Title, l.urlLink(), l.Repo)
	}
	for i, ref := range l.References {
		s[i+1] = fmt.Sprintf("%s - %s | Commits: %s", ref, l.treeLink(ref), l.commitLink(ref))
	}
	return strings.Join(s, "\n")
}
