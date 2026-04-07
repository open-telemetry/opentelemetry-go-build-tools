#!/usr/bin/env bash

# Copyright The OpenTelemetry Authors
# SPDX-License-Identifier: Apache-2.0

AFFECTED_FILES=$@

declare -A affected_modules

for FILE in $AFFECTED_FILES; do
    DIR=$(dirname "$FILE")

    # Walk up the directory tree and collect all directories with go.mod files.
    while [[ "$DIR" != "." ]]; do
        if [[ -f "$DIR/go.mod" ]]; then
            affected_modules["$DIR"]=1
        fi
        DIR=$(dirname "$DIR")
    done

    # Check if root (.) has a go.mod file and add it if it does.
    if [[ -f "$DIR/go.mod" ]]; then
        affected_modules["$DIR"]=1
    fi
done

echo "${!affected_modules[@]}"