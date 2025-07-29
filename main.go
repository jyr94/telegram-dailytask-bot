package main

import (
	"log"

	"github.com/jyr94/telegram-dailytask-bot/config"

	"github.com/jyr94/telegram-dailytask-bot/bot"
)

func main() {
	cfg := config.Load()

	log.Println("🤖 Starting Daily Task Bot...")
	bot.Start(cfg)
}
