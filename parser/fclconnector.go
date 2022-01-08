package parser

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/robertkrimen/otto"
)

func stringifyVals(args []otto.Value) []string {
	var arr []string
	for _, arg := range args {
		arr = append(arr, arg.String())
	}

	return arr
}

func wrap(f *FCLConnector, item string, v otto.Value) otto.Value {
	wrapper := map[string]func(call otto.FunctionCall) otto.Value{
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

func (f *FCLConnector) Call(call otto.FunctionCall) otto.Value {
	command := call.Argument(0).String()
	args := stringifyVals(call.ArgumentList[1:])

	cmd := exec.Command(command, args...)

	out, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	value, err := f.runtime.ToValue(string(out))
	if err != nil {
		panic(err)
	}

	return value
}

func (f *FCLConnector) OS(call otto.FunctionCall) otto.Value {
	currentOS := runtime.GOOS

	switch currentOS {
	case "darwin":
		currentOS = "macos"
	}

	value, err := f.runtime.ToValue(currentOS)
	if err != nil {
		panic(err)
	}

	return value
}

func (f *FCLConnector) Key(call otto.FunctionCall) otto.Value {
	key := call.Argument(0).String()
	value := call.Argument(1).String()

	if key != "" && value != "" {
		tagStore.Set(key, value)

		data := strings.TrimSpace(
			fmt.Sprintf(
				"<%s>%s</%s>",
				key,
				tagStore.GetStr(key),
				key,
			),
		)

		if _, err := f.fcl.File.Write([]byte("\n" + data + "\n")); err != nil {
			panic(err)
		}
	}

	return otto.Value{}
}
