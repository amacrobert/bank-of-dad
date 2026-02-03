package family

import (
	"bank-of-dad/internal/store"
	"fmt"
	"strings"
)

func ValidateSlug(slug string) error {
	return store.ValidateSlug(slug)
}

func ValidateChildPassword(password string) error {
	if len(password) < 6 {
		return fmt.Errorf("password must be at least 6 characters")
	}
	return nil
}

func ValidateChildName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return fmt.Errorf("first name is required")
	}
	if trimmed != name {
		return fmt.Errorf("first name cannot have leading or trailing whitespace")
	}
	if strings.ContainsAny(name, "<>&\"'") {
		return fmt.Errorf("first name contains invalid characters")
	}
	if len(name) > 50 {
		return fmt.Errorf("first name must be 50 characters or less")
	}
	return nil
}
