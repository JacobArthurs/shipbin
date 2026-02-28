import os
import subprocess
import sys
import platform


def main():
    system = platform.system().lower()

    if system == "windows":
        binary_name = "__BIN_NAME__.exe"
    else:
        binary_name = "__BIN_NAME__"

    binary_path = os.path.join(os.path.dirname(__file__), "bin", binary_name)

    if not os.path.isfile(binary_path):
        print(
            f"__BIN_NAME__: binary not found at {binary_path}\n"
            f"try reinstalling: pip install __BIN_NAME__",
            file=sys.stderr,
        )
        sys.exit(1)

    if system == "windows":
        proc = subprocess.run([binary_path] + sys.argv[1:])
        sys.exit(proc.returncode)
    else:
        os.execv(binary_path, [binary_path] + sys.argv[1:])
