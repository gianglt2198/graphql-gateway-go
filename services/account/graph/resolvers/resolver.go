package resolvers

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	// federatedSDL string
}

// func (r *Resolver) Service() *ServiceResolver {
// 	return &ServiceResolver{r}
// }

// type ServiceResolver struct {
// 	*Resolver
// }

// func (r *ServiceResolver) SDL(ctx context.Context) (string, error) {
// 	// Cache the SDL to avoid reading the file for every request
// 	if r.federatedSDL == "" {
// 		// Read the schema file
// 		sdl, err := loadAllSchemaFiles("graph/schemas")
// 		if err != nil {
// 			return "", err
// 		}
// 		r.federatedSDL = sdl
// 	}
// 	return r.federatedSDL, nil
// }

// func loadAllSchemaFiles(schemaDir string) (string, error) {
// 	var schemaContent strings.Builder

// 	err := filepath.WalkDir(schemaDir, func(path string, d fs.DirEntry, err error) error {
// 		if err != nil {
// 			return err
// 		}

// 		if d.IsDir() {
// 			return nil
// 		}

// 		if !strings.HasSuffix(d.Name(), ".graphqls") && !strings.HasSuffix(d.Name(), ".gql") {
// 			return nil
// 		}

// 		content, err := os.ReadFile(path)
// 		if err != nil {
// 			return fmt.Errorf("error reading schema file %s: %w", path, err)
// 		}

// 		schemaContent.Write(content)
// 		schemaContent.WriteString("\n")
// 		return nil
// 	})

// 	if err != nil {
// 		return "", fmt.Errorf("error walking schema directory: %w", err)
// 	}

// 	return schemaContent.String(), nil
// }
