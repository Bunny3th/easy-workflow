package event

import(
	."easy-workflow/pkg/workflow/model"
	"fmt"
)

type EventRun func(event *Event,ProcessInstanceID int, CurrentNode Node, PrevNode Node) error

type Event struct {}

func(e *Event) MyEvent(ProcessInstanceID int, CurrentNode Node, PrevNode Node) error{
	fmt.Println("fucking shit!!!!!")
	return nil
}
