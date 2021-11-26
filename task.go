package main

import (
	"errors"
	"time"

	"github.com/sunshineplan/database/mongodb"
)

type task struct {
	ID       string    `json:"id"`
	ObjectID string    `json:"_id,omitempty" bson:"_id"`
	Task     string    `json:"task"`
	List     string    `json:"list"`
	Created  time.Time `json:"created"`
	Seq      int       `json:"seq,omitempty"`
}

func checkTask(id mongodb.ObjectID, userID interface{}, completed bool) bool {
	var client mongodb.Client
	if completed {
		client = completedClient
	} else {
		client = incompleteClient
	}

	n, _ := client.CountDocuments(mongodb.M{"_id": id.Interface(), "user": userID}, nil)
	return n > 0
}

func getTask(list, userID string, completed bool) ([]task, error) {
	var client mongodb.Client
	var opt *mongodb.FindOpt
	if completed {
		client = completedClient
		opt = &mongodb.FindOpt{Sort: mongodb.M{"created": -1}, Limit: 10}
	} else {
		client = incompleteClient
		opt = &mongodb.FindOpt{Sort: mongodb.M{"seq": -1}}
	}

	tasks := []task{}
	if err := client.Find(mongodb.M{"list": list, "user": userID}, opt, &tasks); err != nil {
		return nil, err
	}
	for i := range tasks {
		tasks[i].ID = tasks[i].ObjectID
		tasks[i].ObjectID = ""
		tasks[i].Seq = 0
	}

	return tasks, nil
}

func addTask(t task, userID string, completed bool) (interface{}, error) {
	doc := struct {
		Task    string      `json:"task" bson:"task"`
		List    string      `json:"list" bson:"list"`
		User    string      `json:"user" bson:"user"`
		Created interface{} `json:"created" bson:"created"`
		Seq     int         `json:"seq,omitempty" bson:"seq,omitempty"`
	}{
		Task: t.Task,
		List: t.List,
		User: userID,
	}

	var client mongodb.Client
	if completed {
		client = completedClient
	} else {
		client = incompleteClient

		var tasks []task
		if err := client.Find(
			mongodb.M{"user": userID, "list": t.List},
			&mongodb.FindOpt{Sort: mongodb.M{"seq": -1}, Limit: 1},
			&tasks,
		); err != nil {
			return nil, err
		}

		if len(tasks) == 0 {
			doc.Seq = 1
		} else {
			doc.Seq = tasks[0].Seq + 1
		}
	}
	doc.Created = client.Date(time.Now()).Interface()

	return client.InsertOne(doc)
}

func deleteTask(id mongodb.ObjectID, userID string, completed bool) error {
	var client mongodb.Client
	if completed {
		client = completedClient
	} else {
		client = incompleteClient
	}

	var task task
	if err := client.FindOneAndDelete(mongodb.M{"_id": id.Interface()}, nil, &task); err != nil {
		return err
	}

	if !completed {
		_, err := client.UpdateMany(
			mongodb.M{"user": userID, "list": task.List, "seq": mongodb.M{"$gt": task.Seq}},
			mongodb.M{"$inc": mongodb.M{"seq": -1}},
			nil,
		)
		return err
	}

	return nil
}

func reorderTask(userID, list string, orig, dest mongodb.ObjectID) error {
	var origTask, destTask task

	c := make(chan error, 1)
	go func() {
		c <- incompleteClient.FindOne(mongodb.M{"_id": orig.Interface()}, nil, &origTask)
	}()
	if err := incompleteClient.FindOne(mongodb.M{"_id": dest.Interface()}, nil, &destTask); err != nil {
		return err
	}
	if err := <-c; err != nil {
		return err
	}

	if origTask.List != list || destTask.List != list {
		return errors.New("list not match")
	}

	var filter, update mongodb.M
	if origTask.Seq > destTask.Seq {
		filter = mongodb.M{"user": userID, "list": list, "seq": mongodb.M{"$gte": destTask.Seq, "$lt": origTask.Seq}}
		update = mongodb.M{"$inc": mongodb.M{"seq": 1}}
	} else {
		filter = mongodb.M{"user": userID, "list": list, "seq": mongodb.M{"$gt": origTask.Seq, "$lte": destTask.Seq}}
		update = mongodb.M{"$inc": mongodb.M{"seq": -1}}
	}

	if _, err := incompleteClient.UpdateMany(filter, update, nil); err != nil {
		return err
	}

	if _, err := incompleteClient.UpdateOne(
		mongodb.M{"_id": orig.Interface()},
		mongodb.M{"$set": mongodb.M{"seq": destTask.Seq}},
		nil,
	); err != nil {
		return err
	}

	return nil
}
