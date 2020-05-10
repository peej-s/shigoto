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
	Priority *PriorityValue
	Task     *TaskDescription
	UserID   string // This should be the user ID, not the username
	TaskID   string
}

type TaskUpdate struct {
	Priority *PriorityValue
	Task     *TaskDescription
}

// This can be used for both registration and login (for now)
// Maybe later we can have mandatory emails for password resets, but do it later since we do need to validate email
type User struct {
	Username string
	Password string
	UserID   string
	// Email    string
}

type AccessToken struct {
	Token  string
	UserID string
	Expiry time.Time
}

type CreateResponse struct {
	Success string
}

type UpdateResponse struct {
	Success string
	Updated int
}

type DeleteResponse struct {
	Success string
	Deleted int
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
