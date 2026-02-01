package image

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// ParsePlatform parses "os/arch" into a v1.Platform.
// Returns the host platform if s is empty.
func ParsePlatform(s string) (v1.Platform, error) {
	if s == "" {
		return v1.Platform{OS: "linux", Architecture: runtime.GOARCH}, nil
	}
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 {
		return v1.Platform{}, fmt.Errorf("invalid platform %q, expected os/arch", s)
	}
	return v1.Platform{OS: parts[0], Architecture: parts[1]}, nil
}

// LoadImage resolves an image reference, trying the local Docker daemon first,
// then falling back to a remote registry.
func LoadImage(ref string, platform v1.Platform) (v1.Image, error) {
	parsed, err := name.ParseReference(ref)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", ref, err)
	}

	// Try local daemon first
	img, err := daemon.Image(parsed)
	if err == nil {
		cfg, cfgErr := img.ConfigFile()
		if cfgErr == nil && cfg.Architecture == platform.Architecture && cfg.OS == platform.OS {
			return img, nil
		}
		// Platform mismatch or config error — fall through to remote
	}

	// Fallback to remote registry
	img, err = remote.Image(parsed,
		remote.WithAuthFromKeychain(authn.DefaultKeychain),
		remote.WithPlatform(platform),
	)
	if err != nil {
		return nil, fmt.Errorf("load %s: %w", ref, err)
	}
	return img, nil
}
