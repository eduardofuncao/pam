package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// FormatSpec describes how a dump format maps to a dumper flag and file extension.
type FormatSpec struct {
	Flag string
	Ext  string
}

// DumpOptions configures a native database dump.
type DumpOptions struct {
	Format     string
	OutPath    string
	SchemaOnly bool
	Tables     []string
}

// Dumper performs a native, engine-specific database dump.
type Dumper interface {
	DefaultFormat() string
	Formats() map[string]FormatSpec
	Dump(connString string, opts DumpOptions) error
}

// ResolveFormat determines the dump format from a file extension and/or an
// explicit --format flag. Exactly one of fileExt/formatFlag may be set;
// setting both is an error.
func ResolveFormat(fileExt, formatFlag string, d Dumper) (string, error) {
	extName := ""
	if fileExt != "" {
		ext := fileExt
		if ext[0] == '.' {
			ext = ext[1:]
		}
		for name, spec := range d.Formats() {
			if spec.Ext == ext {
				extName = name
				break
			}
		}
		if extName == "" {
			return "", fmt.Errorf("unknown file extension %q for backup format", fileExt)
		}
	}

	if extName != "" && formatFlag != "" {
		return "", fmt.Errorf("format declared twice; use the filename extension or --format, not both")
	}

	if extName != "" {
		return extName, nil
	}

	if formatFlag != "" {
		if _, ok := d.Formats()[formatFlag]; !ok {
			names := make([]string, 0, len(d.Formats()))
			for name := range d.Formats() {
				names = append(names, name)
			}
			return "", fmt.Errorf("unknown format %q; valid formats: %v", formatFlag, names)
		}
		return formatFlag, nil
	}

	return d.DefaultFormat(), nil
}

// ResolvePath determines the output file path for a dump given a user-supplied
// path (which may be empty, a directory, or a file with/without extension),
// the database name, and the target file extension.
func ResolvePath(path, dbName, ext string) string {
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("%s_%s.%s", dbName, timestamp, ext)

	if path == "" {
		cwd, err := os.Getwd()
		if err != nil {
			cwd = "."
		}
		return filepath.Join(cwd, filename)
	}

	if info, err := os.Stat(path); err == nil && info.IsDir() {
		return filepath.Join(path, filename)
	}

	if filepath.Ext(path) == "" {
		return path + "." + ext
	}

	return path
}
