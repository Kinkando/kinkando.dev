package gemini

import (
	"strings"

	"github.com/kinkando/personal-dashboard/internal/tools"
	"google.golang.org/genai"
)

// toolDecls returns Gemini FunctionDeclarations for tools whose names start with
// any of the given prefixes. Pass a single empty string to include all tools.
// Schemas are derived from each ToolDef's Input struct via SchemaOf — names,
// descriptions, and field definitions live only in internal/tools/tools.go.
func toolDecls(prefixes ...string) []*genai.FunctionDeclaration {
	defs := tools.All()
	decls := make([]*genai.FunctionDeclaration, 0, len(defs))
	for _, def := range defs {
		if !matchesAnyPrefix(def.Name, prefixes) {
			continue
		}
		decls = append(decls, &genai.FunctionDeclaration{
			Name:        def.Name,
			Description: def.Description,
			Parameters:  SchemaOf(def.Input, def.Required...),
		})
	}
	return decls
}

// matchesAnyPrefix reports whether name starts with at least one of the given
// prefixes. An empty string prefix matches everything.
func matchesAnyPrefix(name string, prefixes []string) bool {
	for _, p := range prefixes {
		if p == "" || strings.HasPrefix(name, p) {
			return true
		}
	}
	return false
}

// AllTools returns FunctionDeclarations for every tool in the registry.
func AllTools() []*genai.FunctionDeclaration {
	return toolDecls("")
}
