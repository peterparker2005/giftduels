package version

// Version — default value, in CI set via ldflags.
//
//	-X "github.com/peterparker2005/giftduels/packages/version-go.Version=v1.2.3"
//
//nolint:gochecknoglobals // fx module pattern
var Version = "dev"
