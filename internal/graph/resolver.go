package graph

import "database/sql"

// Resolver serves as dependency injection for your app, add any dependencies you need here.
type Resolver struct {
	DB *sql.DB
}
