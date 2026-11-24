package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	out := relay(in, done, false)
	for _, stage := range stages {
		out = relay(stage(out), done, true)
	}
	return out
}

func relay(in, done In, drainOnCancel bool) Out {
	if done == nil {
		return in
	}

	out := make(Bi)
	go func() {
		defer func() {
			close(out)
			if drainOnCancel && in != nil {
				for open := true; open; {
					_, open = <-in
				}
			}
		}()
		for {
			select {
			case <-done:
				return
			default:
			}
			select {
			case <-done:
				return
			case value, ok := <-in:
				if !ok {
					return
				}
				select {
				case <-done:
					return
				case out <- value:
				}
			}
		}
	}()
	return out
}
