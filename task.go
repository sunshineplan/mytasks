package main

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
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

func checkTask(id mongodb.ObjectID, userID any, completed bool) bool {
	var client mongodb.Client
	if completed {
		client = completedClient
	} else {
		client = incompleteClient
	}

	n, _ := client.CountDocuments(mongodb.M{"_id": id.Interface(), "user": userID}, nil)
	return n > 0
}

func fetchTask(list, userID string, completed bool) ([]task, error) {
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
	}

	return tasks, nil
}

func getTask(c *gin.Context) {
	var data struct{ List string }
	if err := c.BindJSON(&data); err != nil {
		c.Status(400)
		return
	}

	user, err := getUser(c)
	if err != nil {
		svc.Print(err)
		c.Status(500)
		return
	}

	var incomplete []task
	ec := make(chan error, 1)
	go func() {
		var err error
		incomplete, err = fetchTask(data.List, user.ID, false)
		ec <- err
	}()

	completed, err := fetchTask(data.List, user.ID, true)
	if err != nil {
		svc.Println("Failed to get completed tasks:", err)
		c.Status(500)
		return
	}

	if err := <-ec; err != nil {
		svc.Println("Failed to get incomplete tasks:", err)
		c.Status(500)
		return
	}

	c.JSON(200, gin.H{"incomplete": incomplete, "completed": completed})
}

func addTask(t task, userID string, completed bool) (any, int, error) {
	doc := struct {
		Task    string `json:"task" bson:"task"`
		List    string `json:"list" bson:"list"`
		User    string `json:"user" bson:"user"`
		Created any    `json:"created" bson:"created"`
		Seq     int    `json:"seq,omitempty" bson:"seq,omitempty"`
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
			return nil, 0, err
		}

		if len(tasks) == 0 {
			doc.Seq = 1
		} else {
			doc.Seq = tasks[0].Seq + 1
		}
	}
	doc.Created = client.Date(time.Now()).Interface()
	id, err := client.InsertOne(doc)
	if err != nil {
		return nil, 0, err
	}
	return id, doc.Seq, nil
}

func deleteTask(id mongodb.ObjectID, userID string, completed bool) (err error) {
	var client mongodb.Client
	if completed {
		client = completedClient
	} else {
		client = incompleteClient
	}
	var task task
	if err = client.FindOneAndDelete(mongodb.M{"_id": id.Interface()}, nil, &task); err != nil {
		return
	}
	if !completed {
		_, err = client.UpdateMany(
			mongodb.M{"user": userID, "list": task.List, "seq": mongodb.M{"$gt": task.Seq}},
			mongodb.M{"$inc": mongodb.M{"seq": -1}},
			nil,
		)
	}
	return
}

func reorderTask(userID, list string, orig, dest mongodb.ObjectID) (err error) {
	var origTask, destTask task
	c := make(chan error, 1)
	go func() {
		c <- incompleteClient.FindOne(mongodb.M{"_id": orig.Interface()}, nil, &origTask)
	}()
	if err = incompleteClient.FindOne(mongodb.M{"_id": dest.Interface()}, nil, &destTask); err != nil {
		return
	}
	if err = <-c; err != nil {
		return
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
	if _, err = incompleteClient.UpdateMany(filter, update, nil); err != nil {
		return
	}
	_, err = incompleteClient.UpdateOne(
		mongodb.M{"_id": orig.Interface()},
		mongodb.M{"$set": mongodb.M{"seq": destTask.Seq}},
		nil,
	)
	return
}
