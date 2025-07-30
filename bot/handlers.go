package bot

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"github.com/jyr94/telegram-dailytask-bot/firestore"
)

type BotHandler struct {
	Bot       *tgbotapi.BotAPI
	Firestore *firestore.FirestoreService
}

func (h *BotHandler) HandleUpdate(update tgbotapi.Update) {
	if update.Message == nil { // ignore non-message updates
		return
	}

	if update.Message.IsCommand() {
		switch update.Message.Command() {
		case "start":
			h.handleStart(update)
		case "add":
			h.handleAdd(update)
		case "list":
			h.handleListAll(update, "")
		case "done":
			args := update.Message.CommandArguments()
			h.handleDone(update, args)
		case "list_all":
			h.handleListAll(update, "all")
		case "tasks":
			args := update.Message.CommandArguments()
			h.handleTasksByDate(update, strings.TrimSpace(args))
		default:
			h.reply(update.Message.Chat.ID, "❓ Unknown command. Use /add, /list, or /done.")
		}
	}
}

func (h *BotHandler) handleAdd(update tgbotapi.Update) {
	args := strings.TrimSpace(update.Message.CommandArguments())
	if args == "" {
		h.reply(update.Message.Chat.ID, "⚠️ Usage: `/add Your task`")
		return
	}

	user := update.Message.From
	userID := strconv.FormatInt(user.ID, 10)

	task := firestore.Task{
		ID:       uuid.New().String(),
		Text:     args,
		Done:     false,
		Date:     time.Now().Format("2006-01-02"),
		Username: user.UserName,
		Created:  time.Now(),
	}

	err := h.Firestore.AddTask(userID, task)
	if err != nil {
		h.reply(update.Message.Chat.ID, "❌ Failed to add task.")
		log.Println("add task error:", err)
		return
	}

	h.reply(update.Message.Chat.ID, "✅ Task added: "+args)
}

func (h *BotHandler) handleDone(update tgbotapi.Update, args string) {
	chatID := update.Message.Chat.ID
	userID := fmt.Sprintf("%d", update.Message.From.ID)

	if args == "" {
		h.reply(chatID, "⚠️ Please provide task number or ID.\nExample: `/done 1` or `/done <task_id>`")
		return
	}

	index, convErr := strconv.Atoi(args)
	if convErr == nil {
		tasks, err := h.Firestore.GetTodayTasks(userID)
		if err != nil {
			h.reply(chatID, "❌ Failed to fetch your tasks.")
			return
		}

		if index < 1 || index > len(tasks) {
			h.reply(chatID, "⚠️ Invalid task number.")
			return
		}

		args = tasks[index-1].ID
	}

	err := h.Firestore.MarkTaskDone(userID, args)
	if err != nil {
		h.reply(chatID, "❌ Failed to mark task as done.")
		return
	}

	h.reply(chatID, "✅ Task marked as done!")
}

func (h *BotHandler) reply(chatID int64, message string) {
	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"
	_, err := h.Bot.Send(msg)
	if err != nil {
		log.Println("❌ Failed to send message:", err)
	}
}

func (h *BotHandler) handleStart(update tgbotapi.Update) {
	message := "👋 Hi! I can help you manage your daily tasks.\n\n" +
		"Use:\n" +
		"• `/add Your task here` to add a task\n" +
		"• `/list` to view tasks today\n" +
		"• `/list_all` to view tasks all date\n" +
		"• `/done 1` to mark task #1 as completed"

	h.reply(update.Message.Chat.ID, message)
}

func (h *BotHandler) handleListAll(update tgbotapi.Update, arg string) {
	chatID := update.Message.Chat.ID
	userID := fmt.Sprintf("%d", update.Message.From.ID)

	var tasks []firestore.Task
	var err error

	if arg == "all" {
		tasks, err = h.Firestore.GetAllTasks(userID)
	} else {
		tasks, err = h.Firestore.GetTodayTasks(userID)
	}

	if err != nil {
		h.reply(chatID, "❌ Failed to retrieve your tasks.")
		return
	}

	if len(tasks) == 0 {
		h.reply(chatID, "✅ No tasks found.")
		return
	}

	// Grouping by date
	grouped := make(map[string][]firestore.Task)
	for _, task := range tasks {
		grouped[task.Date] = append(grouped[task.Date], task)
	}

	var msg strings.Builder
	msg.WriteString("*📝 Your tasks:*\n")

	// Print by date (sorted optional)
	var dates []string
	for date := range grouped {
		dates = append(dates, date)
	}
	sort.Strings(dates)
	for _, date := range dates {
		msg.WriteString(fmt.Sprintf("\n🗓 *%s*\n", date))
		for i, task := range grouped[date] {
			status := "❌"
			if task.Done {
				status = "✅"
			}
			msg.WriteString(fmt.Sprintf("%d. %s %s \n", i+1, status, task.Text))
		}
	}

	h.reply(chatID, msg.String())
}

func (h *BotHandler) handleTasksByDate(update tgbotapi.Update, date string) {
	chatID := update.Message.Chat.ID
	userID := fmt.Sprintf("%d", update.Message.From.ID)

	_, err := time.Parse("2006-01-02", date)
	if err != nil {
		h.reply(chatID, "❌ Invalid date format. Please use YYYY-MM-DD (e.g., `/tasks 2025-08-01`).")
		return
	}

	tasks, err := h.Firestore.GetTasksByDate(userID, date)
	if err != nil {
		h.reply(chatID, "❌ Failed to retrieve tasks.")
		return
	}

	if len(tasks) == 0 {
		h.reply(chatID, fmt.Sprintf("✅ No tasks found on %s.", date))
		return
	}

	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("*📝 Tasks on %s:*\n", date))
	for i, task := range tasks {
		status := "❌"
		if task.Done {
			status = "✅"
		}
		msg.WriteString(fmt.Sprintf("%d. %s %s (`%s`)\n", i+1, status, task.Text, task.ID))
	}

	h.reply(chatID, msg.String())
}
