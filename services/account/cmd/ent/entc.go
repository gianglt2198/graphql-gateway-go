//go:build ignore

package main

import (
	"log"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"

	"github.com/gianglt2198/federation-go/package/modules/db"
)

func main() {
	ex, err := entgql.NewExtension(
		entgql.WithSchemaGenerator(),
		entgql.WithWhereInputs(true),
		entgql.WithSchemaPath("./graphql/schema/ent.gql"),
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
		Target:    "./generated/ent",
		Package:   "github.com/gianglt2198/federation-go/services/account/generated/ent",
		Templates: templates,
		Features: []gen.Feature{
			gen.FeatureIntercept,
			gen.FeatureUpsert,
			gen.FeatureVersionedMigration,
			gen.FeatureEntQL,
			gen.FeatureNamedEdges,
		},
	}, opts...); err != nil {
		log.Fatalf("running ent codegen: %v", err)
	}
}
