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

## Listing Gocharts

You can list all sources within a project using the `gochart unreal list` command.
It will find all the .gochart files within a project, if it would re-generate the files (or the
current ones are up to date) and the expected paths of these outputs.

```
$ gochart unreal list test/unreal/GochartTest

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

## Generation Strategies

Gochart has several strategies about where to output the generated files. The decision of where they
should be generated depends on how you want to handle them.

> [!IMPORTANT]
> Some of these strategies permit to add the generated files into the version control systems. While
> there are benefits to this (ie. No need to run the tool every time), this can create headaches
> with VCS systems that sets the files as read-only, like Perforce.
>
> In those cases, most likely you would like a rule to match the generated files as writeable and
> reconcile.

### Strategy: Inlined

This is the simplest strategy, the files will be generated next to the .gochart file.
This is similar to what `gochart backend cpp generate` does.

TODO: Add example

*Pros*
- Simple, specially for includes.
- Permits to add the files to version control.

*Cons*
- Can be quite noisy, as you generate quite a bit of generated code along side the normal source.

### Strategy: Module Generated Directory

This approach will create a `Generated` folder at the root of each module and insert the generated
files there, keeping the same internal path of the .gochart file within the module.

TODO: Add Examples

*Pros*
- Permits to add the files into version control.
- Keeps the generated files in a separate tree-directory structure, which is cleaner.

*Cons*
- Requires minor configuration to the .Build.cs files so that the include paths can be found.

### Strategy: Intermediate UHT

> [!NOTE]
> While the approach of this strategy is solid, using the UHT directory for this is a *hack*, but
> have not found a way to get UBT to include directories within the Intermediate. As soon as we find
> one, this strategy will be deprecated.

This strategy generates the header files into the same directory where UHT would generate its own
generated source files within the Intermediate folder. We piggy-back on this directory because it is
meant for generated files and UBT already correctly sets up the C++ includes for us.

*Pros*
- Clean, as all the files will be generated outside the Source directory, in the Intermediate dir.

*Cons*
- Not meant for versioning files, would need to run the tool every time.
    - See [how to hook gochart to the build](#hooking_gochart_to_the_build)

## Unreal Generate

TODO: Write

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
