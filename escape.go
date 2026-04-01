package raymond

import (
	"bytes"
	"strings"
)

//
// That whole file is borrowed from https://github.com/golang/go/tree/master/src/html/escape.go
//
// With changes:
//    &#39 => &apos;
//    &#34 => &quot;
//
// To stay in sync with JS implementation, and make mustache tests pass.
//

type writer interface {
	WriteString(string) (int, error)
}

const escapedChars = `&'<>"`

func escape(w writer, s string) error {
	i := strings.IndexAny(s, escapedChars)
	for i != -1 {
		if _, err := w.WriteString(s[:i]); err != nil {
			return err
		}
		var esc string
		switch s[i] {
		case '&':
			esc = "&amp;"
		case '\'':
			esc = "&apos;"
		case '<':
			esc = "&lt;"
		case '>':
			esc = "&gt;"
		case '"':
			esc = "&quot;"
		default:
			panic("unrecognized escape character")
		}
		s = s[i+1:]
		if _, err := w.WriteString(esc); err != nil {
			return err
		}
		i = strings.IndexAny(s, escapedChars)
	}
	_, err := w.WriteString(s)
	return err
}

// EscapeFunc is the function used to escape expression output in {{double}}
// brace expressions. Defaults to HTML escaping. Set to a no-op function to
// disable escaping (useful when output is JSON/Discord/Telegram, not HTML):
//
//	raymond.EscapeFunc = func(s string) string { return s }
var EscapeFunc func(string) string = DefaultEscape

// Escape escapes special HTML characters using EscapeFunc.
//
// It can be used by helpers that return a SafeString and that need to escape some content by themselves.
func Escape(s string) string {
	return EscapeFunc(s)
}

// DefaultEscape is the built-in HTML escaper. It is exported so callers can
// use it in helpers (e.g. {{escapeUrl}}) even when EscapeFunc has been
// overridden to a no-op.
func DefaultEscape(s string) string {
	if strings.IndexAny(s, escapedChars) == -1 {
		return s
	}
	var buf bytes.Buffer
	escape(&buf, s)
	return buf.String()
}
