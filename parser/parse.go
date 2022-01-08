package parser

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/robertkrimen/otto"
)

var (
	once       sync.Once
	globalOtto *otto.Otto
	tagStore   = make(TagStore)
)

func makeOtto(fcl *FCL) *otto.Otto {
	runtime := otto.New()
	runtime.Set("$$", &FCLConnector{
		fcl:     fcl,
		runtime: runtime,
	})

	// hacky way to convert golang public struct methods to
	// lower case javascript methods
	runtime.Run(`
		$$ = Object.keys($$).reduce(function(acc, key) {
			acc[key.toLowerCase()] = $$[key];
			return acc;
		}, {});
	`)

	return runtime
}

func getOtto(fcl *FCL, singleton bool) *otto.Otto {
	if singleton {
		once.Do(func() {
			globalOtto = makeOtto(fcl)
		})

		return globalOtto
	}

	return makeOtto(fcl)
}

func getAttr(name string, attrs []xml.Attr) *string {
	for _, attr := range attrs {
		if attr.Name.Local == name {
			return &attr.Value
		}
	}

	return nil
}

func validAttr(s *string) bool {
	return s != nil && *s != ""
}

func yesNoAttr(s *string) bool {
	if s == nil {
		return false
	}

	return *s == "yes"
}

func listenSignal(fcl *FCL) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGABRT, syscall.SIGTERM)
	<-c
	fcl.File.Close()
}

func processTag(dec *xml.Decoder, elem *xml.StartElement, fcl *FCL) error {
	tagName := elem.Name.Local
	switch tagName {
	case "fcl":
		version := getAttr("version", elem.Attr)
		if !validAttr(version) {
			return errors.New("version MUST be supplied when creating an FCL config")
		}

		fcl.Version = *version
	case "scripts":
		isShared := getAttr("shared", elem.Attr)
		if !validAttr(isShared) {
			return errors.New("shared attribute for scripts was used incorrectly")
		}

		var scripts ScriptsTag
		if err := dec.DecodeElement(&scripts, elem); err != nil {
			return err
		}

		scripts.Shared = yesNoAttr(isShared)
		fcl.ScriptData = &scripts
	default:
		if tagName != "" {
			var data GenericTag
			if err := dec.DecodeElement(&data, elem); err != nil {
				return err
			}

			tagStore.Set(tagName, data.Content)
		}
	}

	return nil
}

func ParseInput(fileName string) (*FCL, error) {
	cfgFile, err := os.OpenFile(fileName, os.O_RDWR, 0777)
	if err != nil {
		return nil, err
	}

	fcl := FCL{
		File:       cfgFile,
		ScriptData: new(ScriptsTag),
	}

	dec := xml.NewDecoder(cfgFile)
	go listenSignal(&fcl)

	for {
		tok, err := dec.Token()
		if tok == nil || err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("unable to parse token! Error: %s", err.Error())
		}

		switch t := tok.(type) {
		case xml.StartElement:
			if err := processTag(dec, &t, &fcl); err != nil {
				return nil, err
			}
		case xml.EndElement:
			continue
		}
	}

	if err := ParseScripts(&fcl); err != nil {
		return nil, err
	}

	return &fcl, nil
}

func ParseScripts(fcl *FCL) error {
	scripts := fcl.ScriptData

	for _, script := range scripts.Scripts {
		exe := getOtto(fcl, scripts.Shared)
		if err := UseEventLoop(exe, script.Content); err != nil {
			return err
		}
	}

	return nil
}
