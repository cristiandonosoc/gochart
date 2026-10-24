package unreal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	gCompdb_ubtArgs = []string{
		"-ProjectFiles",
		"-VSCode",
		"-Game",
		"-Engine",
		//"-Progress",
		"-NoIntelliSense",
	}

	gExtraClangFlagsRsp = `
/Zc:inline
/nologo
/Oi
/FC
/c
/Gw
/Gy
/Zm1000
/wd4819
/Zc:hiddenFriend
/Zc:__cplusplus
/D_CRT_STDIO_LEGACY_WIDE_SPECIFIERS=1
/D_SILENCE_STDEXT_HASH_DEPRECATION_WARNINGS=1
/D_WINDLL
/D_DISABLE_EXTENDED_ALIGNED_STORAGE
/source-charset:utf-8
/execution-charset:utf-8
/Ob2
/fastfail
/Ox
/Ot
/GF
/errorReport:prompt
/EHsc
/DPLATFORM_EXCEPTIONS_DISABLED=0
/Z7
/MD
/bigobj
/fp:fast
/Zo
/Zp8
/we4456
/we4458
/we4459
/wd4463
/we4668
/wd4244
/wd4838
/TP
/GR-
/W4
/std:c++20`
)

type compdbEntry struct {
	File      string   `json:"file"`
	Arguments []string `json:"arguments"`
	Directory string   `json:"directory"`
}

func (p *Project) GenerateCompDB() error {
	// Use UBT to generate the VSCode compilation database.
	if err := p.UBT(gCompdb_ubtArgs); err != nil {
		return fmt.Errorf("generating project files: %w", err)
	}

	entries, err := readCompdbEntries(p.Config.ProjectDir, p.Config.ProjectName)
	if err != nil {
		return fmt.Errorf("reading compdb entries: %w", err)
	}
	fmt.Printf("Read %d entries\n", len(entries))

	// Rewrite the flags.
	compdbDir := filepath.Join(p.Config.ProjectDir, ".gunreal")
	if err := os.MkdirAll(compdbDir, 0644); err != nil {
		return fmt.Errorf("creating dir %q: %w", compdbDir, err)
	}

	rspPath, err := writeExtraFlagsRsp(compdbDir)
	if err != nil {
		return fmt.Errorf("writing extra clang flags rsp file: %w", err)
	}

	if err := writeOutCompdb(p.Config.ProjectDir, rspPath, entries); err != nil {
		return fmt.Errorf("writing out compdb: %w", err)
	}

	return nil
}

func readCompdbEntries(projectDir, projectName string) ([]*compdbEntry, error) {
	compdbName := fmt.Sprintf("compileCommands_%s.json", projectName)
	compdbPath := filepath.Join(projectDir, ".vscode", compdbName)

	data, err := os.ReadFile(compdbPath)
	if err != nil {
		return nil, fmt.Errorf("reading %q: %w", compdbPath, err)
	}

	var entries []*compdbEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("unmarshalling compdb entries: %w", err)
	}

	return entries, nil
}

func writeExtraFlagsRsp(dir string) (string, error) {
	rspPath := filepath.Join(dir, "extra_clang_flags.rsp")

	file, err := os.OpenFile(rspPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return "", fmt.Errorf("opening %q: %w", rspPath, err)
	}
	defer file.Close()

	if _, err := file.WriteString(strings.TrimSpace(gExtraClangFlagsRsp)); err != nil {
		return "", fmt.Errorf("writing extra clang flags: %w", err)
	}

	return rspPath, nil
}

func writeOutCompdb(projectDir, rspFilePath string, entries []*compdbEntry) error {
	rspArgument := fmt.Sprintf("@%s", rspFilePath)
	for _, entry := range entries {
		entry.Arguments = append(entry.Arguments, rspArgument)
	}

	content, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling comdb entries: %w", err)
	}

	compdbPath := filepath.Join(projectDir, "compile_commands.json")
	file, err := os.OpenFile(compdbPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("opening %q: %w", compdbPath, err)
	}
	defer file.Close()

	if _, err := file.Write(content); err != nil {
		return fmt.Errorf("writing compdb entries: %w", err)
	}

	fmt.Printf("Wrote compilation database to %s\n", compdbPath)
	return nil
}
