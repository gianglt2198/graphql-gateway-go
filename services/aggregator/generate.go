//go:build ignore

package main

//go:generate go run entgo.io/ent/cmd/ent generate ./ent/schema
//go:generate go run github.com/99designs/gqlgen generate
