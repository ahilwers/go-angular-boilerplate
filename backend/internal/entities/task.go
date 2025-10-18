package entities

import (
	"fmt"
	"time"
)

type TaskStatus int

const (
	TaskStatusTodo TaskStatus = iota
	TaskStatusInProgress
	TaskStatusDone
)

func (s TaskStatus) String() string {
	switch s {
	case TaskStatusTodo:
		return "TODO"
	case TaskStatusInProgress:
		return "IN_PROGRESS"
	case TaskStatusDone:
		return "DONE"
	default:
		return "UNKNOWN"
	}
}

func (s TaskStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, s.String())), nil
}

func (s *TaskStatus) UnmarshalJSON(data []byte) error {
	str := string(data)
	// Remove quotes
	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}

	switch str {
	case "TODO":
		*s = TaskStatusTodo
	case "IN_PROGRESS":
		*s = TaskStatusInProgress
	case "DONE":
		*s = TaskStatusDone
	default:
		return fmt.Errorf("invalid task status: %s", str)
	}
	return nil
}

func ParseTaskStatus(s string) (TaskStatus, error) {
	switch s {
	case "TODO":
		return TaskStatusTodo, nil
	case "IN_PROGRESS":
		return TaskStatusInProgress, nil
	case "DONE":
		return TaskStatusDone, nil
	default:
		return TaskStatusTodo, fmt.Errorf("invalid task status: %s", s)
	}
}

type Task struct {
	ID          string     `json:"id"`
	ProjectID   string     `json:"projectId"`
	Title       string     `json:"title"`
	Status      TaskStatus `json:"status"`
	DueDate     *time.Time `json:"dueDate,omitempty"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}
