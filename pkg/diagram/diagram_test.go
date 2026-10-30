package diagram

import (
	"testing"

	"github.com/cristiandonosoc/gochart/internal/files"
	"github.com/cristiandonosoc/gochart/pkg/frontend/gochart_lang"
	"github.com/cristiandonosoc/gochart/pkg/ir"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func diagram(t *testing.T, source string) string {
	t.Helper()

	lf, err := files.NewFromData("test.gochart", []byte(source), true)
	require.NoError(t, err)
	gf, err := gochart_lang.NewGochartLangFrontend()
	require.NoError(t, err)
	gcdata, err := gf.Process(lf)
	require.NoError(t, err)
	gc, err := ir.ProcessGochartData(gcdata)
	require.NoError(t, err)

	return Mermaid(gc)
}

func TestRustChart(t *testing.T) {
	got := diagram(t, `
gochart Hauler(language = "rust") {
	trigger Tick()
	trigger Reserved(job: u32)
	trigger Reset()

	transition Reset > Idle

	state Free {
		transition Reserved > ToPickup

		state Idle {
			transition Tick > Returning if away do go_home
		}

		state Returning {
			transition null > Idle if arrived
		}
	}

	state Hauling(data = "Vec<u32>") {
		state ToPickup {
			enter
			capture Tick if lost
			transition Tick > Free if done
		}
	}
}`)

	assert.Equal(t, `stateDiagram-v2
    [*] --> Free
    state Free {
        [*] --> Idle
        Idle
        Returning
    }
    state Hauling {
        [*] --> ToPickup
        ToPickup
        ToPickup : enter
        ToPickup : capture Tick [lost]
    }
    Hauling : data Vec#lt;u32#gt;
    state "any state" as __gochart
    __gochart --> Idle : Reset
    Free --> ToPickup : Reserved
    Idle --> Returning : Tick [away] / go_home
    Returning --> Idle : [arrived]
    ToPickup --> Free : Tick [done]
`, got)
}

func TestParallelGochartLevel(t *testing.T) {
	got := diagram(t, `
gochart Machine(parallel) {
	trigger Go()

	capture Go

	state Left {
		state A {
			transition Go > B if { return ready(); }
		}
		state B {
			transition Go > A
		}
	}

	state Right {
		state C {
			exit Go
			transition Go > C do { count++; }
		}
	}
}`)

	assert.Equal(t, `stateDiagram-v2
    [*] --> Machine
    state Machine {
        state Left {
            [*] --> A
            A
            B
        }
        --
        state Right {
            [*] --> C
            C
            C : exit on Go
        }
    }
    state "any state" as __gochart
    __gochart : capture Go
    A --> B : Go […]
    B --> A : Go
    C --> C : Go / …
`, got)
}
