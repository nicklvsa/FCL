package parser

import (
	"os"

	"github.com/robertkrimen/otto"
)

func wrap(f *FCLConnector, item string, v otto.Value) otto.Value {
	wrapper := map[string]func(call otto.FunctionCall) otto.Value {
		"set": func(call otto.FunctionCall) otto.Value {
			updatedValue := call.Argument(0).String()

			value, err := f.runtime.ToValue(updatedValue)
			if err != nil {
				panic(err)
			}

			tagStore.Set(item, updatedValue)
			return value
		},
		"assign": func(call otto.FunctionCall) otto.Value {
			updatedValue := call.Argument(0).String()
			tagStore.Set(item, updatedValue)
			
			return v
		},
		"val": func(call otto.FunctionCall) otto.Value {
			return v
		},
	}

	value, err := f.runtime.ToValue(wrapper)
	if err != nil {
		panic(err)
	}

	return value
}

func (f *FCLConnector) Get(call otto.FunctionCall) otto.Value {
	item := call.Argument(0).String()
	value, err := f.runtime.ToValue(tagStore.GetStr(item))
	if err != nil {
		panic(err)
	}

	return wrap(f, item, value)
}

func (f *FCLConnector) Env(call otto.FunctionCall) otto.Value {
	item := call.Argument(0).String()
	value, err := f.runtime.ToValue(os.Getenv(item))
	if err != nil {
		panic(err)
	}

	return wrap(f, item, value)
}