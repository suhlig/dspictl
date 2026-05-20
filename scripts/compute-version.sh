#!/bin/bash
set -euo pipefail

desc=$(git describe --tags --match 'v*' 2>/dev/null || true)

if [ -z "$desc" ]; then
	printf '%s\n' "unknown"

	exit 0
fi

# Remove -dirty suffix if present (should not happen in CI)
desc=${desc%-dirty}

# Exact tag match: v1.2.3
if [[ "$desc" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
	printf '%s\n' "$desc"

	exit 0
fi

# Commits after tag: v1.2.3-5-gabc1234
if [[ "$desc" =~ ^(v[0-9]+\.[0-9]+\.[0-9]+)-([0-9]+)-g([0-9a-f]+)$ ]]; then
	printf '%s-dev.%s+%s\n' "${BASH_REMATCH[1]}" "${BASH_REMATCH[2]}" "${BASH_REMATCH[3]}"

	exit 0
fi

printf '%s\n' "unknown"
