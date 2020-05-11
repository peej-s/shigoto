package resources

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// App-Specific Structs
var DB *mongo.Database

type PriorityValue int
type TaskDescription string

type TaskItem struct {
	Priority *PriorityValue   `json:"priority"`
	Task     *TaskDescription `json:"task"`
	UserID   string           `json:"userid"`
	TaskID   string           `json:"taskid"`
}

type TaskUpdate struct {
	Priority *PriorityValue   `json:"priority"`
	Task     *TaskDescription `json:"task"`
}

// This can be used for both registration and login (for now)
// Maybe later we can have mandatory emails for password resets, but do it later since we do need to validate email
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	UserID   string `json:"userid"`
	// Email    string
}

type AccessToken struct {
	Token  string    `json:"token"`
	UserID string    `json:"userid"`
	Expiry time.Time `json:"expiry"`
}

type CreateResponse struct {
	Success string `json:"success"`
}

type UpdateResponse struct {
	Success string `json:"success"`
	Updated int    `json:"updated"`
}

type DeleteResponse struct {
	Success string `json:"success"`
	Deleted int    `json:"deleted"`
}

func InitializeResources() {
	var MongoURI string = os.Getenv("SHIGOTO_MDB_STRING")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	session, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoURI))
	if err != nil {
		log.Fatal(err)
	}

	DB = session.Database("shigoto")
}
