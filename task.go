package main

import (
	"errors"
	"time"

	"github.com/sunshineplan/database/mongodb/api"
)

type task struct {
	ID       string    `json:"id"`
	ObjectID string    `json:"_id,omitempty"`
	Task     string    `json:"task"`
	List     string    `json:"list"`
	Created  time.Time `json:"created"`
	Seq      int       `json:"seq,omitempty"`
}

func checkTask(id string, userID interface{}, completed bool) bool {
	var client *api.Client
	if completed {
		client = &completedClient
	} else {
		client = &incompleteClient
	}

	n, _ := client.CountDocuments(api.M{"_id": api.ObjectID(id), "user": userID}, nil)
	return n > 0
}

func getTask(list, userID string, completed bool) ([]task, error) {
	var client *api.Client
	var opt *api.FindOpt
	if completed {
		client = &completedClient
		opt = &api.FindOpt{Sort: api.M{"created": -1}, Limit: 10}
	} else {
		client = &incompleteClient
		opt = &api.FindOpt{Sort: api.M{"seq": -1}}
	}

	tasks := []task{}
	if err := client.Find(api.M{"list": list, "user": userID}, opt, &tasks); err != nil {
		return nil, err
	}
	for i := range tasks {
		tasks[i].ID = tasks[i].ObjectID
		tasks[i].ObjectID = ""
		tasks[i].Seq = 0
	}

	return tasks, nil
}

func addTask(t task, userID string, completed bool) (string, error) {
	doc := struct {
		Task    string      `json:"task"`
		List    string      `json:"list"`
		User    string      `json:"user"`
		Created interface{} `json:"created"`
		Seq     int         `json:"seq,omitempty"`
	}{
		Task:    t.Task,
		List:    t.List,
		User:    userID,
		Created: api.Date(time.Now()),
	}

	var client *api.Client
	if completed {
		client = &completedClient
	} else {
		client = &incompleteClient

		var tasks []task
		if err := client.Find(
			api.M{"user": userID, "list": t.List},
			&api.FindOpt{Sort: api.M{"seq": -1}, Limit: 1},
			&tasks,
		); err != nil {
			return "", err
		}

		if len(tasks) == 0 {
			doc.Seq = 1
		} else {
			doc.Seq = tasks[0].Seq + 1
		}
	}

	insertID, err := client.InsertOne(doc)
	if err != nil {
		return "", err
	}

	return insertID, nil
}

func deleteTask(id string, userID string, completed bool) error {
	var client *api.Client
	if completed {
		client = &completedClient
	} else {
		client = &incompleteClient
	}

	var task task
	if err := client.FindOneAndDelete(api.M{"_id": api.ObjectID(id)}, nil, &task); err != nil {
		return err
	}

	if !completed {
		_, err := client.UpdateMany(
			api.M{"user": userID, "list": task.List, "seq": api.M{"$gt": task.Seq}},
			api.M{"$inc": api.M{"seq": -1}},
			nil,
		)
		return err
	}

	return nil
}

func reorderTask(userID, list string, orig, dest string) error {
	var origTask, destTask task

	c := make(chan error, 1)
	go func() {
		c <- incompleteClient.FindOne(api.M{"_id": api.ObjectID(orig)}, nil, &origTask)
	}()
	if err := incompleteClient.FindOne(api.M{"_id": api.ObjectID(dest)}, nil, &destTask); err != nil {
		return err
	}
	if err := <-c; err != nil {
		return err
	}

	if origTask.List != list || destTask.List != list {
		return errors.New("list not match")
	}

	var filter, update api.M
	if origTask.Seq > destTask.Seq {
		filter = api.M{"user": userID, "list": list, "seq": api.M{"$gte": destTask.Seq, "$lt": origTask.Seq}}
		update = api.M{"$inc": api.M{"seq": 1}}
	} else {
		filter = api.M{"user": userID, "list": list, "seq": api.M{"$gt": origTask.Seq, "$lte": destTask.Seq}}
		update = api.M{"$inc": api.M{"seq": -1}}
	}

	if _, err := incompleteClient.UpdateMany(filter, update, nil); err != nil {
		return err
	}

	if _, err := incompleteClient.UpdateOne(
		api.M{"_id": api.ObjectID(orig)},
		api.M{"$set": api.M{"seq": destTask.Seq}},
		nil,
	); err != nil {
		return err
	}

	return nil
}
