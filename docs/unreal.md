# Unreal Gochart Integration

Gochart has understanding of an Unreal project as is able to generate all the gochart present in a
project, instead of having to write scripts to generate them manually. All the commands for this are
in the `unreal` section of the cli:

```
$ gochart unreal
```

## Unreal introspection

Gochart actually is able to analyze an unreal project and understand modules within them, permitting
it to generate module specific C++ code. Gochart leverages Go parallelism to be able to index big
Unreal projects quickly (indexing the complete Unreal Editor source takes under a second).

## Generation Strategies

Gochart has several strategies about where to output the generated files. The decision of where they
should be generated depends on how you want to handle them.

You need to specify the selected strategy by using the `--unreal_strategy` flag in the unreal
commands:
```
gochart unreal list --unreal_strategy=Inline [args...]
```

> [!IMPORTANT]
> Some of these strategies permit to add the generated files into the version control systems. While
> there are benefits to this (ie. No need to run the tool every time), this can create headaches
> with VCS systems that sets the files as read-only, like Perforce.
>
> In those cases, most likely you would like a rule to match the generated files as writeable and
> reconcile.

### Strategy: Inlined

| --- | --- |
| Flag | `--unreal_strategy=Inline` |

This is the simplest strategy, the files will be generated next to the .gochart file.
This is similar to what `gochart backend cpp generate` does.

TODO: Add example

*Pros*
- Simple, specially for includes.
- Permits to add the files to version control.

*Cons*
- Can be quite noisy, as you generate quite a bit of generated code along side the normal source.

#### Example Listing

```
$ gochart unreal list --unreal_strategy=Inline test/unreal/GochartTest

Listing gocharts at Unreal project at D:\projects\gochart\test\unreal\GochartTest\
Found 2 gocharts:
-----------------------------------------------------------------------
Counter [*** WILL REGENERATE ***]
- PATH: Source\GochartTest\AnotherDir\Counter.gochart
- EXPECTED OUTPUTS:
  > Source\GochartTest\AnotherDir\Counter.gochart.forward_declare.h
  > Source\GochartTest\AnotherDir\Counter.gochart.declare.h
  > Source\GochartTest\AnotherDir\Counter.gochart.define.h
-----------------------------------------------------------------------
GochartTestActor [*** WILL REGENERATE ***]
- PATH: Source\GochartTest\GochartTestActor.gochart
- EXPECTED OUTPUTS:
  > Source\GochartTest\GochartTestActor.gochart.forward_declare.h
  > Source\GochartTest\GochartTestActor.gochart.declare.h
  > Source\GochartTest\GochartTestActor.gochart.define.h
```

### Strategy: Module Generated Directory

| --- | --- |
| Flag | `--unreal_strategy=ModuleGeneratedDir` |

This approach will create a `Generated` folder at the root of each module and insert the generated
files there, keeping the same internal path of the .gochart file within the module.

TODO: Add Examples

*Pros*
- Permits to add the files into version control.
- Keeps the generated files in a separate tree-directory structure, which is cleaner.

*Cons*
- Requires minor configuration to the .Build.cs files so that the include paths can be found.

#### Example Listing

```
$ gochart unreal list --unreal_strategy=ModuleGeneratedDir test/unreal/GochartTest

Listing gocharts at Unreal project at D:\projects\gochart\test\unreal\GochartTest\
Found 2 gocharts:
-----------------------------------------------------------------------
Counter [*** WILL REGENERATE ***]
- PATH: Source\GochartTest\AnotherDir\Counter.gochart
- EXPECTED OUTPUTS:
  > Source\GochartTest\Generated\AnotherDir\Counter.gochart.forward_declare.h
  > Source\GochartTest\Generated\AnotherDir\Counter.gochart.declare.h
  > Source\GochartTest\Generated\AnotherDir\Counter.gochart.define.h
-----------------------------------------------------------------------
GochartTestActor [*** WILL REGENERATE ***]
- PATH: Source\GochartTest\GochartTestActor.gochart
- EXPECTED OUTPUTS:
  > Source\GochartTest\Generated\GochartTestActor.gochart.forward_declare.h
  > Source\GochartTest\Generated\GochartTestActor.gochart.declare.h
  > Source\GochartTest\Generated\GochartTestActor.gochart.define.h
```

### Strategy: Intermediate UHT

| --- | --- |
| Flag | `--unreal_strategy=UHT` |

> [!NOTE]
> While the approach of this strategy is solid, using the UHT directory for this is a *hack*, but
> have not found a way to get UBT to include directories within the Intermediate. As soon as we find
> one, this strategy will be deprecated.

This strategy generates the header files into the same directory where UHT would generate its own
generated source files within the Intermediate folder. We piggy-back on this directory because it is
meant for generated files and UBT already correctly sets up the C++ includes for us.

This strategy generates the module internal path into the UHT folder. Meaning that if the gochart
file was in `<MODULE_DIR>/Some/Internal/Folder/MyGochart.gochart`, the files will be generated in
`<INTERMEDIATE_PATH>/Some/Internal/Folder`. This has consequences of the includes you have to add:

```
#include "Some/Internal/Folder/MyGochart.gochart.forward_declare.h
```

