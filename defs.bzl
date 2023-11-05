def generate_gochart_cpp(name, src, header_out, cpp_out, **kwargs):
  native.genrule(
    name = name,
    srcs = [src],
    outs = [
      header_out,
      cpp_out,
    ],
    cmd = ("$(execpath //cmd/gochart) backend cpp generate " +
              "$(location {src}) $(location {header_out}) $(location {cpp_out})".format(
            src = src,
            header_out = header_out,
            cpp_out = cpp_out,
          )),
    tools = ["//cmd/gochart"],
  )
