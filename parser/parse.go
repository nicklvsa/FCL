package parser

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/robertkrimen/otto"
)

var (
	once sync.Once
	globalOtto *otto.Otto
	tagStore = make(TagStore)
)

func makeOtto() *otto.Otto {
	runtime := otto.New()
	runtime.Set("$$", &FCLConnector{
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

func getOtto(singleton bool) *otto.Otto {
	if singleton {
		once.Do(func() {
			globalOtto = makeOtto()
		})

		return globalOtto
	}

	return makeOtto()
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
	cfgFile, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	defer cfgFile.Close()

	fcl := FCL{
		ScriptData: new(ScriptsTag),
	}

	dec := xml.NewDecoder(cfgFile)

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
	
	if err := ParseScripts(fcl.ScriptData); err != nil {
		return nil, err
	}

	return &fcl, nil
}

func ParseScripts(scripts *ScriptsTag) error {	
	for _, script := range scripts.Scripts {
		exe := getOtto(scripts.Shared)
		_, err := exe.Run(script.Content)
		if err != nil {
			return err
		}
	}

	return nil
}