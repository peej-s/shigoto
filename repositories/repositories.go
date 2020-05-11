package repositories

import (
	"context"
	"fmt"
	"log"
	"shigoto/resources"
	u "shigoto/resources"
	"time"

	b "golang.org/x/crypto/bcrypt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Task Repository
type TaskRepository struct{}

func (r *TaskRepository) Create(task *u.TaskItem) *u.CreateResponse {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := u.DB.Collection("tasks")
	_, err := collection.InsertOne(ctx, task)
	if err != nil {
		log.Fatal(err)
	}
	return &u.CreateResponse{Success: task.TaskID}
}

func (r *TaskRepository) ReadByUserID(userID string) map[u.PriorityValue][]*u.TaskItem {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := u.DB.Collection("tasks")

	results := make(map[u.PriorityValue][]*u.TaskItem)

	filter := bson.M{"userid": userID}
	cur, err := collection.Find(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}

	for cur.Next(ctx) {
		var elem u.TaskItem
		var priority resources.PriorityValue
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}

		if elem.Priority == nil {
			priority = 0
		} else {
			priority = *elem.Priority
		}

		results[priority] = append(results[priority], &elem)
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	return results
}

func (r *TaskRepository) Update(userID string, taskID string, updates *u.TaskUpdate) *u.UpdateResponse {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := u.DB.Collection("tasks")
	filter := bson.M{
		"userid": userID,
		"taskid": taskID,
	}

	update := bson.M{
		"$set": updates,
	}
	updateResult, err := collection.UpdateOne(ctx, filter, update, nil)
	if err != nil {
		log.Fatal(err)
	}
	return &u.UpdateResponse{Success: fmt.Sprintf("%t", true), Updated: int(updateResult.MatchedCount)}
}

func (r *TaskRepository) Delete(userID string, taskID string) *u.DeleteResponse {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := u.DB.Collection("tasks")
	filter := bson.M{
		"userid": userID,
		"taskid": taskID,
	}
	deleteResult, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}
	return &u.DeleteResponse{Success: fmt.Sprintf("%t", true), Deleted: int(deleteResult.DeletedCount)}
}

// User Repository
type UserRepository struct{}

func (r *UserRepository) Create(newUser *u.User) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := u.DB.Collection("users")
	hashedPassword, err := b.GenerateFromPassword([]byte(newUser.Password), b.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	newUser.Password = string(hashedPassword)
	_, err = collection.InsertOne(ctx, newUser)
	if err != nil {
		log.Fatal(err)
	}
	return
}

func (r *UserRepository) ReadByUsername(username string) *u.User {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := u.DB.Collection("users")
	result := &u.User{}

	filter := bson.M{"username": username}
	err := collection.FindOne(ctx, filter).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return nil
	} else if err != nil {
		log.Fatal(err)
	}

	return result
}

// Token Repository
type TokenRepository struct{}

func (r *TokenRepository) Upsert(token *u.AccessToken) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := u.DB.Collection("tokens")
	opts := options.Update().SetUpsert(true)
	filter := bson.M{"userid": token.UserID}
	update := bson.M{
		"$set": bson.M{
			"token":  token.Token,
			"userid": token.UserID,
			"expiry": token.Expiry,
		},
	}
	_, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		log.Fatal(err)
	}
	return
}

func (r *TokenRepository) ReadByUserID(userID string) *u.AccessToken {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := u.DB.Collection("tokens")
	result := &u.AccessToken{}

	filter := bson.M{"userid": userID}
	err := collection.FindOne(ctx, filter).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return nil
	} else if err != nil {
		log.Fatal(err)
	}

	return result
}
