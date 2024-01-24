import platform
import os
import os.path as p
import subprocess

DIR_OF_THIS_SCRIPT = p.abspath( p.dirname( __file__ ) )

# These are the compilation flags that will be used in case there's no
# compilation database set (by default, one is not set).
# CHANGE THIS LIST OF FLAGS. YES, THIS IS THE DROID YOU HAVE BEEN LOOKING FOR.
flags = [
    '-Wall',
    '-Wextra',
    '-Werror',
    '-Wno-long-long',
    '-Wno-variadic-macros',
    '-fexceptions',
    '-DNDEBUG',
    '-xc++',
    '-std=c++20',
    '-Wno-unused-parameter',
    '-Wno-unused-variable',
    "-I.",
    "-Ibazel-out/x64_windows-fastbuild/bin",
    "-Ibazel-gochart/external/com_google_googletest",
    "-Ibazel-gochart/external/bazel_tools",
    "-Ibazel-gochart/external/com_google_googletest/googlemock",
    "-Ibazel-gochart/external/com_google_googletest/googlemock/include",
    "-Ibazel-gochart/external/com_google_googletest/googletest",
    "-Ibazel-gochart/external/com_google_googletest/googletest/include",
]

def AppendVirtualIncludesForDir(directory, out):
  for child in os.listdir(directory):
    path = os.path.join(directory, child)
    if not os.path.isdir(path):
      continue
    out.append(path)


def SearchForVirtualIncludes(directory, out):
  child_dirs = []
  for child in os.listdir(directory):
    path = os.path.join(directory, child)
    if not os.path.isdir(path):
      continue
    child_dirs.append(path)

  for child_dir in child_dirs:
    # If it is a virtual include, we add it to the matched.
    if '_virtual_includes' in child_dir:
      AppendVirtualIncludesForDir(child_dir, out)
      continue
    # We look into this child directory.
    SearchForVirtualIncludes(child_dir, out)

def AppendBazelVirtualIncludes(base_directory, current_flags):
  matches = []

  # We search in the bazel-out directory.
  base_dir = os.path.join(base_directory, "bazel-out", "x64_windows-fastbuild", "bin")
  SearchForVirtualIncludes(base_dir, matches)

  # Append the found matches to the flags, relative to the current path.
  for match in matches:
    rel = os.path.relpath(match, base_directory)
    current_flags.append("-I%s" % rel)

AppendBazelVirtualIncludes(os.path.dirname(DIR_OF_THIS_SCRIPT), flags)

for flag in flags:
    print("{}".format(flag.replace("\\", "/")))
