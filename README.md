[![Build Status](https://travis-ci.org/sairam/gitnotify.svg?branch=master)](https://travis-ci.org/sairam/gitnotify)
[![Code Climate](https://codeclimate.com/github/sairam/gitnotify/badges/gpa.svg)](https://codeclimate.com/github/sairam/gitnotify)
[![Issue Count](https://codeclimate.com/github/sairam/gitnotify/badges/issue_count.svg)](https://codeclimate.com/github/sairam/gitnotify)
[![codecov](https://codecov.io/gh/sairam/gitnotify/branch/master/graph/badge.svg)](https://codecov.io/gh/sairam/gitnotify)


# gitnotify
## Github and Gitlab Release/Branch Version Watcher
Get periodic emails about the code diff for Gitlab and Github repositories

## How to Setup
### Fetch Dependencies
```bash
go get -u golang.org/x/oauth2
go get -u gopkg.in/gomail.v2
go get -u gopkg.in/robfig/cron.v2 # requires modification to file aka bugfix
go get -u gopkg.in/yaml.v2
go get -u github.com/google/go-github/github
go get -u github.com/gorilla/mux
go get -u github.com/gorilla/sessions
go get -u github.com/markbates/goth
go get -u github.com/sairam/timezone
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

## FAQ
### Can I run this inside my own organisation
Only the Configuration needs to be setup.

## Disclaimer
I started learning Go (~Sep 2016) and this is my first moderate sized project trying to "solve" a problem

## Flow of user
### Development Flow
1. Users login via Github
1. Users keep "an eye" on a project
1. Notifies users on
  1. creation of new branches/tags
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
1. User logs in via Github/Gitlab
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


## TODO
Get an update on demand from slack - http://www.hongkiat.com/blog/custom-slash-command-slack/ 
