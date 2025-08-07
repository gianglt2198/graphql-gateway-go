package graphql

import (
	"embed"
)

var (
	//go:embed schemas
	schemas embed.FS
)

var listSchemas = []string{
	"schemas/definition.gql",
	"schemas/user/user.gql",
}

func GetAllSchemas() []byte {
	contents := make([]byte, 0)

	for _, schema := range listSchemas {
		data, err := schemas.ReadFile(schema)
		if err != nil {
			panic(err)
		}
		contents = append(contents, data...)
		contents = append(contents, '\n') // Add newline between schemas
	}
	// Ensure the final content is a valid GraphQL schema
	contents = append(contents, '\n')
	return contents
}
