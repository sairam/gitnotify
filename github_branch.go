package main

func (e *BranchInfo) String() string {
	return Stringify(e)
}
func (e *CommitRef) String() string {
	return Stringify(e)
}

// In `Tag` ignore zipball_url, tarball_url
type TagInfo struct {
	Name   string     `json:"name" yaml:"name"`
	Commit *CommitRef `json:"commit" yaml:"commit"`
}

type BranchInfo struct {
	Name   string     `json:"name" yaml:"name"`
	Commit *CommitRef `json:"commit" yaml:"commit"`
}

type CommitRef struct {
	Sha string `json:"sha" yaml:"sha"`
	URL string `json:"url" yaml:"url"`
}
