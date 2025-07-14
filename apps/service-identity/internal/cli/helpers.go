package cli

import (
	"bufio"
	"log/slog"
	"os"
	"strings"

	migratepg "github.com/peterparker2005/giftduels/apps/service-identity/internal/adapter/pg/migrate"
	"github.com/peterparker2005/giftduels/apps/service-identity/internal/config"
)

func newRunner() (*migratepg.Runner, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	return migratepg.NewWithDSN(cfg.Database.DSN())
}

func confirm(prompt, want string) bool {
	slog.Default().Info(prompt)
	in, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	return strings.TrimSpace(in) == want
}

func writeFile(filename, content string) error {
	return os.WriteFile(filename, []byte(content), 0o600)
}
