package main

// https://docs.gitlab.com/ce/api/branches.html
// https://gitlab.com/api/v3/projects/2020683/repository/branches

// https://docs.gitlab.com/ce/api/tags.html
// https://gitlab.com/api/v3/projects/2020683/repository/tags
// :proto://:hostname/api/v3/projects/:id/repository/tags
/*
Example:
  [{
    "name":"master",
    "commit":{
      "id":"c36c69c0613a359a41fe5da8e70047bffe7f97c2"
    }
  }]
*/
//

// // Helpers
// func gitlabRepoLink(repo string) string {
// 	return fmt.Sprintf(gitlabRepoEndPoint, repo)
// }
//
// func gitlabTreeLink(repo, ref string) string {
// 	return fmt.Sprintf(gitlabTreeURLEndPoint, repo, ref)
// }
//
// func gitlabCommitLink(repo, ref string) string {
// 	return fmt.Sprintf(gitlabCommitURLEndPoint, repo, ref)
// }
//
// func gitlabCompareLink(repo, oldCommit, newCommit string) string {
// 	return fmt.Sprintf(gitlabCompareURLEndPoint, repo, oldCommit, newCommit)
// }
