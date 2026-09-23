#!/bin/sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
output_dir=${1:-internal/docs/static}
case "$output_dir" in
	/*) ;;
	*) output_dir="$repo_root/$output_dir" ;;
esac
# busybox mktemp (Alpine, used by the Docker build) requires the template to
# end in XXXXXX, so make a temp directory and name the archive inside it.
archive_dir=$(mktemp -d "${TMPDIR:-/tmp}/nexus-aigateway-docs.XXXXXX")
archive="$archive_dir/docs.zip"
staging_dir="${output_dir}.staging.$$"
trap 'rm -rf "$archive_dir"; rm -rf "$staging_dir"' EXIT

cd "$repo_root/docs"
npx --yes -p node@22 -p mint@latest mint export \
	--disable-openapi \
	--output "$archive"

rm -rf "$staging_dir"
mkdir -p "$staging_dir"
unzip -q "$archive" -d "$staging_dir"

# Mintlify exports assume they are hosted at /. The gateway mounts them at
# /docs, so rewrite internal HTML links and asset references during the build.
find "$staging_dir" -type f -name '*.html' -exec sh -c '
	for file do
		tmp="$file.tmp"
		sed \
			-e '\''s#href="/#href="/docs/#g'\'' \
			-e '\''s#src="/#src="/docs/#g'\'' \
			-e '\''s#content="/#content="/docs/#g'\'' \
			-e '\''s#action="/#action="/docs/#g'\'' \
			"$file" > "$tmp"
		mv "$tmp" "$file"
	done
' sh {} +

rm -rf "$output_dir"
mv "$staging_dir" "$output_dir"
