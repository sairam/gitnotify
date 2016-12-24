package main

type localGitnull struct {
	provider string
}

func (*localGitnull) WebsiteLink() string {
	return ""
}
func (*localGitnull) RepoLink(string) string {
	return ""
}
func (*localGitnull) TreeLink(_, _ string) string {
	return ""
}
func (*localGitnull) CommitLink(_, _ string) string {
	return ""
}
func (*localGitnull) CompareLink(_, _, _ string) string {
	return ""
}

func (g *localGitnull) Branches(_ string) ([]*GitRefWithCommit, error) {
	return nil, &providerNotPresent{g.provider}
}
func (g *localGitnull) Tags(_ string) ([]*GitRefWithCommit, error) {
	return nil, &providerNotPresent{g.provider}
}
func (g *localGitnull) SearchRepos(_ string) ([]*searchRepoItem, error) {
	return []*searchRepoItem{}, &providerNotPresent{g.provider}
}
func (g *localGitnull) SearchUsers(_ string) ([]*searchRepoItem, error) {
	return []*searchRepoItem{}, &providerNotPresent{g.provider}
}
func (g *localGitnull) DefaultBranch(_ string) (string, error) {
	return "", &providerNotPresent{g.provider}
}
func (g *localGitnull) BranchesWithoutRefs(_ string) ([]string, error) {
	return []string{}, &providerNotPresent{g.provider}
}
func (g *localGitnull) RemoteOrgType(_ string) (string, error) {
	return "", &providerNotPresent{g.provider}
}
