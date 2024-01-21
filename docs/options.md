# Options

These are the state options available for Gochart, States and other elements of Gochart.

> [!NOTE]
> This fill is automatically generated from the contents of the actual options code located at
> [pkg/ir/options.go](/pkg/ir/options.go).

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
