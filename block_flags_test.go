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
	// echoBlock returns its argument (or hash value v) instead of rendering
	// its body, exposing what the argument evaluated to.
	RegisterHelper("echoBlock", func(options *Options) interface{} {
		if len(options.Params()) > 0 {
			return options.Param(0)
		}
		return options.HashProp("v")
	})
	defer RemoveHelper("echoBlock")
	RegisterHelper("modeProbe2", func(_ interface{}, options *Options) interface{} {
		switch {
		case options.IsSubExpression():
			return "sub"
		case options.IsBlock():
			return "block:" + options.Fn()
		default:
			return "inline"
		}
	})
	defer RemoveHelper("modeProbe2")
	// isBlockProbe reports IsBlock() alone, as helpers that only check
	// IsBlock (and not IsSubExpression) see it.
	RegisterHelper("isBlockProbe", func(options *Options) interface{} {
		if options.IsBlock() {
			return "block"
		}
		return "notblock"
	})
	defer RemoveHelper("isBlockProbe")
	RegisterHelper("eq2", func(a, b interface{}) interface{} { return a == b })
	defer RemoveHelper("eq2")

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
		// A block helper's own arguments are separate invocations, evaluated
		// before the block helper runs.
		{"subexpression as builtin if argument", `{{#if (modeProbe)}}body{{/if}}`, "body"},
		{"subexpression as block helper argument", `{{#echoBlock (modeProbe)}}body{{/echoBlock}}`, "sub"},
		{"subexpression as block helper hash value", `{{#echoBlock v=(modeProbe)}}body{{/echoBlock}}`, "sub"},
		{"nested subexpression as block argument", `{{#echoBlock (echo (modeProbe))}}body{{/echoBlock}}`, "sub"},
		{"subexpression argument inside block body", `{{#passBody}}{{#echoBlock (modeProbe)}}b{{/echoBlock}}{{/passBody}}`, "sub|"},
		{"IsBlock false for builtin if argument", `{{#if (eq2 (isBlockProbe) 'notblock')}}ok{{else}}leak{{/if}}`, "ok"},
		{"IsBlock false for block helper argument", `{{#echoBlock (isBlockProbe)}}body{{/echoBlock}}`, "notblock"},
		{"IsBlock false for block helper hash value", `{{#echoBlock v=(isBlockProbe)}}body{{/echoBlock}}`, "notblock"},
		{"IsBlock false for nested block argument", `{{#echoBlock (echo (isBlockProbe))}}body{{/echoBlock}}`, "notblock"},
		{"IsBlock true for block helper itself", `{{#isBlockProbe}}b{{/isBlockProbe}}`, "block"},
		{"block helper still sees block after args", `{{#modeProbe2 (echo 'x')}}b{{/modeProbe2}}`, "block:b"},
		{"inline helper with subexpression arg", `{{modeProbe2 (echo 'x')}}`, "inline"},
		{"subexpression with subexpression arg", `{{echo (modeProbe2 (echo 'x'))}}`, "sub"},
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
