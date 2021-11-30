package taskset

// Property represents a key-value property assigned to a task on creation.
// Properties are used by Middlewares.
//
// For some examples of properties, see github.com/bennydictor/taskset/properties.
type Property func(*Task)
