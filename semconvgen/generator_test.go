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
		want            []string
	}{
		{
			name:            "No additional capitalizations",
			capitalizations: "",
			want:            staticCapitalizations,
		},
		{
			name:            "Some additional capitalizations",
			capitalizations: "ASPNETCore\nJVM",
			want:            append(staticCapitalizations, "ASPNETCore", "JVM"),
		},
		{
			name:            "Wrong separator for capitalizations",
			capitalizations: "ASPNETCore,JVM",
			want:            append(staticCapitalizations, "ASPNETCore,JVM"),
		},
		{
			name: "Copius amounts of whitespace",
			capitalizations: `

			 ASPNETCore

			    JVM


			`,
			want: append(staticCapitalizations, "ASPNETCore", "JVM"),
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

			if !reflect.DeepEqual(customCapitalizations, tt.want) {
				t.Errorf("capitalizations() = %v, want %v", customCapitalizations, tt.want)
			}
		})
	}
}
