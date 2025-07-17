package cli

import (
	"bufio"
	"log/slog"
	"os"
	"strings"
)

func confirm(prompt, want string) bool {
	slog.Default().Info(prompt)
	in, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	return strings.TrimSpace(in) == want
}

func writeFile(filename, content string) error {
	return os.WriteFile(filename, []byte(content), 0o600)
}
