package raymond

import "testing"

// A block helper renders its body while its own invocation is still on the
// stack. Helpers evaluated inside that body must see their own invocation
// mode, not the enclosing block's.
func TestBlockFlagsDoNotLeakIntoBody(t *testing.T) {
	RegisterHelper("modeProbe", func(options *Options) interface{} {
		switch {
		case options.IsSubExpression():
			return "sub"
		case options.IsBlock():
			return "block:" + options.Fn()
		default:
			return "inline"
		}
	})
	defer RemoveHelper("modeProbe")
	RegisterHelper("passBody", func(options *Options) interface{} {
		return options.Fn() + "|" + options.Inverse()
	})
	defer RemoveHelper("passBody")
	RegisterHelper("echo", func(v interface{}) interface{} { return v })
	defer RemoveHelper("echo")

	tests := []struct {
		name, tpl, want string
	}{
		{"inline at top level", `{{modeProbe}}`, "inline"},
		{"inline inside block body", `{{#passBody}}{{modeProbe}}{{/passBody}}`, "inline|"},
		{"inline inside else body", `{{#passBody}}a{{else}}{{modeProbe}}{{/passBody}}`, "a|inline"},
		{"inline inside builtin if", `{{#if true}}{{modeProbe}}{{/if}}`, "inline"},
		{"inline inside each", `{{#each list}}{{modeProbe}}{{/each}}`, "inlineinline"},
		{"nested block inside block body", `{{#passBody}}{{#modeProbe}}x{{/modeProbe}}{{/passBody}}`, "block:x|"},
		{"subexpression inside block body", `{{#passBody}}{{echo (modeProbe)}}{{/passBody}}`, "sub|"},
		{"inline after block", `{{#passBody}}x{{/passBody}}{{modeProbe}}`, "x|inline"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Render(tc.tpl, map[string]interface{}{"list": []int{1, 2}})
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
