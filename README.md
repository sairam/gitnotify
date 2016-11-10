## Git(hub|lab) Release/Branch Version watch

### Project Name
* gitnotify

### Development Flow
1. Users login via Github
1. Users keep "an eye" on a project
1. Notifies users on
  1. creation of new branches/tags.
  1. Track a branch for latest commits and links
1. Send email with summary

### User Flow
1. Landing Page at / or /home
1. User logins via Github to authenticate/track
1. User adds url of repository to track creation of branches/tags and/or branches for latest commits
1. User receives a confirmation email with the current list of branches
1. Users receives an email once a day about the information that has changed

### Backend Flow
1. A non logged in user sees the static page present at /home at /
1. User logs in via Github
1. Settings page has a set of repositories he is tracking
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
