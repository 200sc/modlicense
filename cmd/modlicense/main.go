package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/200sc/modlicense"
)

func main() {
	err := run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}

var (
	dir     = flag.String("directory", ".", "directory to scan")
	file    = flag.String("modfile", "go.mod", "filename to scan")
	known   = flag.String("known", "", "existing dependency version file, used to populate unknown versions")
	version = flag.Bool("version", false, "print version and exit")
)

func run() error {
	flag.Parse()
	if *version {
		fmt.Println(modlicense.Version)
		return nil
	}
	deps, err := modlicense.FromModFilePath(filepath.Join(*dir, *file))
	if err != nil {
		return err
	}
	if *known != "" {
		knownDeps := modlicense.ModLicenses{}
		knownFile, err := os.Open(*known)
		if err != nil {
			return err
		}
		defer knownFile.Close()
		dec := json.NewDecoder(knownFile)
		err = dec.Decode(&knownDeps)
		if err != nil {
			return err
		}
		for k, v := range deps.Licenses {
			if v == "unknown" && knownDeps.Licenses[k] != "unknown" {
				deps.Licenses[k] = knownDeps.Licenses[k]
			}
		}
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "\t")
	err = enc.Encode(deps)
	if err != nil {
		return err
	}
	return nil
}
