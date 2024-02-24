package cpp

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"

	"github.com/cristiandonosoc/gochart/pkg/backend/common"
)

const (
	kMetadataHeader_BeginMarker = "GOCHART_HEADER_START"
	kMetadataHeader_EndMarker   = "GOCHART_HEADER_END"

	kMetadataKey_Version      = "GOCHART_VERSION"
	kMetadataKey_UnrealModule = "GOCHART_UNREAL_MODULE"
	kMetadataKey_ModulePath   = "GOCHART_MODULE_PATH"
)

func ReadMetadataFromBackendOutput(bo *common.BackendOutput) (*common.Metadata, error) {
	if err := bo.EnsureLoadFromDisk(); err != nil {
		return nil, fmt.Errorf("ensure loaded file: %w", err)
	}

	lines, err := bo.DiskContent.Lines()
	if err != nil {
		return nil, fmt.Errorf("reading lines from %q: %w", bo.DiskContent.Key, err)
	}

	return parseLines(lines)
}

func ReadMetadata(content string) (*common.Metadata, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning file line by line: %w", err)
	}

	return parseLines(lines)
}

func parseLines(lines []string) (*common.Metadata, error) {
	foundBeginMarker := false
	foundEndMarker := false

	metadata := &common.Metadata{}

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
			if strings.Contains(line, kMetadataHeader_BeginMarker) {
				foundBeginMarker = true
			}

			continue
		}

		// Check if we're not done with the header.
		if strings.Contains(line, kMetadataHeader_EndMarker) {
			foundEndMarker = true
			break
		}

		// Otherwise we try to extract the value out of this line.
		if err := parseMetadataLine(metadata, line); err != nil {
			return nil, fmt.Errorf("parsing metadata line: %w", err)
		}
	}

	if !foundBeginMarker {
		fmt.Printf("WARNING: did not find %q header marker\n", kMetadataHeader_BeginMarker)
	} else if !foundEndMarker {
		fmt.Printf("WARNING: did not find %q header marker\n", kMetadataHeader_EndMarker)
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
	case kMetadataKey_Version:
		metadata.Version = value
	case kMetadataKey_UnrealModule:
		metadata.UnrealModule = value
	case kMetadataKey_ModulePath:
		metadata.GochartModulePath = value
	default:
		return fmt.Errorf("unknown header key %q", key)
	}

	return nil
}
