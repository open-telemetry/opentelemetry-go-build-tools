// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package environment

import (
	"context"
	"testing"
	"path"

	"go.opentelemetry.io/build-tools/grater/internal/module"
	"go.opentelemetry.io/build-tools/grater/internal/dockercontainer"
	"go.opentelemetry.io/build-tools/grater/internal/container"
	"github.com/stretchr/testify/assert"
)
