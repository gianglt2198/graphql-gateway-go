package federation

import "github.com/vektah/gqlparser/v2/ast"

// EDFSDirectiveSchema contains the GraphQL schema definitions for EDFS directives
const EDFSDirectiveSchema = `directive @edfs_Request(
  subject: String!
  providerId: String! = "default"
) on FIELD_DEFINITION
directive @edfs_Publish(
  subject: String!
  providerId: String! = "default"
) on FIELD_DEFINITION
directive @edfs_Subscribe(
  subjects: [String!]!
  providerId: String! = "default"
  streamConfiguration: edfs_StreamConfiguration
) on FIELD_DEFINITION

input edfs_StreamConfiguration {
  consumerInactiveThreshold: Int! = 30
  consumerName: String!
  streamName: String!
}

type edfs_PublishResult {
  success: Boolean!
}
`

// GetEDFSSchema returns the EDFS directive schema that can be merged with subgraph schemas
func GetEDFSSchema() *ast.Source {
	return &ast.Source{
		Name:    "edfs",
		Input:   EDFSDirectiveSchema,
		BuiltIn: false,
	}
}
