# Rust Backend

The Rust backend generates a single `.rs` file from a gochart written with `language = "rust"`. The
generated code has no dependencies (not even on a gochart runtime crate), does not allocate, and the
chart is a plain value: it can be cloned, compared, hashed and serialized as part of its owner. This
makes it usable in deterministic simulations, where the chart state is saved and hashed like the
rest of the game state.

```
gochart backend rust generate hauler.gochart src/settlement hauler
```

Generates `src/settlement/hauler.gochart.rs`, which is included as a module:

```rust
#[path = "hauler.gochart.rs"]
mod hauler;
```

Names used in the gochart (data and argument types) are resolved in the module that includes the
file (the generated code starts with `use super::*;`).

## Bazel

`defs.bzl` has a rule that runs the tool. It also works for Bzlmod projects that depend on gochart
(see [MODULE.bazel](/MODULE.bazel)):

```python
load("@gochart//:defs.bzl", "generate_gochart_rust")

generate_gochart_rust(
    name = "hauler_gochart",
    src = "hauler.gochart",
    file_prefix = "hauler",
)

rust_library(
    name = "sim",
    srcs = glob(["src/**/*.rs"]) + [":hauler_gochart"],
    ...
)
```

The generated file is placed next to the other sources of the crate, so `#[path]` finds it.

## The Language

Rust gocharts use the same language as C++ ones (see [the docs](gochart.md)), with these
differences:

- **Trigger arguments** are written as in Rust: `trigger Reserved(job: HaulJob, route: Vec<Fx>)`.
  References and lifetimes are not allowed, as the arguments are stored in an owned `Event`.
- **No inlined code.** Conditions and actions are named callbacks that the delegate implements:
  `transition Tick > Loading if docked`, `transition Reserved > ToPickup do accept`,
  `capture Tick if busy`.
- **State data.** `state Hauling(data = "HaulJob")` gives the state a value that is created when the
  state is entered and dropped when it is exited. The callbacks of the state and of its children
  receive it. With `data_default`, the data is created with `Default::default()` instead of a
  callback.
- **Derives.** The generated types derive `Clone, Debug, PartialEq, Eq`. The `derives` option adds
  more: `gochart Hauler(language = "rust", derives = "serde::Serialize, serde::Deserialize")`. The
  data types must implement the same traits.

Example:

```
gochart Hauler(language = "rust", derives = "serde::Serialize, serde::Deserialize") {
	trigger Tick()
	trigger Reserved(job: HaulJob)

	state Free(data = "FreeData", data_default) {
		transition Reserved > ToPickup do accept

		state Idle {
			transition Tick > Returning if away_from_home
		}

		state Returning {
			transition Tick > Idle if route_done
		}
	}

	state Hauling(data = "HaulJob") {
		state ToPickup {
			transition Tick > Loading if docked
		}

		state Loading(data = "Transfer") {
			transition Tick > Idle if loaded_all
		}
	}
}
```

## The Generated Code

| Item | What it is |
| --- | --- |
| `Chart` | The state: which states are active and their data. A default chart is inactive (or in its default states, with `default_active`). `Chart::INACTIVE` is an inactive chart. |
| `Delegate` | The trait with the callbacks. The chart lives in the delegate (`fn chart(&mut self)`). |
| `activate`, `deactivate` | Enter the default states, exit every state. |
| `trigger_<name>(delegate, args...)` | Send a trigger. `dispatch(delegate, Event)` does the same with an `Event`. Both return whether a state handled it. |
| `StateId`, `Leaf`, `TriggerId` | Enums of every state, of the leaf states and of the triggers, with `ALL` and `name()`. |
| `Event` | A trigger with its arguments. `Event::id()` returns its `TriggerId`. |
| `REGION_COUNT` | The number of parallel regions. |

Queries on `Chart`:

- `is_active()`, `is_in(StateId)`.
- `leaf()`: the current leaf (only for charts without parallel states).
- `leaves()`: the current leaf of each parallel region.
- `<state>()` and `<state>_mut()`: the data of a state, if it is active (eg. `chart.hauling()`).

### Callbacks

Callbacks are named after the state (in snake case, `gochart` for the gochart level). Every callback
receives the data of the states from the top level down to the state that declares it, as
`&mut` (or `&` for conditions), followed by the trigger arguments (as `&`).

