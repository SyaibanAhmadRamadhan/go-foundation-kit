package databases

import (
	"bytes"
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

func IsNullLiteral(b []byte) bool {
	return bytes.Equal(b, []byte("null")) || bytes.Equal(b, []byte("NULL"))
}
