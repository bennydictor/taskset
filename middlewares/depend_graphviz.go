package middlewares

import (
	"bytes"
	"context"
	"io"
	"sync"

	"github.com/bennydictor/taskset"
	"github.com/bennydictor/taskset/properties"
)

// DependGraphviz provides a middleware that records
// all dependency declarations by all tasks, and makes this
// information available as a graphviz source file.
type DependGraphviz struct {
	sync.Mutex
	info map[*taskset.Task]map[*taskset.Task]struct{}
}

// NewDependGraphviz creates a new DependGraphviz.
func NewDependGraphviz() *DependGraphviz {
	return &DependGraphviz{
		info: make(map[*taskset.Task]map[*taskset.Task]struct{}),
	}
}

// Middleware provides the taskset.Middleware.
func (d *DependGraphviz) Middleware() taskset.Middleware {
	return taskset.Middleware{
		Depend: d.depend,
	}
}

func (d *DependGraphviz) depend(ctx context.Context, task, dependency *taskset.Task, next func(ctx context.Context) taskset.Result) taskset.Result {
	result := next(ctx)

	d.Lock()
	defer d.Unlock()

	if _, ok := d.info[task]; !ok {
		d.info[task] = make(map[*taskset.Task]struct{})
	}
	d.info[task][dependency] = struct{}{}

	return result
}

// Write writes the generated graphviz source file to an io.Writer.
func (d *DependGraphviz) Write(w io.Writer) error {
	d.Lock()
	defer d.Unlock()

	_, err := io.WriteString(w, "digraph {\n")
	if err != nil {
		return err
	}

	for task, dependencies := range d.info {
		_, err = io.WriteString(w, "    \"")
		if err != nil {
			return err
		}

		_, err = io.WriteString(w, properties.Name(task))
		if err != nil {
			return err
		}

		_, err = io.WriteString(w, "\" -> {")
		if err != nil {
			return err
		}

		for dependency := range dependencies {
			_, err = io.WriteString(w, " \"")
			if err != nil {
				return err
			}

			_, err = io.WriteString(w, properties.Name(dependency))
			if err != nil {
				return err
			}

			_, err = io.WriteString(w, "\"")
			if err != nil {
				return err
			}
		}

		_, err = io.WriteString(w, " };\n")
		if err != nil {
			return err
		}
	}

	_, err = io.WriteString(w, "}\n")
	if err != nil {
		return err
	}

	return nil
}

// String returns the generated graphviz source file as a string.
func (d *DependGraphviz) String() string {
	var buf bytes.Buffer
	_ = d.Write(&buf)
	return buf.String()
}