| Declaration in state `Loading` | Callback |
| --- | --- |
| `enter` | `fn loading_enter(&mut self, hauling: &mut HaulJob, loading: &mut Transfer)` |
| `enter Tick` | `fn loading_enter_on_tick(&mut self, ...)` |
| `exit`, `exit Tick` | `fn loading_exit(...)`, `fn loading_exit_on_tick(...)` |
| `capture Loaded` | `fn loading_capture_loaded(&mut self, ..., units: &u32)` |
| `... if loaded_all` | `fn loading_loaded_all(&self, hauling: &HaulJob, loading: &Transfer) -> bool` |
| `... do finish` | `fn loading_finish(&mut self, ...)` |
| `state Loading(data = "Transfer")` | `fn loading_data(&mut self, hauling: &mut HaulJob, event: Option<&Event>) -> Transfer` |

The data constructor receives the data of the parents and the trigger that caused the entry (`None`
on activation and for null transitions).

The delegate also has three hooks with default empty implementations, called for every state of
every chart. They are meant for generic tooling (inspectors, logs, animation events):

- `on_enter_state(StateId)`, after a state is entered.
- `on_exit_state(StateId)`, after a state is exited.
- `on_transition(from, to, trigger)`, after a transition entered its states.

While a trigger is processed, the chart is taken out of the delegate (`chart()` returns an inactive
chart). Callbacks get the data they need as arguments instead.

#### Delegates for some triggers only

A chart can be sent triggers from places that have different things at hand, for example a game
tick where the agent sees the world and a later phase that only touches the chart. With the
`delegate_defaults` option, every callback gets a default body that panics with its name, so each
of those places can use a delegate that implements only the callbacks its triggers reach:

```
gochart Worker(language = "rust", delegate_defaults) {
	...
}
```

```rust
/// Only sends Reserved, which enters Busy.
struct Apply<'a>(&'a mut worker::Chart);

impl worker::Delegate for Apply<'_> {
    fn chart(&mut self) -> &mut worker::Chart { self.0 }
    fn busy_data(&mut self, event: Option<&Event>) -> u32 { ... }
}
```

`chart()` stays required. Without the option, the compiler checks that a delegate implements every
callback. With it, a missing one panics when it is reached:
`gochart Worker: idle_bored is not implemented by this delegate`.

#### Charts that start active

A default chart is inactive, so an owner that derives `Default` gets a chart that needs `activate`
(and a delegate) before it is used. With the `default_active` option, `Chart::default()` is the
chart in its default states instead:

```
gochart Door(language = "rust", default_active) {
	...
}
```

```rust
#[derive(Default)]
struct House {
    chart: door::Chart, // Already in Shut > Latched.
}
```

Entering the default states must not call any callback, as there is no delegate to call: no enter
reactions on the way in (at the gochart level or in a default state), state data only through
`data_default`, and no null transitions on the default leaves. The generator rejects the gochart
otherwise, naming the callbacks. Callbacks elsewhere (enter reactions of states entered later, data
of other states) are fine. The `on_enter_state` hook isn't called for the default states either.
`activate` does nothing on an active chart, and `deactivate` still leaves the chart inactive.

## Semantics

- **Dispatch.** A trigger is offered to each current leaf, then to its parents, until one handles
  it. In a state, transitions and captures are evaluated in the order they are written; a failed
  condition lets the trigger through to the next entry, then to the parent. In parallel charts,
  every region does this, but a state is offered a trigger at most once.
- **Transitions.** The action runs first (the source states still have their data). Then the
  states between the source and the target are exited (deepest first) and the ones down to the
  target are entered (and its default children):
  - If the source contains the target, the source stays active and keeps its data.
  - Otherwise both are left up to their lowest common parent. Transitions to the source itself or
    to one of its parents exit and enter those states again.
  - In a parallel state, only the region of the transition is exited and entered.
- **Reactions.** Entering a state creates its data, then calls the enter reaction for the trigger
  (if any), then the default one. Exiting is the same with exit reactions, and then drops the data.
- **Null transitions** are evaluated when a leaf state is entered, once the transition that entered
  it is done. They are only allowed on leaf states.
- **Callbacks cannot send triggers.** There is no trigger queue. Chains of triggers are written as
  conditions on the next trigger (eg. a `Tick` trigger with `transition Tick > Done if finished`) or
  as null transitions.
- **Handled triggers.** `dispatch` and `trigger_<name>` return true when a transition or a capture
  took the trigger (its condition, if any, passed), in any region. A trigger that no state takes
  changes nothing and returns false.
- **Inactive charts** ignore triggers (and return false).

## Differences with the C++ Backend

- Routes between states at different depths below a common parent don't exit and enter the parent.
- Several conditional transitions for the same trigger in a state are evaluated in order. The C++
  backend only evaluates the first one.
- In parallel charts, a parent reached by several regions is offered the trigger once.
- Callbacks cannot send triggers (see above).
