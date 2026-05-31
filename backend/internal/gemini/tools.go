package gemini

import (
	"google.golang.org/genai"
	"github.com/kinkando/personal-dashboard/internal/tools"
)

// AllTools converts the shared tool registry into Gemini FunctionDeclarations.
// Schemas are derived from each ToolDef's Input struct via SchemaOf — names,
// descriptions, and field definitions live only in internal/tools/tools.go.
func AllTools() []*genai.FunctionDeclaration {
	defs := tools.All()
	decls := make([]*genai.FunctionDeclaration, len(defs))
	for i, def := range defs {
		decls[i] = &genai.FunctionDeclaration{
			Name:        def.Name,
			Description: def.Description,
			Parameters:  SchemaOf(def.Input, def.Required...),
		}
	}
	return decls
}
