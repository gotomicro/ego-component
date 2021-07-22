package ejenkins

import (
	"strconv"
)

type Queue struct {
	Jenkins *Jenkins
	Raw     *queueResponse
	Base    string
}

type queueResponse struct {
	Items []taskResponse
}

type Task struct {
	Raw     *taskResponse
	Jenkins *Jenkins
	Queue   *Queue
	Base    string
}

type taskResponse struct {
	Actions                    []generalAction `json:"actions"`
	Blocked                    bool            `json:"blocked"`
	Buildable                  bool            `json:"buildable"`
	BuildableStartMilliseconds int64           `json:"buildableStartMilliseconds"`
	ID                         int64           `json:"id"`
	InQueueSince               int64           `json:"inQueueSince"`
	Params                     string          `json:"params"`
	Pending                    bool            `json:"pending"`
	Stuck                      bool            `json:"stuck"`
	Task                       struct {
		Color string `json:"color"`
		Name  string `json:"name"`
		URL   string `json:"url"`
	} `json:"task"`
	URL        string `json:"url"`
	Why        string `json:"why"`
	Executable struct {
		Number int64  `json:"number"`
		URL    string `json:"url"`
	} `json:"executable"`
}

type generalAction struct {
	Causes     []map[string]interface{}
	Parameters []parameter
}

func (q *Queue) Tasks() []*Task {
	tasks := make([]*Task, len(q.Raw.Items))
	for i, t := range q.Raw.Items {
		tasks[i] = &Task{Jenkins: q.Jenkins, Queue: q, Raw: &t}
	}
	return tasks
}

func (q *Queue) GetTaskById(id int64) *Task {
	for _, t := range q.Raw.Items {
		if t.ID == id {
			return &Task{Jenkins: q.Jenkins, Queue: q, Raw: &t}
		}
	}
	return nil
}

func (q *Queue) GetTasksForJob(name string) []*Task {
	tasks := make([]*Task, 0)
	for _, t := range q.Raw.Items {
		if t.Task.Name == name {
			tasks = append(tasks, &Task{Jenkins: q.Jenkins, Queue: q, Raw: &t})
		}
	}
	return tasks
}

func (q *Queue) CancelTask(id int64) (bool, error) {
	task := q.GetTaskById(id)
	return task.Cancel()
}

func (t *Task) Cancel() (bool, error) {
	qr := map[string]string{
		"id": strconv.FormatInt(t.Raw.ID, 10),
	}
	response, err := t.Jenkins.Requester.Post(t.Jenkins.GetQueueUrl()+"/cancelItem", nil, t.Raw, qr)
	if err != nil {
		return false, err
	}
	return response.StatusCode() == 200, nil
}

func (t *Task) GetJob() (*Job, error) {
	return t.Jenkins.GetJob(t.Raw.Task.Name)
}

func (t *Task) GetWhy() string {
	return t.Raw.Why
}

func (t *Task) GetParameters() []parameter {
	for _, a := range t.Raw.Actions {
		if a.Parameters != nil {
			return a.Parameters
		}
	}
	return nil
}

func (t *Task) GetCauses() []map[string]interface{} {
	for _, a := range t.Raw.Actions {
		if a.Causes != nil {
			return a.Causes
		}
	}
	return nil
}

func (q *Queue) Poll() (int, error) {
	response, err := q.Jenkins.Requester.GetJSON(q.Base, nil, q.Raw, nil)
	if err != nil {
		return 0, err
	}
	return response.StatusCode(), nil
}

func (t *Task) Poll() (int, error) {
	response, err := t.Jenkins.Requester.GetJSON(t.Base, nil, t.Raw, nil)
	if err != nil {
		return 0, err
	}
	return response.StatusCode(), nil
}
