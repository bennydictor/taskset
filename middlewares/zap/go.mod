module github.com/bennydictor/taskset/middlewares/zap

go 1.17

replace github.com/bennydictor/taskset => ../..

require (
	github.com/bennydictor/taskset v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.19.1
)

require (
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
)
