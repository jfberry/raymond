package raymond

import (
	"fmt"
	"testing"
)

func TestInlineHelperInsideEachBlock(t *testing.T) {
	// Reproduces the infinite recursion: getPowerUpCost used as a subexpression
	// inside {{#each}}, matching the user's DTS template pattern:
	// {{#each pvp_rankings_great_league}}
	//   {{{replace (getPowerUpCost ../level this.level) 'x' 'y'}}}
	// {{/each}}

	// Register a block helper that calls FnWith (like getPowerUpCost)
	RegisterHelper("powerUpCost", func(startLevel, endLevel interface{}, options *Options) interface{} {
		result := map[string]interface{}{
			"stardust": 1000,
			"candy":    5,
		}
		if options.IsBlock() {
			return options.FnWith(result)
		}
		return fmt.Sprintf("%v stardust, %v candy", result["stardust"], result["candy"])
	})

	// Register a simple replace helper (like the real one)
	RegisterHelper("testReplace", func(s, old, new interface{}) interface{} {
		return fmt.Sprintf("%v", s)
	})

	tpl := `{{#each items}}{{testReplace (powerUpCost ../baseLevel this.targetLevel) "x" "y"}} | {{/each}}`

	ctx := map[string]interface{}{
		"baseLevel": 25,
		"items": []map[string]interface{}{
			{"targetLevel": 40, "name": "First"},
			{"targetLevel": 50, "name": "Second"},
		},
	}

	result, err := Render(tpl, ctx)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	t.Logf("Result: %s", result)

	// Should contain the inline format, not recurse infinitely
	if result == "" {
		t.Error("empty result")
	}
}
