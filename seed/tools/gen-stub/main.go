package main

import (
	"flag"
	"log"
	"os"
	"runtime/debug"

	"github.com/spf13/pflag"
)

type Options struct {
	SrcDir     string
	BasePkg    string
	DstDirList []string
	DocName    string
	Prefix     string

	Debug bool
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("** ")

	var options Options
	pflag.StringVar(&options.BasePkg, "base", "", "base package")
	pflag.StringVar(&options.SrcDir, "src", "", "source directory")
	pflag.StringSliceVar(&options.DstDirList, "dst", nil, "destination directory")
	pflag.StringVar(&options.DocName, "doc", "", "target openapi doc (OAS3.0)")
	pflag.StringVar(&options.Prefix, "prefix", "", "prefix")
	pflag.BoolVar(&options.Debug, "debug", false, "debug")
	pflag.Parse()

	if options.BasePkg == "" {
		if info, ok := debug.ReadBuildInfo(); ok {
			options.BasePkg = info.Main.Path
		}
		if options.Debug {
			log.Printf("[DEBUG] base package is %q (from buildinfo)", options.BasePkg)
		}
	}

	if options.SrcDir == "" || options.DstDirList == nil || options.DocName == "" {
		flag.Usage()
		os.Exit(1)
	}
	if err := run(options); err != nil {
		log.Fatalf("!! %+v", err)
	}
}

func run(options Options) error {
	return nil
}
