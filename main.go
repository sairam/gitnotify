package main

import "github.com/sairam/gitnotify/gitnotify"

func main() {
	gitnotify.LoadConfig("config.yml")
	gitnotify.InitMail()
	go gitnotify.InitCron()
	gitnotify.InitRouter()
}
