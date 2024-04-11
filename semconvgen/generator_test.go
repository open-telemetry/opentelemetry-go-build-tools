package main

import (
	"os"
	"reflect"
	"testing"
)

func TestCapitalizations(t *testing.T) {
	tests := []struct {
		name            string
		capitalizations string
		expected        []string
	}{
		{
			name:            "No additional capitalizations",
			capitalizations: "",
			expected:        staticCapitalizations,
		},
		{
			name:            "Some additional capitalizations",
			capitalizations: "ASPNETCore\nJVM",
			expected:        append(staticCapitalizations, "ASPNETCore", "JVM"),
		},
		{
			name:            "Wrong separator for capitalizations",
			capitalizations: "ASPNETCore,JVM",
			expected:        append(staticCapitalizations, "ASPNETCore,JVM"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpfile, err := os.CreateTemp("", "test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())

			if _, err = tmpfile.Write([]byte(tt.capitalizations)); err != nil {
				t.Fatal(err)
			}
			if err = tmpfile.Close(); err != nil {
				t.Fatal(err)
			}

			customCapitalizations, err := capitalizations(tmpfile.Name())
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(customCapitalizations, tt.expected) {
				t.Errorf("customCapitalizations() = %v, want %v", customCapitalizations, tt.expected)
			}
		})
	}
}