*Pros*
- Clean, as all the files will be generated outside the Source directory, in the Intermediate dir.

*Cons*
- Not meant for versioning files, would need to run the tool every time.
    - See [how to hook gochart to the build](#hooking_gochart_to_the_build)

#### Example Listing

> [!IMPORTANT]
> Note in the case of `Counter.gochart`, it is under the `AnotherDir` directory, which means that it
> was placed under that directory within the UHT folder. That means that you would need to prepend
> `AnotherDir` to the include path.

```
$ gochart unreal list --unreal_strategy=UHT test/unreal/GochartTest

Listing gocharts at Unreal project at D:\projects\gochart\test\unreal\GochartTest\
Found 2 gocharts:
-----------------------------------------------------------------------
Counter [*** WILL REGENERATE ***]
- PATH: Source\GochartTest\AnotherDir\Counter.gochart
- EXPECTED OUTPUTS:
  > Intermediate\Build\Win64\UnrealEditor\Inc\GochartTest\UHT\AnotherDir\Counter.gochart.forward_declare.h
  > Intermediate\Build\Win64\UnrealEditor\Inc\GochartTest\UHT\AnotherDir\Counter.gochart.declare.h
  > Intermediate\Build\Win64\UnrealEditor\Inc\GochartTest\UHT\AnotherDir\Counter.gochart.define.h
-----------------------------------------------------------------------
GochartTestActor [*** WILL REGENERATE ***]
- PATH: Source\GochartTest\GochartTestActor.gochart
- EXPECTED OUTPUTS:
  > Intermediate\Build\Win64\UnrealEditor\Inc\GochartTest\UHT\GochartTestActor.gochart.forward_declare.h
  > Intermediate\Build\Win64\UnrealEditor\Inc\GochartTest\UHT\GochartTestActor.gochart.declare.h
  > Intermediate\Build\Win64\UnrealEditor\Inc\GochartTest\UHT\GochartTestActor.gochart.define.
```

## Listing Gocharts

You can list all sources within a project using the `gochart unreal list` command.
It will find all the .gochart files within a project, if it would re-generate the files (or the
current ones are up to date) and the expected paths of these outputs.

See examples of each strategy of how this looks.

```
$ gochart unreal list --unreal_strategy=UHT test/unreal/GochartTest

Listing gocharts at Unreal project at D:\projects\gochart\test\unreal\GochartTest
Found 1 gocharts:
-----------------------------------------------------------------------
GochartTestActor *WILL REGENERATE*
- PATH: Source\GochartTest\GochartTestActor.gochart
- EXPECTED OUTPUTS:
  > Intermediate\Build\Win64\UnrealEditor\Inc\GochartTest\UHT\GochartTestActor.forward_declare.h
  > Intermediate\Build\Win64\UnrealEditor\Inc\GochartTest\UHT\GochartTestActor.declare.h
  > Intermediate\Build\Win64\UnrealEditor\Inc\GochartTest\UHT\GochartTestActor.define.h
```

## Unreal Generate

Generate works similar to `List`, but will actually go ahead and generate the files according to
the provided strategy:

```
$ gochart unreal generate --unreal_strategy=UHT test/unreal/GochartTest

Found 1 gocharts:
- D:\projects\gochart\test\unreal\GochartTest\Source\GochartTest\GochartTestActor.gochart
Generation result:
- GochartTestActor:
  > Generated D:\projects\gochart\test\unreal\GochartTest\Intermediate\Build\Win64\UnrealEditor\Inc\GochartTest\UHT\GochartTestActor.gochart.forward_declare.h
  > Generated D:\projects\gochart\test\unreal\GochartTest\Intermediate\Build\Win64\UnrealEditor\Inc\GochartTest\UHT\GochartTestActor.gochart.declare.h
  > Generated D:\projects\gochart\test\unreal\GochartTest\Intermediate\Build\Win64\UnrealEditor\Inc\GochartTest\UHT\GochartTestActor.gochart.define.h
```

Gochart should be smart enough to detect when it does not need to regenerate the files when
appropriate and will inform in that case:

```
$ gochart unreal generate --unreal_strategy=UHT test/unreal/GochartTest

Found 1 gocharts:
- D:\projects\gochart\test\unreal\GochartTest\Source\GochartTest\GochartTestActor.gochart
Generation result:
- GochartTestActor UP TO DATE
```

## Hooking Gochart to the Build

It is possible to hook Gochart so that it runs automatically when compiling the unreal project,
regardless from where we're doing this (command line, Visual Studio, Rider, etc.).

This is done by using the [PreBuildSteps](https://docs.unrealengine.com/4.27/en-US/API/Runtime/Projects/FProjectDescriptor/PreBuildSteps/), which are hooks we can instruct UBT to run *before* it starts
building. With this we can ensure all the gochart files are up to date generated before building.

As stated in the previous section, Gochart is careful to only generate the files when there is a
difference, in order to minimize invalidating the build cache. Gochart itself is very quick to run
(we aim for under-second run times), meaning that if you don't change your gochart files, you should
not pay for the cost of running Gochart.

See the [test example](../test/unreal/GochartTest/GochartTest.uproject) for an example of an
integration.
