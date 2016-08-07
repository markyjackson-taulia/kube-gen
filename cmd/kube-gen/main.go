// kube-gen is a
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	kubegen "github.com/kylemcc/kube-gen"
)

type stringSlice []string

var (
	// flags
	host        string
	types       stringSlice
	watch       bool
	notifyCmd   string
	overwrite   bool
	wait        string
	showVersion bool

	// build info
	version   string
	buildTime string
	revision  string
)

// implement flag.Value interface
func (s *stringSlice) String() string {
	return strings.Join(*s, ", ")
}

func (s *stringSlice) Set(v string) error {
	*s = append(*s, v)
	return nil
}

func usage() {
	fmt.Printf(`Usage: kube-gen [options] <template> [<output>]

Render templates using Kubernetes metadata and events

Options:
`)
	flag.PrintDefaults()

	fmt.Printf(`
Arguments:
  template: path or URL of the template file to render, or - to read from STDIN
  output: (Optional) path to write the rendered content. If not specified,
          rendered content is printed to STDOUT. By default, this file will
          be overwritten if it exists. Use -overwrite=false to return an
          error instead

Environment Variables:
  KUBERNETES_SERVICE_HOST
  KUBERNETES_SERVICE_PORT

Examples:
`)
}

func parseFlags() {
	flag.StringVar(&host, "host", "http://localhost:8001", "")
	flag.Var(&types, "type", "types of resources to pull [pods, services, endpoints] - May be specified multiple times. "+
		"If not specified, all types will be returned")
	flag.BoolVar(&showVersion, "version", false, "display version information")
	flag.BoolVar(&watch, "watch", false, "watch for new events")
	flag.StringVar(&notifyCmd, "notify", "", "command to run after template generation in complete")
	flag.BoolVar(&overwrite, "overwrite", true, "overwrite the output file if it exists")
	flag.StringVar(&wait, "wait", "", "<minimum>[:<maximum>] - the minimum and optional maximum time to wait after an event fires."+
		"E.g.: 500ms:5s")

	flag.Usage = usage
	flag.Parse()
}

func printVersion() {
	fmt.Printf(`kube-gen v%s
built at: %s
revision: %s
`, version, buildTime, revision)
}

func parseWait() (min time.Duration, max time.Duration, err error) {
	if len(wait) == 0 {
		return 0, 0, nil
	}
	parts := strings.Split(wait, ":")
	if min, err = time.ParseDuration(parts[0]); err != nil {
		return
	}
	if len(parts) > 1 {
		max, err = time.ParseDuration(parts[1])
		if err == nil && max < min {
			err = errors.New("max must be greater than or equal to min")
		}
	}
	return
}

func main() {
	parseFlags()

	if showVersion {
		printVersion()
		return
	}

	if narg := flag.NArg(); narg < 1 || narg > 2 {
		flag.Usage()
		os.Exit(1)
	}

	minWait, maxWait, err := parseWait()
	if err != nil {
		log.Fatalf("invalid wait value: %v", err)
	}

	conf := kubegen.Config{
		Host:          host,
		Template:      flag.Arg(0),
		Output:        flag.Arg(1),
		Overwrite:     overwrite,
		Watch:         watch,
		NotifyCmd:     notifyCmd,
		ResourceTypes: types,
		MinWait:       minWait,
		MaxWait:       maxWait,
	}

	gen, err := kubegen.NewGenerator(conf)
	if err != nil {
		log.Fatalf("error initializing generator: %v", err)
	}

	if err := gen.Generate(); err != nil {
		log.Fatalf("error generating output: %v", err)
	}
}
