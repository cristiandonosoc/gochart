def generate_gochart_cpp(name, src, forward_declare_out, declare_out, define_out, **kwargs):
  native.genrule(
    name = name,
    srcs = [src],
    outs = [
      forward_declare_out,
      declare_out,
      define_out,
    ],
    cmd = ("$(execpath //cmd/gochart) backend cpp generate " +
              "$(location {src}) $(location {forward_declare_out}) $(location {declare_out}) $(location {define_out})".format(
            src = src,
            forward_declare_out = forward_declare_out,
            declare_out = declare_out,
            define_out = define_out,
          )),
    tools = ["//cmd/gochart"],
  )
