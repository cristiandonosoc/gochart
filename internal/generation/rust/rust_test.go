package rust

import (
	"strings"
	"testing"

	"github.com/cristiandonosoc/gochart/internal/files"
	"github.com/cristiandonosoc/gochart/pkg/backend/common"
	"github.com/cristiandonosoc/gochart/pkg/frontend/gochart_lang"
	"github.com/cristiandonosoc/gochart/pkg/ir"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToSnakeCase(t *testing.T) {
	testcases := map[string]string{
		"Idle":       "idle",
		"ToPickup":   "to_pickup",
		"HTTPServer": "http_server",
		"DeepA1":     "deep_a1",
		"Level2Boss": "level2_boss",
		"already_ok": "already_ok",
		"Snake_Case": "snake_case",
	}
	for input, want := range testcases {
		assert.Equal(t, want, toSnakeCase(input), input)
	}
}

func TestLeadingIdentifiers(t *testing.T) {
	assert.Equal(t, []string{"Vec", "crate", "Leaf"}, leadingIdentifiers("Vec<(crate::Fx, Leaf)>"))
	assert.Equal(t, []string{"super"}, leadingIdentifiers("super::Event"))
	assert.Equal(t, []string{"u8"}, leadingIdentifiers("[u8; 4]"))
}

func generate(t *testing.T, source string) (string, error) {
	t.Helper()

	lf, err := files.NewFromData("test.gochart", []byte(source), true)
	require.NoError(t, err)

	gf, err := gochart_lang.NewGochartLangFrontend()
	require.NoError(t, err)

	gcdata, err := gf.Process(lf)
	require.NoError(t, err)

	gc, err := ir.ProcessGochartData(gcdata)
	require.NoError(t, err)

	return GenerateContent(gc, &common.Metadata{Version: "test"})
}

func TestGenerate(t *testing.T) {
	content, err := generate(t, `
gochart Hauler(language = "rust", derives = "serde::Serialize, Hash") {
	trigger Tick()
	trigger Reserved(job: HaulJob)

	state Free {
		transition Reserved > Hauling do accept
	}

	state Hauling(data = "HaulJob") {
		transition Tick > Free if done
	}
}`)
	require.NoError(t, err)

	for _, want := range []string{
		"use super::*;",
		"pub enum StateId {",
		// Extra derives are added once.
		"#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, serde::Serialize)]",
		"Reserved { job: HaulJob },",
		"fn free_accept(&mut self, job: &HaulJob);",
		"fn hauling_data(&mut self, event: Option<&Event>) -> HaulJob;",
		"fn hauling_done(&self, hauling: &HaulJob) -> bool;",
		"pub fn hauling(&self) -> Option<&HaulJob> {",
		"pub fn hauling_mut(&mut self) -> Option<&mut HaulJob> {",
		"pub fn trigger_reserved<D: Delegate>(delegate: &mut D, job: HaulJob) {",
		"fn rt_dispatch<D: Delegate>(",
	} {
		assert.Contains(t, content, want)
	}
}

func TestGenerateErrors(t *testing.T) {
	testcases := []struct {
		name    string
		source  string
		wantErr string
	}{
		{
			name: "cpp gochart",
			source: `
gochart Door {
	trigger Tick()
	state Closed {
		capture Tick
	}
}`,
			wantErr: `needs language = "rust"`,
		},
		{
			name: "null transition on a parent",
			source: `
gochart Door(language = "rust") {
	trigger Tick()
	state Closed {
		transition null > Open if ready
		state Locked {
			capture Tick
		}
	}
	state Open {
		capture Tick
	}
}`,
			wantErr: "null transitions are only evaluated on leaf states",
		},
		{
			name: "states with the same Rust name",
			source: `
gochart Door(language = "rust") {
	trigger Tick()
	state OpenDoor {
		capture Tick
	}
	state Open_Door {
		capture Tick
	}
}`,
			wantErr: `same Rust name "open_door"`,
		},
		{
			name: "state named like a keyword",
			source: `
gochart Door(language = "rust") {
	trigger Tick()
	state Loop {
		capture Tick
	}
}`,
			wantErr: "is a keyword",
		},
		{
			name: "data type shadowed by a generated type",
			source: `
gochart Door(language = "rust") {
	trigger Tick()
	state Closed(data = "Event") {
		capture Tick
	}
}`,
			wantErr: `type "Event" is shadowed by the generated event enum, use a path instead`,
		},
		{
			name: "callback names that collide",
			source: `
gochart Door(language = "rust") {
	trigger Tick()
	trigger Push(force: u32)
	state Closed {
		transition Tick > Open if ready
		transition Push > Open if ready
	}
	state Open {
		capture Tick
	}
}`,
			wantErr: `callback "closed_ready" is needed with two different signatures`,
		},
		{
			name: "callback named like a delegate method",
			source: `
gochart Door(language = "rust") {
	trigger Tick()
	state On {
		transition Tick > Off do enter_state
	}
	state Off {
		capture Tick
	}
}`,
			wantErr: `callback "on_enter_state"`,
		},
		{
			name: "data getter that collides with a chart method",
			source: `
gochart Door(language = "rust") {
	trigger Tick()
	state IsIn(data = "u32", data_default) {
		capture Tick
	}
}`,
			wantErr: `data getter "is_in" of state "IsIn" collides with a chart query`,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := generate(t, tc.source)
			require.Error(t, err)
			assert.True(t, strings.Contains(err.Error(), tc.wantErr), "error %q should contain %q", err.Error(), tc.wantErr)
		})
	}
}
