# Options

These are the state options available for Gochart, States and other elements of Gochart.

> [!NOTE]
> This fill is automatically generated from the contents of the actual options code located at
> [pkg/ir/options.go](/pkg/ir/options.go).

## Option Kinds

Options are mostly located within parenthesis after the element they are defined
There are 3 kinds of values:

### Boolean Options

This is a simple on/off kind of flag. You can define it as either true or false:

```
// Just stating the flag name is the same as writing the explicit value for true.
// The two following are identical.
state Foo(boolean_flag) { ... }
state Foo(boolean_flag = true) { ... }

// For disabling the flag:
state Foo(boolean_flag = false) { ... }
```

### Integer Options

These are flags intended to support a numeric value associated with them.
If declared, they must have a value associated to them.

```
state Foo(int_flag = 22) { ... }
```

### String Options

These have a string associated with them.
If declared, they must have a value associated with them.

```
state Foo(string_flag = "string_value!") { ... }
```

### Multiple Options

An element can have multiple options at the same time. They are separated by comma:

```
state Foo(boolean_without_value, boolean_with_value = false, int_flag = 33, string_flag = "value!") {
	...
}
```

## Gochart-level options

These are options that are supported on the Gochart element, which defines the Gochart as a whole.

### derives

| Name | Description |
| --- | --- |
| Flag | derives |
| Kind | string |

**Description**

Extra derives for the generated Rust types, comma separated (Rust only).

The generated chart always derives Clone, Debug, PartialEq and Eq. This option adds more, such as
serde's, so the chart can be saved or hashed as part of its owner.

**Synopsis:**
```
gochart Hauler(language = "rust", derives = "serde::Serialize, serde::Deserialize") {
	...
}
```

### language

| Name | Description |
| --- | --- |
| Flag | language |
| Kind | string |

**Description**

The language the gochart is written for: "cpp" (default) or "rust".

It determines how trigger arguments are parsed and which backend can generate the gochart.
Rust gocharts use named callbacks ("if some_condition", "do some_action") instead of inlined code,
and support state data (see the state "data" option).

See https://github.com/cristiandonosoc/gochart/docs/rust.md for more details.

**Synopsis:**
```
gochart Hauler(language = "rust") {
	trigger Reserved(job: HaulJob)
	...
}
```

### namespace

| Name | Description |
| --- | --- |
| Flag | namespace |
| Kind | string |

**Description**

The name of the namespace to generate in |gochart::<NAMESPACE>|.
Otherwise the name of gochart will be used.

**Synopsis:**
```
gochart(namespace = "SomeOtherNamespace") {
	...
}
```

### parallel

| Name | Description |
| --- | --- |
| Flag | parallel |
| Kind | boolean |

**Description**

Whether the root states of the gochart should be considered running in parallel within each other.

When this option is defined, all child states will run as separate parallel gochart groups and the
gochart is considered to be in all those states simultaneously.

See https://github.com/cristiandonosoc/gochart/docs/concurrent.md for more details.

**Synopsis:**
```
gochart(parallel) {
	...
	state ParallelChild1 {
		...
	}

	state ParallelChild2 {
		...
	}
}
```

## State-level options

These are options that are supported on the State element and change the behavior for that state only.

### data

| Name | Description |
| --- | --- |
| Flag | data |
| Kind | string |

**Description**

The type of the data that lives while this state is active (Rust only).

The data is created when the state is entered (by a callback, or with Default::default() if
"data_default" is set) and dropped when the state is exited. Callbacks of the state and its
children receive it.

**Synopsis:**
```
state Hauling(data = "HaulJob") {
	...
}
```

### data_default

| Name | Description |
| --- | --- |
| Flag | data_default |
| Kind | boolean |

**Description**

Create the state data with Default::default() instead of a constructor callback (Rust only).

Requires the "data" option.

**Synopsis:**
```
state Loading(data = "Transfer", data_default) {
	...
}
```

### default

| Name | Description |
| --- | --- |
| Flag | default |
| Kind | boolean |

**Description**

Whether this particular state should be considered the default entry point for the parent state for the parent state.

By default, the first child is considered to be the default entry point, but this can be override with this option.
There can only be one default child per state.

While valid, it is redundant to set the default flag for the first child.

**Synopsis:**
```
state Parent {
	state Child1 {
		...
	}

	state Child2(default) {
		...
	}
}
```

### parallel

| Name | Description |
| --- | --- |
| Flag | parallel |
| Kind | boolean |

**Description**

Whether the child states of this state should be considered running in parallel within each other.

When this option is defined, all child states will run as separate parallel gochart groups and the
gochart is considered to be in all those states simultaneously.

See https://github.com/cristiandonosoc/gochart/docs/concurrent.md for more details.

**Synopsis:**
```
state SomeState(parallel) {
	...
	state ParallelChild1 {
		...
	}

	state ParallelChild2 {
		...
	}
}
```
