package bot

import (
	"log"
	"strconv"

	"github.com/jyr94/telegram-dailytask-bot/config"
	"github.com/jyr94/telegram-dailytask-bot/firestore"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func Start(cfg config.Config) {
	// Initialize Telegram bot
	bot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		log.Fatalf("❌ Failed to create Telegram bot: %v", err)
	}

	bot.Debug = false
	log.Printf("🤖 Bot is running as @%s", bot.Self.UserName)

	// Initialize Firestore
	fs := firestore.NewService(cfg.CredentialsJSON, cfg.FirebaseProjectId)

	// Create handler
	handler := &BotHandler{
		Bot:       bot,
		Firestore: fs,
	}

	// Optional: start reminder scheduler
	// handler.StartReminderScheduler(20)

	// Start polling updates
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil && update.Message.From != nil {
			// Save user into Firestore (optional auto track)
			userID := strconv.FormatInt(update.Message.From.ID, 10)
			fs.EnsureUserExists(userID, update.Message.From.UserName)
		}

		go handler.HandleUpdate(update)
	}
}
