package rust

import (
	"strings"
	"unicode"
)

// kRustKeywords are the identifiers that cannot be used as plain Rust names.
var kRustKeywords = map[string]struct{}{
	"as": {}, "async": {}, "await": {}, "break": {}, "const": {}, "continue": {}, "crate": {},
	"dyn": {}, "else": {}, "enum": {}, "extern": {}, "false": {}, "fn": {}, "for": {}, "if": {},
	"impl": {}, "in": {}, "let": {}, "loop": {}, "match": {}, "mod": {}, "move": {}, "mut": {},
	"pub": {}, "ref": {}, "return": {}, "self": {}, "Self": {}, "static": {}, "struct": {},
	"super": {}, "trait": {}, "true": {}, "type": {}, "unsafe": {}, "use": {}, "where": {},
	"while": {}, "abstract": {}, "become": {}, "box": {}, "do": {}, "final": {}, "gen": {},
	"macro": {}, "override": {}, "priv": {}, "try": {}, "typeof": {}, "unsized": {},
	"virtual": {}, "yield": {},
}

func isRustKeyword(s string) bool {
	_, ok := kRustKeywords[s]
	return ok
}

// toSnakeCase converts a CamelCase name into snake_case: "ToPickup" -> "to_pickup",
// "HTTPServer" -> "http_server".
func toSnakeCase(name string) string {
	runes := []rune(name)
	var b strings.Builder
	for i, r := range runes {
		if unicode.IsUpper(r) {
			// We start a new word when coming from a lowercase letter or a digit, or at the last
			// capital of an acronym ("HTTPServer": the "S").
			if i > 0 {
				prev := runes[i-1]
				nextIsLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])
				if unicode.IsLower(prev) || unicode.IsDigit(prev) || (unicode.IsUpper(prev) && nextIsLower) {
					if !strings.HasSuffix(b.String(), "_") {
						b.WriteRune('_')
					}
				}
			}
			b.WriteRune(unicode.ToLower(r))
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// leadingIdentifiers returns the identifiers of a Rust type expression that are resolved by name in
// the current scope, ie. the ones that are not after a "::". Eg.
// "Vec<(crate::Fx, Leaf)>" -> "Vec", "crate", "Leaf".
func leadingIdentifiers(typeExpr string) []string {
	var result []string
	runes := []rune(typeExpr)
	for i := 0; i < len(runes); {
		r := runes[i]
		if !(unicode.IsLetter(r) || r == '_') {
			i++
			continue
		}

		start := i
		for i < len(runes) && (unicode.IsLetter(runes[i]) || unicode.IsDigit(runes[i]) || runes[i] == '_') {
			i++
		}

		prefix := strings.TrimRight(string(runes[:start]), " \t")
		if !strings.HasSuffix(prefix, "::") {
			result = append(result, string(runes[start:i]))
		}
	}
	return result
}
