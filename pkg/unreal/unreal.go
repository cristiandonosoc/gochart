// Package unreal is an utlity package to integrate gochart with Unreal projects.
package unreal

import (
	"context"
	"fmt"
	"time"
)

const (
	GochartExtension = ".gochart"
	FindTimeout      = 1 * time.Minute
)

type unrealIntegrator struct {
	Gocharts []string
}

func NewUnrealIntegrator(ctx context.Context, sourceDir string) (*unrealIntegrator, error) {
	ctx, cancel := context.WithTimeout(ctx, FindTimeout)
	defer cancel()

	gocharts, err := findAllGochartFiles(ctx, sourceDir, GochartExtension)
	if err != nil {
		return nil, fmt.Errorf("finding all gochart files in %q: %w", sourceDir, err)
	}

	return &unrealIntegrator{
		Gocharts: gocharts,
	}, nil
}
