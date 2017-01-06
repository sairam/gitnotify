package main

import "./gitnotify"

func main() {
	gitnotify.LoadConfig()
	gitnotify.InitMailer()
	go gitnotify.InitCron()
	gitnotify.InitRouter()
}
