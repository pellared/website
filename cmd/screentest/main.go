// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Command screentest runs the screentest check for a set of scripts.

	Usage: screentest [flags] [glob]

The flags are:

	-test
	  URL to test against
	-want
	  URL for expected results
	-headers
	  HTTP headers to send
	-o
	  URL for output
	-u
	  update cached screenshots
	-v
	  variables provided to script templates as comma separated KEY:VALUE pairs
	-c
	  number of testcases to run concurrently
	-d
	  chrome debugger url
	-run
	  run only tests matching regexp
*/
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

var (
	testURL     = flag.String("test", "", "URL or file path to test")
	wantURL     = flag.String("want", "", "URL or file path with expected results")
	update      = flag.Bool("u", false, "update cached screenshots")
	vars        = flag.String("v", "", "variables provided to script templates as comma separated KEY:VALUE pairs")
	concurrency = flag.Int("c", (runtime.NumCPU()+1)/2, "number of testcases to run concurrently")
	debuggerURL = flag.String("d", "", "chrome debugger url")
	run         = flag.String("run", "", "regexp to match test")
	outputURL   = flag.String("o", "", "path for output: file path or URL with 'file' or 'gs' scheme")
	headers     = flag.String("headers", "", "HTTP headers: comma-separated list of name:value")
)

func main() {
	flag.Usage = func() {
		fmt.Printf("Usage: screentest [flags] [glob]\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	args := flag.Args()
	// Require testdata glob when invoked as an installed command.
	if len(args) != 1 && os.Args[0] == "screentest" {
		flag.Usage()
		os.Exit(1)
	}
	glob := filepath.Join("cmd", "screentest", "testdata", "*")
	if len(args) == 1 {
		glob = args[0]
	}
	parsedVars := make(map[string]string)
	if *vars != "" {
		for _, pair := range strings.Split(*vars, ",") {
			parts := strings.SplitN(pair, ":", 2)
			if len(parts) != 2 {
				log.Fatal(fmt.Errorf("invalid key value pair, %q", pair))
			}
			parsedVars[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	var splitHeaders []string
	if len(*headers) > 0 {
		splitHeaders = strings.Split(*headers, ",")
	}
	opts := CheckOptions{
		TestURL:        *testURL,
		WantURL:        *wantURL,
		Update:         *update,
		MaxConcurrency: *concurrency,
		Vars:           parsedVars,
		DebuggerURL:    *debuggerURL,
		OutputURL:      *outputURL,
		Headers:        splitHeaders,
	}
	if *run != "" {
		re, err := regexp.Compile(*run)
		if err != nil {
			log.Fatal(err)
		}
		opts.Filter = re.MatchString
	}
	if err := CheckHandler(glob, opts); err != nil {
		log.Fatal(err)
	}
}
