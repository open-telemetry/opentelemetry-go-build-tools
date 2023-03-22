// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// gotmpl is a tool to generate files from [text/template] and JSON data.
package main

import (
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
)

var (
	body = flag.StringP("body", "b", "", "Template body filepath.")
	data = flag.StringP("data", "d", "", "Data in JSON format.")
	out  = flag.StringP("out", "o", "", "Output filepath.")
)

func main() {
	flag.Parse()

	if err := gotmpl(*body, *data, *out); err != nil {
		fmt.Println(err)
		flag.PrintDefaults()
		os.Exit(1)
	}
}
