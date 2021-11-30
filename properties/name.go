package properties

import "github.com/bennydictor/taskset"

type nameProperty struct{}

// WithName adds a human-readable name to a task. Used for logging, tracing, etc.
func WithName(name string) taskset.Property {
	return func(task *taskset.Task) {
		task.ModifyProperty(nameProperty{}, func(_ interface{}) interface{} {
			return name
		})
	}
}

// Name gets the task's name added by WithName. If no name is found,
// returns an empty string. This function should only be used by middlewares.
func Name(task *taskset.Task) string {
	name := task.Property(nameProperty{})
	if name == nil {
		return ""
	}

	return name.(string)
}
