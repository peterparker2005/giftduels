package migrate

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	migratepg "github.com/peterparker2005/giftduels/apps/service-duel/internal/adapter/pg/migrate"
	"github.com/peterparker2005/giftduels/apps/service-duel/internal/config"
)

func newRunner() (*migratepg.Runner, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	return migratepg.New(cfg)
}

func confirm(prompt, want string) bool {
	fmt.Print(prompt)
	in, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	return strings.TrimSpace(in) == want
}

func writeFile(filename, content string) error {
	return os.WriteFile(filename, []byte(content), 0o644)
}
