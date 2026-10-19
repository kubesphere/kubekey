#! /bin/sh
#
# Download the given PACKAGES and their complete dependency closure as .deb
# files for building an offline repository ISO.
#
# Dependency resolution is delegated to apt itself via `apt-get install -s`
# (simulate) with --no-install-recommends, so only the packages that are
# actually required are resolved. apt prints every package it would install as an
# `Inst <pkg> ...` line; field 2 is the bare package name, with no version
# qualifier attached. This removes the fragile text parsing of `apt-cache
# depends` (which appends version constraints such as `libjansson4 (>= 2.14)`
# that break `apt-get download`), and offloads recursion and cycle handling to
# apt's own resolver. Any download failure aborts the script with a non-zero
# exit code instead of being swallowed, so an incomplete offline repository
# fails loudly at build time instead of breaking a later cluster installation.

TMP_LIST="$(mktemp)"

# Let apt resolve the full transitive closure and emit clean package names.
# Field 2 of each `Inst` line is the bare package name (no version qualifier).
sudo apt-get install -s --no-install-recommends ${PACKAGES} 2>/dev/null \
    | awk '/^Inst/{print $2}' > "$TMP_LIST"

if [ ! -s "$TMP_LIST" ]; then
    echo "ERROR: failed to resolve dependency closure for: ${PACKAGES}" >&2
    rm -f "$TMP_LIST"
    exit 1
fi

# Download each package once, failing loudly on the first error.
while read -r pkg; do
    echo "Downloading $pkg ..."
    if ! sudo apt-get download "$pkg" 2>&1; then
        echo "ERROR: failed to download required package: $pkg" >&2
        rm -f "$TMP_LIST"
        exit 1
    fi
done < "$TMP_LIST"
rm -f "$TMP_LIST"
