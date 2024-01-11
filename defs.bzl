def generate_gochart_cpp(name, src, file_prefix, **kwargs):
    forward_declare_out = "{file_prefix}.forward_declare.h".format(file_prefix = file_prefix)
    declare_out = "{file_prefix}.declare.h".format(file_prefix = file_prefix)
    define_out = "{file_prefix}.define.h".format(file_prefix = file_prefix)

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
