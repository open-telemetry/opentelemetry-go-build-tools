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

/*
gocp copies Go source code.
The copied code is marked as generated.

Usage of gocp:

	-d, --dest string   The destination filepath.
	-p, --pkg string    The destination package name (can contain comments).
	-s, --src string    The source filepath.
*/
package main

import (
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
)

var (
	dest = flag.StringP("dest", "d", "", "The destination filepath.")
	pkg  = flag.StringP("pkg", "p", "", "The destination package name (can contain comments).")
	src  = flag.StringP("src", "s", "", "The source filepath.")
)

func main() {
	flag.Parse()

	if err := copy(*dest, *pkg, *src); err != nil {
		fmt.Println(err)
		flag.PrintDefaults()
		os.Exit(1)
	}
}
