package unreal

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/cristiandonosoc/gochart/internal/option"
	"github.com/cristiandonosoc/golib/pkg/files"
)

// CleanOrphaned finds all the orphaned files that match the gochart filter and deletes them.
// Empty filter means clean every orpheaned.
// Returns a list of deleted files.
func (ugb *UnrealGochartBackend) CleanOrphaned(ctx context.Context, filter string, opts ...option.Option) ([]string, error) {
	// TODO(cdc): Don't make the options global to the backend as it is now.
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}
	ugb.options = options

	if err := ugb.validate(); err != nil {
		return nil, fmt.Errorf("validating: %w", err)
	}

	orphanedGoutputs, err := ugb.listOrphanedGocharts(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing orphaned gocharts: %w", err)
	}

	if filter != "" {
		slices.DeleteFunc(orphanedGoutputs, func(orphaned *GochartOutput) bool {
			return !strings.Contains(orphaned.Gochart.Name(), filter)
		})
	}

	if len(orphanedGoutputs) == 0 {
		return nil, nil
	}

	var deleted []string
	for _, orphaned := range orphanedGoutputs {
		for _, output := range orphaned.Outputs {
			if err := files.DeleteFile(output.ExpectedPath); err != nil {
				return nil, fmt.Errorf("deleting %q: %w", output.ExpectedPath, err)
			}
			deleted = append(deleted, output.ExpectedPath)
		}
	}

	sort.Strings(deleted)
	return deleted, nil
}
