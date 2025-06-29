package federation

import (
	"errors"
	"strings"

	"entgo.io/ent/entc/gen"

	"github.com/vektah/gqlparser/v2/ast"
)

var (
	RemoveNodeQueries = func(_ *gen.Graph, s *ast.Schema) error {
		q, ok := s.Types["Query"]
		if !ok {
			return errors.New("failed to find query definition in schema")
		}

		fields := ast.FieldList{}

		for _, f := range q.Fields {
			if f.Name != "sessions" && f.Name != "users" {
				fields = append(fields, f)
			}
		}
		q.Fields = fields

		return nil
	}

	RemoveMutationInput = func(_ *gen.Graph, s *ast.Schema) error {
		for name, typ := range s.Types {
			if typ.Kind == ast.InputObject &&
				(strings.HasPrefix(name, "Create") && strings.HasSuffix(name, "Input") ||
					strings.HasPrefix(name, "Update") && strings.HasSuffix(name, "Input")) {
				delete(s.Types, name)
			}
		}
		return nil
	}

	RemoveEntitiesImplementingNode = func(_ *gen.Graph, s *ast.Schema) error {
		for name, typ := range s.Types {
			for _, iface := range typ.Interfaces {
				if iface == "Node" {
					delete(s.Types, name)
				}
			}
		}
		return nil
	}

	RemoveEntitiesImplementingOrder = func(_ *gen.Graph, s *ast.Schema) error {
		for name, typ := range s.Types {
			if typ.Kind == ast.InputObject && (strings.Contains(name, "Order")) || (typ.Kind == ast.Enum && name != "OrderDirection") {
				delete(s.Types, name)
			}
		}
		return nil
	}

	RemoveEntitiesImplementingConnection = func(_ *gen.Graph, s *ast.Schema) error {
		for name, typ := range s.Types {
			if typ.Kind == ast.Object && (strings.Contains(name, "Connection") || strings.Contains(name, "Edge")) {
				delete(s.Types, name)
			}
		}
		return nil
	}
)
