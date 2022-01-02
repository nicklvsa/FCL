package parser

import (
	"time"

	"github.com/robertkrimen/otto"
)

// modified from - https://github.com/robertkrimen/natto

func UseEventLoop(runtime *otto.Otto, script string) error {
	ready := make(chan *FCLTimer)
	registry := make(map[*FCLTimer]*FCLTimer)

	makeTimer := func(call otto.FunctionCall, interval bool) (*FCLTimer, otto.Value) {
		delay, _ := call.Argument(1).ToInteger()
		if 0 >= delay {
			delay = 1
		}

		timer := FCLTimer{
			duration: time.Duration(delay) * time.Millisecond,
			interval: interval,
			call: call,
		}

		registry[&timer] = &timer
		timer.timer = time.AfterFunc(timer.duration, func() {
			ready <- &timer
		})

		value, err := call.Otto.ToValue(&timer)
		if err != nil {
			panic(err)
		}

		return &timer, value
	}

	clearFunc := func(call otto.FunctionCall) otto.Value {
		timer, _ := call.Argument(0).Export()
		if timer, ok := timer.(*FCLTimer); ok {
			timer.timer.Stop()
			delete(registry, timer)
		}

		return otto.Value{}
	}


	runtime.Set("setTimeout", func(call otto.FunctionCall) otto.Value {
		_, val := makeTimer(call, false)
		return val
	})

	runtime.Set("setInterval", func(call otto.FunctionCall) otto.Value {
		_, val := makeTimer(call, true)
		return val
	})

	runtime.Set("clearTimeout", clearFunc)
	runtime.Set("clearInterval", clearFunc)

	if _, err := runtime.Run(script); err != nil {
		return err
	}

	for {
		select {
		case timer := <-ready:
			var args []interface{}
			if len(timer.call.ArgumentList) > 2 {
				tmp := timer.call.ArgumentList[2:]
				args = make([]interface{}, len(tmp) + 2)
				for idx, val := range tmp {
					args[idx + 2] = val
				}
			} else {
				args = make([]interface{}, 1)
			}

			args[0] = timer.call.ArgumentList[0]
			if _, err := runtime.Call("Function.call.call", nil, args...); err != nil {
				for _, timer := range registry {
					timer.timer.Stop()
					delete(registry, timer)
					return err
				}
			}

			if !timer.interval {
				delete(registry, timer)
			} else {
				timer.timer.Reset(timer.duration)
			}

		default:
		}

		if len(registry) == 0 {
			break
		}
	}

	return nil
}