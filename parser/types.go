package parser

import "github.com/robertkrimen/otto"

type TagStore map[string]interface{}

type Script struct {
	Content string `xml:",chardata"`
	Ref     string `xml:"ref,attr"`
}

type GenericTag struct {
	Content string `xml:",chardata"`
}

type ScriptsTag struct {
	Shared  bool
	Scripts []Script `xml:"script"`
}

type FCL struct {
	Version    string
	ScriptData *ScriptsTag
}

type FCLConnector struct {
	runtime *otto.Otto
}
