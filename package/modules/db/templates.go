package db

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"embed"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

var (
	//go:embed templates/*.tmpl
	templates embed.FS
)

var (
	// Public Nanoid Template
	PNNIDTemplate = gen.MustParse(gen.NewTemplate("pnnid.tmpl").ParseFS(templates, "templates/pnnid.tmpl"))

	// Repository Template
	RepositoryTemplate = gen.MustParse(gen.NewTemplate("repository.tmpl").ParseFS(templates, "templates/repository.tmpl"))
)

func RepositoryExtention() entc.Option {
	dir := "./generated/ent"
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		fmt.Printf("could not create directory: %v", err)
	}

	hook := func(next gen.Generator) gen.Generator {
		return gen.GenerateFunc(func(g *gen.Graph) error {
			for _, node := range g.Nodes {
				_, skip := node.Annotations["SkipRepository"]
				if skip {
					continue
				}
				path := filepath.Join("./generated/ent", strings.ToLower(node.Name)+"_repository.go")
				f, err := os.Create(path)
				if err != nil {
					return err
				}

				defer f.Close()

				if err := RepositoryTemplate.ExecuteTemplate(f, "repository", node); err != nil {
					return err
				}
			}
			return next.Generate(g)
		})
	}
	return func(cfg *gen.Config) error {
		cfg.Hooks = append(cfg.Hooks, hook)
		return nil
	}
}
