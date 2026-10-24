package unreal

import (
	"fmt"
	"strings"
)

type Platform string

const (
	Platform_Windows = "Win64"
)

// NewUnrealPlatform attempts to unify the unreal platform from identifiers that might come from the
// outside.
func NewUnrealPlatform(id string) (Platform, error) {
	switch strings.ToLower(id) {
	case "win64", "windows":
		return Platform_Windows, nil
	default:
		return "", fmt.Errorf("unrecognized unreal platform %q", id)
	}
}

func (up *Platform) String() string {
	return string(*up)
}
