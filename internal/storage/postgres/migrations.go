package postgres

import (
	"os"
	"path/filepath"
)

const OperationalFoundationMigration = "supabase/migrations/202606200005_operational_foundation.sql"

func ReadOperationalFoundationMigration(repoRoot string) (string, error) {
	data, err := os.ReadFile(filepath.Join(repoRoot, OperationalFoundationMigration))
	if err != nil {
		return "", err
	}
	return string(data), nil
}
