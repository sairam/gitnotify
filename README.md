# gitnotify
## Github Release/Branch Version Watcher

## How to Setup
### Fetch Dependencies
```bash
go get -u golang.org/x/oauth2
go get -u gopkg.in/gomail.v2
go get -u gopkg.in/robfig/cron.v2 # requires modification to file aka bugfix
go get -u gopkg.in/yaml.v2
go get -u github.com/gorilla/mux
go get -u github.com/markbates/goth
go get -u github.com/google/go-github/github
go get -u github.com/gorilla/sessions
go get -u github.com/sairam/timezone # added a method in the original code
go get -u github.com/spf13/cast
```

### Modify package
* `gopkg.in/robfig/cron.v2`
* Edit line 102
* Replace `t = t.Truncate(time.Hour)` with `t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, s.Location)`

### How to build for Linux
* `env GOOS=linux GOARCH=amd64 go build`

### Runtime Dependencies
* `cp .env.sample .env.prod`
* `edit config.yml`
* `ls tmpl/`

### Running on a Linux/Mac environment
1. `mkdir sessions`
1. Fill in the `.env.prod` file and modify config.yml
1. Load the env variables from `.env.prod`
1. Use the binary `gitnotify` and `tmpl/` directory
1. Start with `./gitnotify` in a screen. All logs are currently written to stdout

### Backup
1. take a copy of the `config.yml` -> `dataDir` or `data/` directory
1. take a copy of the `sessions/` directory
1. The `.env.prod` file containing the environment variables
1. The `config.yml` file containing the settings

## TODO
1. Add text output for email
1. Host on separate instance (install Caddy w/ https and configure server)
1. Validate repo name from server side and autofill default branch name
1. Add LICENSE
1. Log when adding a repository from the UI fails

### Known Knowns
1. Clean up fetched_info from `settings.yml` file when a repository is removed
1. Fix manual editing of cron.v2 since it does not work on +0530 like TZs

### Nice to Have
1. JSON output to be used for sending information as webhooks to Zapier like services
1. Allow autofill of branch names based on repo names
1. Suggest names of popular repositories to ease adding first few repositories
1. Add support for Gitlab

## Flow of user
### Development Flow
1. Users login via Github
1. Users keep "an eye" on a project
1. Notifies users on
  1. creation of new branches/tags.
  1. Track a branch for latest commits and links
1. Send email with summary

### User Flow
1. Landing Page at /home
1. User logins via Github to authenticate/track
1. User adds url of repository to track creation of branches/tags and/or branches for latest commits
1. User receives a confirmation email with the current list of branches
1. Users receives an email once a day about the information that has changed

### Backend Flow
1. A non logged in user sees the static page present at /home at /
1. User logs in via Github
1. Settings page has a set of repositories s/he is tracking
1. Suggestions based on code language can be pulled for user to track
1. Once user adds the config, we save it to the settings
1. After validating the config, an email is sent to the user with the action made
1. A cron job or daemon uses the user's token to pull the required information
1. Diffs with the previous information that is present
1. Sends an email to the user with changes based on settings

## Configuration
* Copy `.env.example` to `.env`
* Fill in data into `.env`
* environment variables required to be loaded for the server to run
