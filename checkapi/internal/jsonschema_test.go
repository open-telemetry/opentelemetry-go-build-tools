// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"testing"

	"github.com/kaptinlin/jsonschema"
	"github.com/stretchr/testify/assert"
)

func TestCompareSchemas(t *testing.T) {
	for _, test := range []struct {
		name   string
		before *jsonschema.Schema
		after  *jsonschema.Schema
		error  string
	}{
		{
			name:   "identical, one field",
			before: jsonschema.Object(jsonschema.Prop("foo", jsonschema.String())),
			after:  jsonschema.Object(jsonschema.Prop("foo", jsonschema.String())),
		},
		{
			name:   "different name, one field",
			before: jsonschema.Object(jsonschema.Prop("foo", jsonschema.String())),
			after:  jsonschema.Object(jsonschema.Prop("bar", jsonschema.String())),
			error:  `field "foo" is missing`,
		},
		{
			name:   "different, different type",
			before: jsonschema.Object(jsonschema.Prop("foo", jsonschema.String())),
			after:  jsonschema.Object(jsonschema.Prop("foo", jsonschema.Boolean())),
			error:  `field "foo" type changed`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := CompareJSONSchema(test.before, test.after)
			if test.error != "" {
				assert.ErrorContains(t, err, test.error)
			} else {
				assert.NoError(t, err)
			}

		})
	}
}
