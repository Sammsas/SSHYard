#!/usr/bin/env bash
# Prints the download section appended to the GitHub release notes. Links
# are built from the final asset names, and a missing asset fails instead
# of publishing a dead link. Usage: release-download-notes.sh dist tag owner/repo
set -euo pipefail

if [[ $# -ne 3 ]]; then
  echo "usage: $0 <dist-dir> <tag> <owner/repo>" >&2
  exit 2
fi
dist="$1"
base="https://github.com/$3/releases/download/$2"

# link <label> <glob>: Markdown link to the single asset matching glob
link() {
  local matches=()
  local path
  for path in "$dist"/$2; do
    [[ -e "$path" ]] && matches+=("${path##*/}")
  done
  if [[ ${#matches[@]} -ne 1 ]]; then
    echo "expected one asset matching $2, found ${#matches[@]}" >&2
    return 1
  fi
  printf '[%s](%s/%s)' "$1" "$base" "${matches[0]}"
}

mac_arm_zip="$(link 'ZIP (.app)' 'sshyard-*-darwin-arm64.zip')"
mac_x64_zip="$(link 'ZIP (.app)' 'sshyard-*-darwin-amd64.zip')"
win_x64_setup="$(link Installer 'sshyard-setup-*-amd64.exe')"
win_x64_portable="$(link 'Portable ZIP' 'sshyard-*-windows-amd64-portable.zip')"
checksums="$(link SHA256SUMS 'SHA256SUMS')"

cat <<EOF

---

## Downloads

### macOS

macOS 12 or later. Extract the ZIP and drag \`SSHYard.app\` to \`Applications\`.

| Architecture | Download |
| --- | --- |
| Apple Silicon | $mac_arm_zip |
| Intel | $mac_x64_zip |

Cross-compiled without Apple notarization; if Gatekeeper blocks opening, run \`xattr -cr /Applications/SSHYard.app\` (or allow it in Privacy & Security).

### Windows

| Architecture | Download |
| --- | --- |
| x64 | $win_x64_setup · $win_x64_portable |

The installer installs for the current user into \`%LOCALAPPDATA%\\Programs\` and supports in-app updates; no administrator rights needed.

For the portable version, extract the ZIP into a writable folder and run \`SSHYard.exe\`; no installation needed. App data is stored in \`%APPDATA%\\SSHYard\`, and saved credentials remain tied to the Windows user account.

Verify downloads with $checksums.
EOF
