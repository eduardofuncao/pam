package backup

import "fmt"

// CreateDumper returns the Dumper implementation for the given database type.
func CreateDumper(dbType string) (Dumper, error) {
	switch dbType {
	case "postgres", "postgresql":
		return newPostgresDumper(), nil
	default:
		return nil, fmt.Errorf("native backup not implemented for %s", dbType)
	}
}
