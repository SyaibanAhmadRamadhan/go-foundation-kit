package databases

import (
	"regexp"
	"strings"
)

var (
	spaceRe   = regexp.MustCompile(`\s+`)
	escWSRe   = regexp.MustCompile(`\\[ntr]`) // literal "\n", "\t", "\r"
	commentRe = regexp.MustCompile(`--.*?$|/\*.*?\*/`)
)

func NormalizeSQL(sql string) string {
	s := escWSRe.ReplaceAllString(sql, " ")
	s = commentRe.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)
	s = spaceRe.ReplaceAllString(s, " ")
	return s
}
