package gemini

import (
	"strings"

	"github.com/kinkando/personal-dashboard/internal/tools"
	"google.golang.org/genai"
)

// toolDecls returns Gemini FunctionDeclarations for tools whose names start with prefix.
// Pass an empty prefix to include all tools.
// Schemas are derived from each ToolDef's Input struct via SchemaOf — names,
// descriptions, and field definitions live only in internal/tools/tools.go.
func toolDecls(prefix string) []*genai.FunctionDeclaration {
	defs := tools.All()
	decls := make([]*genai.FunctionDeclaration, 0, len(defs))
	for _, def := range defs {
		if prefix != "" && !strings.HasPrefix(def.Name, prefix) {
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

// AllTools returns FunctionDeclarations for every tool in the registry.
func AllTools() []*genai.FunctionDeclaration {
	return toolDecls("")
}
