package migrate

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	migratepg "github.com/peterparker2005/giftduels/apps/service-gift/internal/adapter/pg/migrate"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/config"
)

func newRunner(cfg *config.Config) (*migratepg.Runner, error) {
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
