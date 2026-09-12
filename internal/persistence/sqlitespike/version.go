package sqlitespike

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

func (s *Store) SQLiteVersion(ctx context.Context) (string, error) {
	var version string
	if err := s.db.QueryRowContext(ctx, `SELECT sqlite_version()`).Scan(&version); err != nil {
		return "", fmt.Errorf("read sqlite runtime version: %w", err)
	}
	if strings.TrimSpace(version) == "" {
		return "", errors.New("sqlite runtime version is empty")
	}
	return version, nil
}
