package temp_workflow

import (
	"context"
	"time"

	"go.temporal.io/sdk/workflow"
)

const (
	MAX_ITERATIONS = 5000
	UPDATE_CHANNEL = "update-chan"
)

type Event struct {
	Status string
}

func (e *Event) IsDone() bool {
	return e != nil && e.Status == "Done"
}

func Workflow(ctx workflow.Context, raised_event *Event) error {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToCloseTimeout: time.Minute,
	})
	var created_id string
	if err := workflow.ExecuteActivity(ctx, ActivityOne, raised_event).Get(ctx, &created_id); err != nil {
		return err
	}
	selector := workflow.NewSelector(ctx)
	signalChan := workflow.GetSignalChannel(ctx, UPDATE_CHANNEL)
	var update_event *Event
	selector.AddReceive(signalChan, func(channel workflow.ReceiveChannel, _ bool) {
		channel.Receive(ctx, &update_event)
	})
	for it := 0; it < MAX_ITERATIONS; it++ {
		selector.Select(ctx)
		workflow.ExecuteActivity(ctx, ActivityTwo, update_event)
		if update_event.IsDone() {
			return nil
		}
	}
	return workflow.NewContinueAsNewError(ctx, Workflow, raised_event)
}

func ActivityOne(ctx context.Context, e *Event) (string, error) {
	return "created_id", nil
}

func ActivityTwo(ctx context.Context, e *Event) error {
	return nil
}
