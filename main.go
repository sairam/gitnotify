package main

import "github.com/sairam/gitnotify/gitnotify"

func main() {
	gitnotify.LoadConfig()
	gitnotify.InitMailer()
	go gitnotify.InitCron()
	gitnotify.InitRouter()
}
