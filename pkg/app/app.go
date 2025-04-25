package app

import (
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/executor"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/gofiber/fiber"
)

type (
	App struct {
		exec *executor.Executor
		app  *fiber.App
	}
)

func NewApp(es graphql.ExecutableSchema) *App {
	exec := executor.New(es)
	exec.Use(extension.Introspection{})
	exec.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	app := fiber.New()

	_ = func(c *fiber.Ctx) error {

		return nil
	}

	return &App{
		app:  app,
		exec: exec,
	}
}
