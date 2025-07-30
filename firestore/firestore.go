package firestore

import (
	"context"
	"log"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

type Task struct {
	ID       string    `firestore:"id"`
	Text     string    `firestore:"text"`
	Done     bool      `firestore:"done"`
	Date     string    `firestore:"date"` // YYYY-MM-DD
	Username string    `firestore:"username"`
	Created  time.Time `firestore:"created"`
}

type FirestoreService struct {
	client *firestore.Client
	ctx    context.Context
}

func NewService(credentialsFile string, projectId string) *FirestoreService {
	ctx := context.Background()

	client, err := firestore.NewClient(ctx, projectId, option.WithCredentialsFile(credentialsFile), option.WithEndpoint("firestore.googleapis.com:443"))
	if err != nil {
		log.Fatalf("❌ Failed to create Firestore client: %v", err)
	}

	return &FirestoreService{
		client: client,
		ctx:    ctx,
	}
}

func (f *FirestoreService) AddTask(userID string, task Task) error {
	docRef := f.client.Collection("users").Doc(userID)

	_, err := docRef.Set(f.ctx, map[string]interface{}{
		"tasks": firestore.ArrayUnion(task),
	}, firestore.MergeAll)

	return err
}

func (f *FirestoreService) ListTasks(userID string) ([]Task, error) {
	var tasks []Task

	docs, err := f.client.Collection("users").
		Doc(userID).
		Collection("tasks").
		OrderBy("created", firestore.Asc).
		Documents(f.ctx).
		GetAll()

	if err != nil {
		return nil, err
	}

	for _, doc := range docs {
		var t Task
		if err := doc.DataTo(&t); err != nil {
			continue
		}
		tasks = append(tasks, t)
	}

	return tasks, nil
}

func (fs *FirestoreService) MarkTaskDone(userID, taskID string) error {
	docRef := fs.client.Collection("users").Doc(userID)
	doc, err := docRef.Get(fs.ctx)
	if err != nil {
		return err
	}

	data := doc.Data()
	rawTasks, ok := data["tasks"].([]interface{})
	if !ok {
		return nil
	}

	var updatedTasks []interface{}
	for _, raw := range rawTasks {
		taskMap := raw.(map[string]interface{})
		if taskMap["id"] == taskID {
			taskMap["done"] = true
		}
		updatedTasks = append(updatedTasks, taskMap)
	}

	_, err = docRef.Set(fs.ctx, map[string]interface{}{
		"tasks": updatedTasks,
	}, firestore.MergeAll)

	return err
}

func (f *FirestoreService) EnsureUserExists(userID, username string) {
	doc := f.client.Collection("users").Doc(userID)
	_, err := doc.Get(f.ctx)
	if err != nil {
		// If user doesn't exist, create it
		_, err := doc.Set(f.ctx, map[string]interface{}{
			"username": username,
			"created":  time.Now(),
		})
		if err != nil {
			log.Printf("⚠️ Failed to create user %s: %v", userID, err)
		}
	}
}

func (fs *FirestoreService) GetTodayTasks(userID string) ([]Task, error) {
	doc, err := fs.client.Collection("users").Doc(userID).Get(fs.ctx)
	if err != nil {
		return nil, err
	}

	data := doc.Data()
	rawTasks, ok := data["tasks"].([]interface{})
	if !ok {
		return []Task{}, nil
	}

	today := time.Now().Format("2006-01-02")
	var tasks []Task

	for _, raw := range rawTasks {
		taskMap, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}

		date, _ := taskMap["date"].(string)
		if date != today {
			continue
		}

		task := Task{
			ID:       taskMap["id"].(string),
			Text:     taskMap["text"].(string),
			Done:     taskMap["done"].(bool),
			Date:     date,
			Username: taskMap["username"].(string),
			// created: ignore for now
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (fs *FirestoreService) GetAllTasks(userID string) ([]Task, error) {
	doc, err := fs.client.Collection("users").Doc(userID).Get(fs.ctx)
	if err != nil {
		return nil, err
	}

	data := doc.Data()
	rawTasks, ok := data["tasks"].([]interface{})
	if !ok {
		return []Task{}, nil
	}

	var tasks []Task
	for _, raw := range rawTasks {
		taskMap, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}

		task := Task{
			ID:       taskMap["id"].(string),
			Text:     taskMap["text"].(string),
			Done:     taskMap["done"].(bool),
			Date:     taskMap["date"].(string),
			Username: taskMap["username"].(string),
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (fs *FirestoreService) GetTasksByDate(userID, date string) ([]Task, error) {
	doc, err := fs.client.Collection("users").Doc(userID).Get(fs.ctx)
	if err != nil {
		return nil, err
	}

	data := doc.Data()
	rawTasks, ok := data["tasks"].([]interface{})
	if !ok {
		return []Task{}, nil
	}

	var tasks []Task
	for _, raw := range rawTasks {
		taskMap, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}

		taskDate, ok := taskMap["date"].(string)
		if !ok || taskDate != date {
			continue
		}

		task := Task{
			ID:       taskMap["id"].(string),
			Text:     taskMap["text"].(string),
			Done:     taskMap["done"].(bool),
			Date:     taskDate,
			Username: taskMap["username"].(string),
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}
