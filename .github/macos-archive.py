#!/usr/bin/env python3
"""ZIP MyGo's cross-compiled macOS .app bundle with Unix permissions preserved."""

import json
import os
from pathlib import Path
import stat
import sys
import zipfile


def package(build_dir, version, arch):
    app_dir = build_dir / "SSHYard.app"
    if not app_dir.is_dir():
        raise FileNotFoundError(f"Missing macOS bundle: {app_dir}")
    output = build_dir / f"sshyard-{version}-darwin-{arch}.zip"
    macos_dir = app_dir / "Contents" / "MacOS"
    with zipfile.ZipFile(output, "w", zipfile.ZIP_DEFLATED) as archive:
        for path in sorted(app_dir.rglob("*")):
            rel = path.relative_to(build_dir)
            arcname = rel.as_posix()
            st = os.lstat(path)
            if stat.S_ISLNK(st.st_mode):
                info = zipfile.ZipInfo(arcname)
                info.create_system = 3
                info.external_attr = (stat.S_IFLNK | 0o777) << 16
                archive.writestr(info, os.readlink(path))
                continue
            if path.is_dir():
                info = zipfile.ZipInfo(arcname + "/")
                info.create_system = 3
                info.external_attr = ((stat.S_IFDIR | (st.st_mode & 0o777) or 0o755) << 16) | 0x10
                archive.writestr(info, b"")
                continue
            mode = st.st_mode & 0o777
            if path.parent == macos_dir:
                mode |= 0o755
            info = zipfile.ZipInfo.from_file(path, arcname)
            info.create_system = 3
            info.external_attr = (stat.S_IFREG | mode) << 16
            archive.writestr(info, path.read_bytes(), compress_type=zipfile.ZIP_DEFLATED)
    return output


if __name__ == "__main__":
    if len(sys.argv) != 3 or sys.argv[2] not in ("amd64", "arm64"):
        sys.exit("usage: macos-archive.py <darwin-build-dir> <amd64|arm64>")
    config = json.loads(Path("mygo.json").read_text(encoding="utf-8"))
    print(package(Path(sys.argv[1]), config["version"], sys.argv[2]))
