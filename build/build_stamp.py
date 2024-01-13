# Script that reads environement variables that contains GOCHART in them and prints them as
# KEY VALUE pairs, so that Bazel can embed them.

import os

for key, value in os.environ.items():
    if "GOCHART" in key:
        print(f"{key} {value}")
