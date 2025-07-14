package telegram

import (
	"fmt"
	"strings"
)

// GetDisplayName returns display name of the user: username (if exists), then first name + last name, otherwise empty.
func GetDisplayName(firstName, lastName, username string) string {
	// 1. First username (according to all Telegram rules)
	if username != "" {
		return username
	}

	// 2. Then first name + last name, if exists
	fullName := strings.TrimSpace(fmt.Sprintf("%s %s", firstName, lastName))
	if fullName != "" {
		return fullName
	}

	// 3. Fallback — nothing
	return ""
}
