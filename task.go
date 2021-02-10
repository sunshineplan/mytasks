package main

import (
	"context"
	"errors"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type task struct {
	ID       string             `json:"id"`
	ObjectID primitive.ObjectID `json:"-" bson:"_id"`
	Task     string             `json:"task"`
	List     string             `json:"list"`
	Created  time.Time          `json:"created"`
	Seq      int                `json:"-"`
}

func deleteTask(objectID primitive.ObjectID, userID string, completed bool) error {
	var collection *mongo.Collection
	if completed {
		collection = collCompleted
	} else {
		collection = collIncomplete
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var task task
	if err := collection.FindOneAndDelete(ctx, bson.M{"_id": objectID}).Decode(&task); err != nil {
		return err
	}

	if !completed {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if _, err := collection.UpdateMany(ctx,
			bson.M{"user": userID, "list": task.List, "seq": bson.M{"$gt": task.Seq}},
			bson.M{"$inc": bson.M{"seq": -1}},
		); err != nil {
			return err
		}
	}

	return nil
}

func checkTask(objecdID primitive.ObjectID, userID interface{}, completed bool) bool {
	var collection *mongo.Collection
	if completed {
		collection = collCompleted
	} else {
		collection = collIncomplete
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := collection.FindOne(ctx, bson.M{"_id": objecdID, "user": userID}).Err(); err != nil {
		if err != mongo.ErrNoDocuments {
			log.Print(err)
		}
		return false
	}

	return true
}

func getTask(list, userID string, completed bool) ([]task, error) {
	var collection *mongo.Collection
	var opts *options.FindOptions
	if completed {
		collection = collCompleted
		opts = options.Find().SetSort(bson.M{"created": 1}).SetLimit(10)
	} else {
		collection = collIncomplete
		opts = options.Find().SetSort(bson.M{"seq": 1})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{"list": list, "user": userID}, opts)
	if err != nil {
		log.Println("Failed to query tasks:", err)
		return nil, err
	}

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tasks := []task{}
	if err = cursor.All(ctx, &tasks); err != nil {
		log.Println("Failed to get tasks:", err)
		return nil, err
	}
	for i := range tasks {
		tasks[i].ID = tasks[i].ObjectID.Hex()
	}

	return tasks, nil
}

func addTask(t task, userID string, completed bool) (string, error) {
	document := bson.D{
		{Key: "task", Value: t.Task},
		{Key: "list", Value: t.List},
		{Key: "created", Value: time.Now()},
	}

	var collection *mongo.Collection
	var seq int
	if completed {
		collection = collCompleted
	} else {
		collection = collIncomplete

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cursor, err := collection.Find(
			ctx, bson.M{"user": userID, "list": t.List}, options.Find().SetSort(bson.M{"seq": -1}).SetLimit(1))
		if err != nil {
			return "", err
		}

		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var tasks []task
		if err := cursor.All(ctx, &tasks); err != nil {
			return "", err
		}

		if len(tasks) == 0 {
			seq = 1
		} else {
			seq = tasks[0].Seq + 1
		}

		document = append(document, bson.E{Key: "seq", Value: seq})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := collection.InsertOne(ctx, document)
	if err != nil {
		return "", err
	}

	insertID, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", err
	}

	return insertID.Hex(), nil
}

func reorderTask(userID, list string, orig, dest primitive.ObjectID) error {
	var origTask, destTask task

	c := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		c <- collIncomplete.FindOne(ctx, bson.M{"_id": orig}).Decode(&origTask)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := collIncomplete.FindOne(ctx, bson.M{"_id": dest}).Decode(&destTask); err != nil {
		return err
	}
	if err := <-c; err != nil {
		return err
	}

	if origTask.List != list || destTask.List != list {
		return errors.New("List not match")
	}

	var filter, update bson.M
	if origTask.Seq > destTask.Seq {
		filter = bson.M{"user": userID, "list": list, "seq": bson.M{"$gte": destTask.Seq, "$lt": origTask.Seq}}
		update = bson.M{"$inc": bson.M{"seq": 1}}
	} else {
		filter = bson.M{"user": userID, "list": list, "seq": bson.M{"$gt": origTask.Seq, "$lte": destTask.Seq}}
		update = bson.M{"$inc": bson.M{"seq": -1}}
	}

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := collIncomplete.UpdateMany(ctx, filter, update); err != nil {
		log.Println("Failed to reorder tasks:", err)
		return err
	}

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := collIncomplete.UpdateOne(
		ctx, bson.M{"_id": orig}, bson.M{"$set": bson.M{"seq": destTask.Seq}}); err != nil {
		log.Println("Failed to reorder task:", err)
		return err
	}

	return nil
}
