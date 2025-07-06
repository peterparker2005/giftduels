package telegram

import (
	"fmt"
	"strings"
)

// Возвращает display name пользователя: username (если есть), иначе ФИО, иначе пусто.
func GetDisplayName(firstName, lastName, username string) string {
	// 1. Сначала username (по всем Telegram правилам)
	if username != "" {
		return username
	}

	// 2. Потом имя + фамилия, если есть
	fullName := strings.TrimSpace(fmt.Sprintf("%s %s", firstName, lastName))
	if fullName != "" {
		return fullName
	}

	// 3. Fallback — ничего
	return ""
}
