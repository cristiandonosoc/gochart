package cpp

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cristiandonosoc/gochart/pkg/backend/common"
)

const (
	kHeaderBeginMarker = "GOCHART_HEADER_START"
	kHeaderEndMarker   = "GOCHART_HEADER_END"

	kVersionKey = "GOCHART_VERSION"
)

func LoadMetadata(bo *common.BackendOutput) error {
	if bo.Metadata != nil {
		return nil
	}

	if !bo.IsLoadedFromDisk() {
		if err := bo.LoadFromDisk(); err != nil {
			return fmt.Errorf("loading file: %w", err)
		}
	}

	metadata, err := parseLines(bo)
	if err != nil {
		return fmt.Errorf("parsing lines for metadata: %w", err)
	}

	bo.Metadata = metadata
	return nil
}

func parseLines(bo *common.BackendOutput) (*common.Metadata, error) {
	foundBeginMarker := false
	foundEndMarker := false

	metadata := &common.Metadata{}

	lines, err := bo.DiskContent.Lines()
	if err != nil {
		return nil, fmt.Errorf("reading lines from %q: %w", bo.DiskContent.Key, err)
	}

	for _, line := range lines {
		line := strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// We only want to look at the beginning comments. As soon as we don't find out, we're out.
		if !strings.HasPrefix(line, "//") {
			break
		}

		// Search for the begin marker.
		if !foundBeginMarker {
			if strings.Contains(line, kHeaderBeginMarker) {
				foundBeginMarker = true
			}

			continue
		}

		// Check if we're not done with the header.
		if strings.Contains(line, kHeaderEndMarker) {
			foundEndMarker = true
			break
		}

		// Otherwise we try to extract the value out of this line.
		if err := parseMetadataLine(metadata, line); err != nil {
			return nil, fmt.Errorf("parsing metadata line: %w", err)
		}
	}

	if !foundBeginMarker {
		fmt.Printf("WARNING: File %q: did not find %q header marker\n", bo.ExpectedPath, kHeaderBeginMarker)
	} else if !foundEndMarker {
		fmt.Printf("WARNING: File %q: did not find %q header marker\n", bo.ExpectedPath, kHeaderEndMarker)
	}

	return metadata, nil
}

var identifier = "[a-zA-Z0-9_]"
var regex = regexp.MustCompile(fmt.Sprintf(`^// \$(%s+)="(.+)"`, identifier))

func parseMetadataLine(metadata *common.Metadata, line string) error {
	matches := regex.FindStringSubmatch(line)

	if len(matches) == 0 {
		return nil
	}

	key := matches[1]
	value := matches[2]

	switch key {
	case kVersionKey:
		metadata.Version = value
	default:
		return fmt.Errorf("unknown header key %q", key)
	}

	return nil
}
