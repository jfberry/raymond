package raymond

import (
	"fmt"
	"testing"
)

// argWrapper stands in for a context wrapper that adds fields alongside a
// value (like PoracleNG's each wrapper) and unwraps when passed to helpers.
type argWrapper struct{ v interface{} }

func (w *argWrapper) HelperArg() interface{} { return w.v }
func (w *argWrapper) String() string         { return fmt.Sprint(w.v) }
func (w *argWrapper) GetField(name string) (interface{}, bool) {
	if name == "extra" {
		return "E", true
	}
	return nil, false
}

func TestHelperArgUnwrap(t *testing.T) {
	RegisterHelper("typeOf", func(v interface{}) interface{} { return fmt.Sprintf("%T", v) })
	defer RemoveHelper("typeOf")
	RegisterHelper("hashTypeOf", func(options *Options) interface{} { return fmt.Sprintf("%T", options.HashProp("v")) })
	defer RemoveHelper("hashTypeOf")
	RegisterHelper("wrapAll", func(list []int, options *Options) interface{} {
		out := ""
		for _, n := range list {
			out += options.FnWith(&argWrapper{v: n})
		}
		return out
	})
	defer RemoveHelper("wrapAll")

	ctx := map[string]interface{}{"list": []int{0, 2}}
	tests := []struct{ name, tpl, want string }{
		{"param is unwrapped", `{{#wrapAll list}}{{typeOf this}};{{/wrapAll}}`, "int;int;"},
		{"subexpression param is unwrapped", `{{#wrapAll list}}{{typeOf (typeOf this)}};{{/wrapAll}}`, "string;string;"},
		{"hash value is unwrapped", `{{#wrapAll list}}{{hashTypeOf v=this}};{{/wrapAll}}`, "int;int;"},
		{"builtin if sees wrapped zero as falsy", `{{#wrapAll list}}{{#if this}}T{{else}}F{{/if}}{{/wrapAll}}`, "FT"},
		{"builtin unless sees wrapped zero as falsy", `{{#wrapAll list}}{{#unless this}}F{{/unless}}{{/wrapAll}}`, "F"},
		{"rendering still uses the wrapper", `{{#wrapAll list}}{{this}}{{extra}};{{/wrapAll}}`, "0E;2E;"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Render(tc.tpl, ctx)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
