package taskset

import "context"

// Middleware is used to inject code into task execution, as well as into
// dependency declaration.
//
// For some examples of middlewares, see github.com/bennydictor/taskset/middlewares.
type Middleware struct {
	// Run injects code into task execution. Middlewares must call next() exactly once
	// during Run.  Middlewares may examine and modify task's properties at any point
	// during Run.  Middlewares may pass a modified context to next(), although it must
	// be derived from the input context.  Middlewares may examine and modify the
	// task's result before returning it.  Leave Run equal to nil to not do anything on
	// task execution.
	Run func(ctx context.Context, task *Task, next func(ctx context.Context) Result) Result

	// Depend injects code into task dependency declaration.  Middlewares must call
	// next() exactly once during Depend.  Middlewares may examine and modify both
	// tasks' properties at any point during Depend.  Middlewares may pass a modified
	// context to next(), although it must be derived from input context.  Middlewares
	// may examine and modify the result of the dependency task before returning it.
	// Note that if the result is modified, the changes are only propagated to the
	// caller of depend(), it doesn't actually modify the dependent task's result.
	// Leave Depend equal to nil to not do anything on dependency declaration.
	Depend func(ctx context.Context, task, dependency *Task, next func(ctx context.Context) Result) Result
}

func composeRun(mw1, mw2 Middleware) func(ctx context.Context, task *Task, next func(ctx context.Context) Result) Result {
	if mw1.Run == nil {
		return mw2.Run
	}
	if mw2.Run == nil {
		return mw1.Run
	}

	return func(ctx context.Context, task *Task, next func(ctx context.Context) Result) Result {
		return mw1.Run(ctx, task, func(ctx context.Context) Result {
			return mw2.Run(ctx, task, next)
		})
	}
}

func composeDepend(mw1, mw2 Middleware) func(ctx context.Context, task, dependency *Task, next func(ctx context.Context) Result) Result {
	if mw1.Depend == nil {
		return mw2.Depend
	}
	if mw2.Depend == nil {
		return mw1.Depend
	}

	return func(ctx context.Context, task, dependency *Task, next func(ctx context.Context) Result) Result {
		return mw1.Depend(ctx, task, dependency, func(ctx context.Context) Result {
			return mw2.Depend(ctx, task, dependency, next)
		})
	}
}

func composeMiddlewares(mw1, mw2 Middleware) Middleware {
	return Middleware{
		Run:    composeRun(mw1, mw2),
		Depend: composeDepend(mw1, mw2),
	}
}

func chainMiddlewares(middlewares []Middleware) (result Middleware) {
	for _, m := range middlewares {
		result = composeMiddlewares(result, m)
	}
	return
}

// middleware is a monoid:
// mempty = Middleware{id, id}
// mappend = composeMiddlewares
// mconcat = chainMiddlewares
