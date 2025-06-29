//go:build ignore

package main

import (
	"log"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"

	"github.com/gianglt2198/federation-go/package/modules/db"
	"github.com/gianglt2198/federation-go/services/account/cmd/ent/federation"
)

func main() {
	ex, err := entgql.NewExtension(
		entgql.WithSchemaGenerator(),
		entgql.WithConfigPath("./gqlgen.yml"),
		entgql.WithSchemaPath("./graphql/schema/definition.gql"),
		entgql.WithRelaySpec(true),
		entgql.WithSchemaHook(federation.RemoveNodeQueries,
			federation.RemoveMutationInput,
			federation.RemoveEntitiesImplementingNode,
			federation.RemoveEntitiesImplementingOrder,
			federation.RemoveEntitiesImplementingConnection,
		),
	)
	if err != nil {
		log.Fatalf("creating entgql extension: %v", err)
	}
	opts := []entc.Option{
		entc.Extensions(ex),
		db.RepositoryExtention(),
	}

	templates := entgql.AllTemplates
	templates = append(templates, db.PNNIDTemplate)

	if err := entc.Generate("./ent/schema", &gen.Config{
		Target:    "./generated/ent",
		Package:   "github.com/gianglt2198/federation-go/services/account/generated/ent",
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
