package raymond

import "testing"

type testResolver struct {
	layers []map[string]interface{}
}

func (r *testResolver) GetField(name string) (interface{}, bool) {
	for _, layer := range r.layers {
		if v, ok := layer[name]; ok {
			return v, true
		}
	}
	return nil, false
}

func TestFieldResolver(t *testing.T) {
	tpl := MustParse("Hello {{name}}, you are level {{level}}")

	ctx := &testResolver{
		layers: []map[string]interface{}{
			{"name": "Pikachu"},                  // top layer
			{"name": "Bulbasaur", "level": 25},   // bottom layer
		},
	}

	result, err := tpl.Exec(ctx)
	if err != nil {
		t.Fatalf("Exec error: %v", err)
	}

	expected := "Hello Pikachu, you are level 25"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestFieldResolverMissing(t *testing.T) {
	tpl := MustParse("{{name}} {{missing}}")

	ctx := &testResolver{
		layers: []map[string]interface{}{
			{"name": "Pikachu"},
		},
	}

	result, err := tpl.Exec(ctx)
	if err != nil {
		t.Fatalf("Exec error: %v", err)
	}

	expected := "Pikachu "
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestFieldResolverNestedAccess(t *testing.T) {
	tpl := MustParse("{{data.name}}")

	ctx := &testResolver{
		layers: []map[string]interface{}{
			{"data": map[string]interface{}{"name": "Pikachu"}},
		},
	}

	result, err := tpl.Exec(ctx)
	if err != nil {
		t.Fatalf("Exec error: %v", err)
	}

	if result != "Pikachu" {
		t.Errorf("got %q, want %q", result, "Pikachu")
	}
}

func TestFieldResolverDotPath(t *testing.T) {
	// Test a.b.c style access where "a" comes from FieldResolver
	// and "b.c" are resolved from the returned map
	tpl := MustParse("{{rewards.first.chance}}% {{rewards.first.name}}")

	ctx := &testResolver{
		layers: []map[string]interface{}{
			{
				"rewards": map[string]interface{}{
					"first": map[string]interface{}{
						"chance": 85,
						"name":   "Dratini",
					},
				},
			},
		},
	}

	result, err := tpl.Exec(ctx)
	if err != nil {
		t.Fatalf("Exec error: %v", err)
	}

	if result != "85% Dratini" {
		t.Errorf("got %q, want %q", result, "85% Dratini")
	}
}

func TestFieldResolverDotPathAcrossLayers(t *testing.T) {
	// Test where the top-level key comes from one layer
	// but contains nested data
	tpl := MustParse("{{pokemon.name}} {{stats.hp}}")

	ctx := &testResolver{
		layers: []map[string]interface{}{
			{"pokemon": map[string]interface{}{"name": "Pikachu"}},     // layer 1
			{"stats": map[string]interface{}{"hp": 35, "atk": 55}},     // layer 2
		},
	}

	result, err := tpl.Exec(ctx)
	if err != nil {
		t.Fatalf("Exec error: %v", err)
	}

	if result != "Pikachu 35" {
		t.Errorf("got %q, want %q", result, "Pikachu 35")
	}
}

func TestFieldResolverWithBlockHelper(t *testing.T) {
	// Test that block helpers (like #each, #if) work with FieldResolver context
	tpl := MustParse("{{#if name}}Hello {{name}}{{/if}}")

	ctx := &testResolver{
		layers: []map[string]interface{}{
			{"name": "Pikachu"},
		},
	}

	result, err := tpl.Exec(ctx)
	if err != nil {
		t.Fatalf("Exec error: %v", err)
	}

	if result != "Hello Pikachu" {
		t.Errorf("got %q, want %q", result, "Hello Pikachu")
	}
}

func TestFieldResolverWithEach(t *testing.T) {
	// Test {{#each}} where the array comes from FieldResolver
	tpl := MustParse("{{#each items}}{{this.name}} {{/each}}")

	ctx := &testResolver{
		layers: []map[string]interface{}{
			{
				"items": []map[string]interface{}{
					{"name": "Pikachu"},
					{"name": "Eevee"},
				},
			},
		},
	}

	result, err := tpl.Exec(ctx)
	if err != nil {
		t.Fatalf("Exec error: %v", err)
	}

	if result != "Pikachu Eevee " {
		t.Errorf("got %q, want %q", result, "Pikachu Eevee ")
	}
}

func TestFieldResolverLayerPriority(t *testing.T) {
	tpl := MustParse("{{name}} {{form}}")

	ctx := &testResolver{
		layers: []map[string]interface{}{
			{"name": "Pikachu"},                           // top: overrides name
			{"name": "Bulbasaur", "form": "Alolan"},       // bottom: provides form
		},
	}

	result, err := tpl.Exec(ctx)
	if err != nil {
		t.Fatalf("Exec error: %v", err)
	}

	if result != "Pikachu Alolan" {
		t.Errorf("got %q, want %q", result, "Pikachu Alolan")
	}
}
