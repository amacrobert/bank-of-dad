package repositories

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

// isNotFound checks whether the error is a GORM record-not-found error.
func isNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// isDuplicateKey checks whether the error is a PostgreSQL unique constraint violation.
func isDuplicateKey(err error) bool {
	return err != nil && strings.Contains(err.Error(), "duplicate key")
}
