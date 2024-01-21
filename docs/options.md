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
