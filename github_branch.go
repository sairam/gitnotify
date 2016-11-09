package main

func (e *BranchInfo) String() string {
	return Stringify(e)
}
func (e *CommitRef) String() string {
	return Stringify(e)
}

type BranchInfo struct {
	Name   *string    `json:"name"`
	Commit *CommitRef `json:"commit"`
}
type CommitRef struct {
	Sha *string `json:"sha"`
	URL *string `json:"url"`
}
