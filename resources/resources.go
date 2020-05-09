package resources

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// App-Specific Structs
var DB *mongo.Database

type TaskItem struct {
	Priority int
	Task     string
	UserID   string // This should be the user ID, not the username
	TaskID   string
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

func InitializeResources() {
	fmt.Println("Program Starting")

	var MongoURI string = os.Getenv("SHIGOTO_MDB_STRING")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	session, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoURI))
	if err != nil {
		log.Fatal(err)
	}

	DB = session.Database("shigoto")
	fmt.Println("Server Ready For Action")
}
