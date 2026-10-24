package unreal

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	cpp_gen "github.com/cristiandonosoc/gochart/internal/generation/cpp"
	"github.com/cristiandonosoc/gochart/internal/option"
	backend_common "github.com/cristiandonosoc/gochart/pkg/backend/common"
	"github.com/cristiandonosoc/gochart/internal/files"
	gunreal "github.com/cristiandonosoc/gochart/internal/unreal"

	"golang.org/x/exp/maps"
)

type ListResult struct {
	Gocharts         []*GochartOutput
	OrphanedGocharts []*GochartOutput
}

// List goes over and encounters all the gocharts and the generated files, without generating them.
func (ugb *UnrealGochartBackend) List(ctx context.Context, opts ...option.Option) (*ListResult, error) {
	// TODO(cdc): Don't make the options global to the backend as it is now.
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}
	ugb.options = options

	if err := ugb.validate(); err != nil {
		return nil, fmt.Errorf("validating: %w", err)
	}

	goutputs, err := ugb.listGocharts(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing existing gocharts: %w", err)
	}

	orphanedGoutputs, err := ugb.listOrphanedGocharts(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing orphaned gocharts: %w", err)
	}

	return &ListResult{
		Gocharts:         goutputs,
		OrphanedGocharts: orphanedGoutputs,
	}, nil
}

func (ugb *UnrealGochartBackend) listGocharts(ctx context.Context) ([]*GochartOutput, error) {
	gochartPaths, err := ugb.Project.SearchForFilesByExtension(ctx, ugb.options.Platform, []string{".gochart"})
	if err != nil {
		return nil, fmt.Errorf("searching for gochart files: %w", err)
	}

	if len(gochartPaths) == 0 {
		fmt.Println("No gocharts found.")
		return nil, nil
	}

	goutputs := make([]*GochartOutput, 0, len(gochartPaths))
	for _, gochartPath := range gochartPaths {
		goutput, _, err := ugb.calculateGochartExpectedOutputs(gochartPath)
		if err != nil {
			return nil, fmt.Errorf("calculating gochart expected outputs: %w", err)
		}

		goutputs = append(goutputs, goutput)
	}

	sort.Slice(goutputs, func(i, j int) bool {
		return goutputs[i].Gochart.Name() < goutputs[j].Gochart.Name()
	})

	return goutputs, nil
}

func (ugb *UnrealGochartBackend) listOrphanedGocharts(ctx context.Context) ([]*GochartOutput, error) {
	extensionsMap := map[string]string{
		"gochart.forward_declare.h": "forward_declare",
		"gochart.declare.h":         "declare",
		"gochart.define.h":          "define",
	}

	candidatePaths, err := ugb.Project.SearchForFilesByExtension(ctx, ugb.options.Platform, maps.Keys(extensionsMap))
	if err != nil {
		return nil, fmt.Errorf("searching for gochart files: %w", err)
	}

	if len(candidatePaths) == 0 {
		return nil, nil
	}

	// Load the files.
	// TODO(cdc): This could be done in parallel.
	bos := make([]*backend_common.BackendOutput, 0, len(candidatePaths))
	for _, candidatePath := range candidatePaths {
		var name string
		base := filepath.Base(candidatePath)

		for key, extName := range extensionsMap {
			if strings.HasSuffix(base, key) {
				name = extName
				break
			}
		}

		bo := &backend_common.BackendOutput{
			Name:         name,
			ExpectedPath: candidatePath,
		}

		metadata, err := cpp_gen.ReadMetadataFromBackendOutput(bo)
		if err != nil {
			return nil, fmt.Errorf("reading metadata for %q: %w", candidatePath, err)
		}
		metadata.IsUnrealGochart = true
		bo.Metadata = metadata

		bos = append(bos, bo)
	}

	// Now that we have the backend outputs, we can group them by gochart.
	// We have this little wrapper struct, so that if goutput exists, we ignore it (as we are
	// searching for orphaned ones), but we don't continually keep querying the OS to see if the
	// underlying file exists or not.
	type gochartTracker struct {
		Goutput *GochartOutput
		Exists  bool
	}
	trackerMap := map[string]*gochartTracker{}
	for _, bo := range bos {
		tracker, ok := trackerMap[bo.Metadata.GochartModulePath]
		if !ok {
			// We need to create the goutput.
			// For that we need the gochart path.
			gochartPath, err := bo.Metadata.FullGochartPath(ugb.Project)
			if err != nil {
				return nil, fmt.Errorf("could not get full gochart path for %q: %w", bo.ExpectedPath, err)
			}

			// We create the goutput and add it to the map.
			goutput := &GochartOutput{
				Gochart: &gunreal.File{
					Path: gochartPath,
				},
			}

			// Check if the file exists.
			_, exists, err := files.StatFile(gochartPath)
			if err != nil {
				return nil, fmt.Errorf("statting %q: %w", gochartPath, err)
			}

			// Add it to the tracker map.
			tracker = &gochartTracker{
				Goutput: goutput,
				Exists:  exists,
			}
			trackerMap[bo.Metadata.GochartModulePath] = tracker
		}

		// We only care about gocharts that do not exists.
		if tracker.Exists {
			continue
		}

		// We append the backend output to the goutput.
		tracker.Goutput.Outputs = append(tracker.Goutput.Outputs, bo)
	}

	// Sort them into an array.
	var goutputs []*GochartOutput
	for _, tracker := range trackerMap {
		if tracker.Exists {
			continue
		}
		sort.Slice(tracker.Goutput.Outputs, func(i, j int) bool {
			return tracker.Goutput.Outputs[i].Name < tracker.Goutput.Outputs[j].Name
		})

		goutputs = append(goutputs, tracker.Goutput)
	}
	sort.Slice(goutputs, func(i, j int) bool {
		return goutputs[i].Gochart.Path < goutputs[j].Gochart.Path
	})

	return goutputs, nil
}
