#!/usr/bin/env bash

# Copyright The OpenTelemetry Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail

# This script automates the installation of the multimod tool.
# It uses 'go install' to pull the latest version from the repository.

# Default values
INSTALL_PATH="${GOBIN:-$(go env GOPATH)/bin}"
PACKAGE="go.opentelemetry.io/build-tools/multimod"
VERSION="latest"

usage() {
    echo "Usage: $0 [-v VERSION] [-p PATH]"
    echo "  -v: Version to install (default: latest)"
    echo "  -p: Path to install the binary to (default: $INSTALL_PATH)"
    exit 1
}

while getopts "v:p:" opt; do
    case "$opt" in
        v) VERSION="$OPTARG" ;;
        p) INSTALL_PATH="$OPTARG" ;;
        *) usage ;;
    esac
done

echo "Installing ${PACKAGE}@${VERSION} to ${INSTALL_PATH}..."

# Ensure the install path exists
mkdir -p "${INSTALL_PATH}"

# Install the tool
GO111MODULE=on go install "${PACKAGE}@${VERSION}"

echo "Successfully installed multimod to ${INSTALL_PATH}/multimod"
echo "Make sure ${INSTALL_PATH} is in your PATH."
