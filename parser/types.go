package parser

import (
	"os"
	"time"

	"github.com/robertkrimen/otto"
)

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
	File *os.File
	ScriptData *ScriptsTag
}

type FCLConnector struct {
	fcl *FCL
	runtime *otto.Otto
}

type FCLTimer struct {
	timer *time.Timer
	duration time.Duration
	interval bool
	call otto.FunctionCall
}
