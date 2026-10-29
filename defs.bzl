def generate_gochart_cpp(name, src, file_prefix, **kwargs):
    forward_declare_out = "{file_prefix}.gochart.forward_declare.h".format(file_prefix = file_prefix)
    declare_out = "{file_prefix}.gochart.declare.h".format(file_prefix = file_prefix)
    define_out = "{file_prefix}.gochart.define.h".format(file_prefix = file_prefix)

    native.genrule(
        name = name,
        srcs = [src],
        outs = [
            forward_declare_out,
            declare_out,
            define_out,
        ],
        cmd = ("$(execpath //cmd/gochart) backend cpp generate " +
               "$(location {src}) $(@D) {file_prefix}".format(
                   src = src,
                   file_prefix = file_prefix,
               )),
        tools = ["//cmd/gochart"],
    )

def _generate_gochart_rust_impl(ctx):
    out = ctx.actions.declare_file("{}.gochart.rs".format(ctx.attr.file_prefix))

    # The tool runs directly (no shell), so this also works with remote execution on another OS.
    ctx.actions.run(
        executable = ctx.executable._gochart,
        arguments = [
            "backend",
            "rust",
            "generate",
            ctx.file.src.path,
            out.dirname,
            ctx.attr.file_prefix,
        ],
        inputs = [ctx.file.src],
        outputs = [out],
        mnemonic = "GochartRust",
        progress_message = "Generating Rust gochart %{output}",
    )

    return [DefaultInfo(files = depset([out]))]

generate_gochart_rust = rule(
    implementation = _generate_gochart_rust_impl,
    doc = """Generates <file_prefix>.gochart.rs from a gochart written with language = "rust".

The output is a plain source file. Add it to the srcs of a Rust target (next to the file that
declares the module) and include it with:

    #[path = "<file_prefix>.gochart.rs"]
    mod my_chart;
""",
    attrs = {
        "src": attr.label(
            doc = "The .gochart file.",
            allow_single_file = [".gochart"],
            mandatory = True,
        ),
        "file_prefix": attr.string(
            doc = "The name of the generated file, without the .gochart.rs extension.",
            mandatory = True,
        ),
        "_gochart": attr.label(
            default = Label("//cmd/gochart"),
            executable = True,
            cfg = "exec",
        ),
    },
)
