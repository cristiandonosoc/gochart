package unreal

import (
	"encoding/json"
	"fmt"
	"os"
)

type UProjectModule struct {
	Name                   string   `json:"Name"`
	Type                   string   `json:"Type"`
	LoadingPhase           string   `json:"LoadingPhase"`
	AdditionalDependencies []string `json:"AdditionalDependencies"`
}

type UProjectPlugin struct {
	Name            string   `json:"Name"`
	Enabled         bool     `json:"Enabled"`
	TargetAllowList []string `json:"TargetAllowList"`
	WorkspaceURL    string   `json:"WorkspaceURL"`
}

type UProject struct {
	FileVersion       int `json:"FileVersion"`
	EngineAssociation string `json:"EngineAssociation"`
	Category          string `json:"Category"`
	Description       string `json:"Description"`

	Modules []*UProjectModule `json:"Modules"`
	Plugins []*UProjectPlugin `json:"Plugins"`
}

func loadUProjectFile(path string) (*UProject, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %q: %w", path, err)
	}

	uproject := &UProject{}
	if err := json.Unmarshal(data, uproject); err != nil {
		return nil, fmt.Errorf("unmarshalling uproject: %w", err)
	}

	return uproject, nil
}
