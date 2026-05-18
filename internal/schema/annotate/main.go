// Command annotate rewrites the ADL JSON Schema in-flight at code-generation
// time so atombender/go-jsonschema emits value types (not pointers) for
// optional scalar and enum fields.
//
// The committed internal/schema/schema.json must stay byte-identical to
// upstream inference-gateway/adl (task verify-schema diffs them), so we keep
// the file untouched and feed go-jsonschema this annotated transient copy
// instead.
//
// Rule: for every property of an object schema that is NOT listed in the
// parent's "required" array AND has type string/boolean/integer/number
// (with or without enum) AND has no $ref, inject
// "goJSONSchema": {"pointer": false}.
//
// Nested optional structs ($ref to another definition) are left alone so the
// generator keeps emitting *T for them — callers in internal/generator rely
// on nil-checks against those substructures.
package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: annotate <schema.json>")
		os.Exit(2)
	}

	raw, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "read schema: %v\n", err)
		os.Exit(1)
	}

	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		fmt.Fprintf(os.Stderr, "parse schema: %v\n", err)
		os.Exit(1)
	}

	walk(doc)

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(doc); err != nil {
		fmt.Fprintf(os.Stderr, "encode schema: %v\n", err)
		os.Exit(1)
	}
}

// walk recursively visits every subschema and applies the annotation to
// optional scalar/enum properties of any object schema it encounters.
func walk(node any) {
	switch n := node.(type) {
	case map[string]any:
		annotateObject(n)
		for _, v := range n {
			walk(v)
		}
	case []any:
		for _, v := range n {
			walk(v)
		}
	}
}

// annotateObject inspects a single schema node. If it has a properties map,
// each optional scalar/enum property is annotated with pointer:false.
func annotateObject(node map[string]any) {
	props, ok := node["properties"].(map[string]any)
	if !ok {
		return
	}

	required := map[string]struct{}{}
	if r, ok := node["required"].([]any); ok {
		for _, name := range r {
			if s, ok := name.(string); ok {
				required[s] = struct{}{}
			}
		}
	}

	for name, raw := range props {
		if _, isRequired := required[name]; isRequired {
			continue
		}
		prop, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if _, hasRef := prop["$ref"]; hasRef {
			continue
		}
		if !isScalarType(prop["type"]) {
			continue
		}
		if _, already := prop["goJSONSchema"]; already {
			continue
		}
		prop["goJSONSchema"] = map[string]any{"pointer": false}
	}
}

func isScalarType(t any) bool {
	s, ok := t.(string)
	if !ok {
		return false
	}
	switch s {
	case "string", "boolean", "integer", "number":
		return true
	}
	return false
}
