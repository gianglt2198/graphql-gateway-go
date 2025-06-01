//go:build ignore

package main

import (
	"log"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

func main() {
	ex, err := entgql.NewExtension(
		// Tell Ent to generate a GraphQL schema for
		// the Ent schema in a file named ent.graphql.
		entgql.WithSchemaGenerator(),
		entgql.WithSchemaPath("../../schemas/ent.graphql"),
		entgql.WithConfigPath("../../gqlgen.yml"),
	)
	if err != nil {
		log.Fatalf("creating entgql extension: %v", err)
	}
	opts := []entc.Option{
		entc.Extensions(ex),
	}
	customCollectionTmpl, err := entgql.CollectionTemplate.ParseFiles("collection.tmpl")
	if err != nil {
		log.Fatalf("creating entgql extension collection: %v", err)
	}
	if err := entc.Generate("../../schemas/entities", &gen.Config{
		Target:  "../../generated/ent",
		Package: "github.com/gianglt2198/{package}/generated/ent",
		Templates: []*gen.Template{
			customCollectionTmpl,
			entgql.EnumTemplate,
			entgql.NodeTemplate,
			entgql.PaginationTemplate,
			entgql.TransactionTemplate,
			entgql.EdgeTemplate,
			entgql.MutationInputTemplate,
		},
		Features: []gen.Feature{
			gen.FeatureIntercept,
			gen.FeatureUpsert,
			gen.FeatureVersionedMigration,
		},
	}, opts...); err != nil {
		log.Fatalf("running ent codegen: %v", err)
	}
}
