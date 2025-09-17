//go:build ignore

package main

import (
	"fmt"
	"log"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"

	"github.com/gianglt2198/federation-go/package/modules/db"
	"github.com/gianglt2198/federation-go/package/modules/graphql"
	"github.com/gianglt2198/federation-go/package/utils"
)

func main() {
	ex, err := entgql.NewExtension(
		entgql.WithSchemaGenerator(),
		entgql.WithConfigPath("./gqlgen.yml"),
		entgql.WithRelaySpec(true),
		entgql.WithSchemaPath("./graphql/schema/schema.gql"),
		entgql.WithSchemaHook(
			graphql.RemoveNodeQueries,
			graphql.RemoveMutationInput,
			graphql.RemoveEntitiesImplementingNode,
			graphql.RemoveEntitiesImplementingOrder,
			// graphql.RemoveEntitiesImplementingConnection,
		),
		entgql.WithNodeDescriptor(true),
	)
	if err != nil {
		log.Fatalf("creating entgql extension: %v", err)
	}
	opts := []entc.Option{
		entc.Extensions(ex),
		db.RepositoryExtention(),
	}

	templates := entgql.AllTemplates
	templates = append(templates, db.PNNIDTemplate, db.EdgeTemplate)

	moduleName, err := utils.GetModuleName()
	if err != nil {
		log.Fatalf("could not get module name: %v", err.Error())
	}

	pkg := fmt.Sprintf("%s/generated/ent", moduleName)

	if err := entc.Generate("./ent/schema", &gen.Config{
		Target:    "./generated/ent",
		Package:   pkg,
		Templates: templates,
		Features: []gen.Feature{
			gen.FeatureIntercept,
			gen.FeatureUpsert,
			gen.FeatureVersionedMigration,
			gen.FeatureNamedEdges,
			gen.FeatureExecQuery,
			gen.FeatureBidiEdgeRefs,
			gen.FeatureModifier,
		},
	}, opts...); err != nil {
		log.Fatalf("running ent codegen: %v", err)
	}
}
