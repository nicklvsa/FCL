package parser

import (
	"regexp"
	"strings"
)

var (
	TagValueTemplateRegex = "%{([^}]*)}"
)

func (t *TagStore) GetStr(tag string) string {
	if value, ok := (*t)[tag].(string); ok {
		return value
	}

	return ""
}

func (t *TagStore) Set(tag, value string) {
	r, _ := regexp.Compile(TagValueTemplateRegex)
	if matching := r.MatchString(value); matching {
		matches := r.FindAllString(value, -1)
		for _, match := range matches {
			key := strings.TrimSpace(match[2:len(match) - 1])
			value = strings.ReplaceAll(value, match, t.GetStr(key))
		}
	}

	(*t)[tag] = value
}