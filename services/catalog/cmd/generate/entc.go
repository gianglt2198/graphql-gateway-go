//go:build ignore

package main

import (
	"log"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	"github.com/gianglt2198/graphql-gateway-go/pkg/modules/db"
)

func main() {
	ex, err := entgql.NewExtension(
		// Tell Ent to generate a GraphQL schema for
		// the Ent schema in a file named ent.gql.
		entgql.WithSchemaGenerator(),
		entgql.WithWhereInputs(true),
		entgql.WithSchemaPath("./graph/schema/ent.gql"),
		entgql.WithConfigPath("./gqlgen.yml"),
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
		Target:    "./ent",
		Templates: templates,
		Features: []gen.Feature{
			gen.FeatureIntercept,
			gen.FeatureUpsert,
			gen.FeatureVersionedMigration,
		},
	}, opts...); err != nil {
		log.Fatalf("running ent codegen: %v", err)
	}
}
