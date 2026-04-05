#!/usr/bin/env bash

# Copyright The OpenTelemetry Authors
# SPDX-License-Identifier: Apache-2.0

FILES=$@

declare -A affected_modules

for FILE in $FILES; do
    DIR=$(dirname "$FILE")
    
    # Climb up to find go.mod
    while [[ "$DIR" != "." && "$DIR" != "/" ]]; do
        if [[ -f "$DIR/go.mod" ]]; then
            affected_modules["$DIR"]=1
        fi
        DIR=$(dirname "$DIR")
    done
    
    # Check root
    if [[ -f "go.mod" ]]; then
        affected_modules["."]=1
    fi
done

echo "${!affected_modules[@]}"