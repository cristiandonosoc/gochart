def generate_gochart_cpp(name, src, header_out, cpp_out, **kwargs):
  native.genrule(
    name = "%s_gochart_genrule" % src,
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
  native.cc_library(
    name = name,
    hdrs = [header_out],
    srcs = [cpp_out],
		include_prefix = ".", # Add -I of dependencies.
    **kwargs
  )


