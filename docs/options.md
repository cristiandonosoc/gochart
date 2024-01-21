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
| Description | The name of the namespace to generate in |gochart::<NAMESPACE>|.
Otherwise the name of gochart will be used. |

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

### parallel

| Name | Description |
| --- | --- |
| Flag | parallel |
| Kind | boolean |
| Description | Whether the root states of the gochart should be considered running in parallel within each other.

When this option is defined, all child states will run as separate parallel gochart groups and the
gochart is considered to be in all those states simultaneously.

See https://github.com/cristiandonosoc/gochart/docs/concurrent.md for more details. |

**Synopsis:**
```

```

### 

| Name | Description |
| --- | --- |
| Flag | test_bool |
| Kind | boolean |
| Description | Option added for testing purposes. |

**Synopsis:**
```

```

### 

| Name | Description |
| --- | --- |
| Flag | test_int |
| Kind | int |
| Description | Option added for testing purposes. |

**Synopsis:**
```

```

### 

| Name | Description |
| --- | --- |
| Flag | test_string |
| Kind | string |
| Description | Option added for testing purposes. |

**Synopsis:**
```

```
