package modlicense

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/google/licensecheck"
)

const Version = "0.0.1"

type License string

type Dependency struct {
	Module  string
	Version string
}

func (d Dependency) String() string {
	return d.Module + " " + d.Version
}

func (d Dependency) MarshalJSON() ([]byte, error) {
	return []byte(d.String()), nil
}

func (d *Dependency) UnmarshalJSON(b []byte) error {
	fmt.Println(string(b))
	return errors.New("unimplemented")
}

type ModLicenses struct {
	Licenses map[Dependency]License `json:"licenses"`
}

func (ml ModLicenses) MarshalJSON() ([]byte, error) {
	mp := make(map[string]string, len(ml.Licenses))
	for k, v := range ml.Licenses {
		mp[k.String()] = string(v)
	}
	return json.Marshal(mp)
}

func (ml *ModLicenses) UnmarshalJSON(in []byte) error {
	mp := make(map[string]string)
	err := json.Unmarshal(in, &mp)
	if err != nil {
		return err
	}
	ml.Licenses = make(map[Dependency]License, len(mp))
	for k, v := range mp {
		splitKey := strings.Split(k, " ")
		if len(splitKey) != 2 {
			return fmt.Errorf("malformed file: line did not match 'module vX.Y.Z'")
		}
		dep := Dependency{
			Module:  splitKey[0],
			Version: splitKey[1],
		}
		ml.Licenses[dep] = License(v)
	}
	return nil
}

func ParseDependency(modline []byte) (Dependency, error) {
	commentAt := bytes.Index(modline, []byte("//"))
	if commentAt != -1 {
		modline = modline[:commentAt]
	}
	modline = bytes.TrimSpace(modline)
	splitLine := bytes.Split(modline, []byte{' '})
	if len(splitLine) != 2 {
		fmt.Println(splitLine)
		return Dependency{}, fmt.Errorf("malformed file: line did not match 'module vX.Y.Z'")
	}
	return Dependency{
		Module:  string(splitLine[0]),
		Version: string(splitLine[1]),
	}, nil
}

func FromModfile(r io.Reader) (ModLicenses, error) {
	requiredDependencies := make(map[Dependency]License)
	br := bufio.NewReader(r)
	var err error
	var inRequireBlock bool
	var ln []byte
	var lineNumber int
	for err == nil {
		lineNumber++
		ln, err = br.ReadBytes('\n')
		if err == nil || errors.Is(err, io.EOF) {
			ln = bytes.TrimSpace(ln)
			if inRequireBlock {
				if ln[0] == ')' {
					inRequireBlock = false
				} else {
					d, err := ParseDependency(ln)
					if err != nil {
						return ModLicenses{}, fmt.Errorf("parse error at line %d: %w", lineNumber, err)
					}
					requiredDependencies[d] = ""
				}
			} else {
				if bytes.HasPrefix(ln, []byte("require")) {
					if bytes.Contains(ln, []byte("(")) {
						inRequireBlock = true
					} else {
						// single dependency
						ln = bytes.TrimPrefix(ln, []byte("require"))
						d, err := ParseDependency(ln)
						if err != nil {
							return ModLicenses{}, fmt.Errorf("parse error at line %d: %w", lineNumber, err)
						}
						requiredDependencies[d] = ""
					}
				}
			}
		}
	}
	if err != io.EOF {
		return ModLicenses{}, fmt.Errorf("failed to read modfile contents: %v", err)
	}
	for dep := range requiredDependencies {
		license, err := dep.GetLicense()
		if errors.Is(err, ErrNoConfidentLicense) {
			license = "unknown"
		} else if err != nil {
			// TODO: differentiate fatal errors and warnings
			return ModLicenses{}, fmt.Errorf("failed to extract license for %v: %w", dep.Module, err)
		}
		requiredDependencies[dep] = license
	}
	return ModLicenses{
		Licenses: requiredDependencies,
	}, nil
}

func FromModFilePath(modpath string) (ModLicenses, error) {
	r, err := os.Open(modpath)
	if err != nil {
		return ModLicenses{}, fmt.Errorf("failed to open path %v: %w", modpath, err)
	}
	defer r.Close()
	return FromModfile(r)
}

func FromDir(dir string) (ModLicenses, error) {
	fullPath := filepath.Join(dir, "go.mod")
	return FromModFilePath(fullPath)
}

func FromWD() (ModLicenses, error) {
	wd, err := os.Getwd()
	if err != nil {
		return ModLicenses{}, fmt.Errorf("failed to getwd: %w", err)
	}
	return FromDir(wd)
}

func (d Dependency) GetLicense() (License, error) {
	if d.Module == "" {
		return "", fmt.Errorf("undefined module")
	}
	if d.Version == "" {
		return "", fmt.Errorf("undefined version")
	}
	modcache := os.Getenv("GOMODCACHE")
	if modcache == "" {
		gopath := os.Getenv("GOPATH")
		modcache = filepath.Join(gopath, "pkg", "mod")
	}
	splitModule := strings.Split(d.Module, "/")
	for i, component := range splitModule {
		// replace uppercase letters with '!' and lowercase
		componentRunes := make([]rune, 0, len(component))
		for _, rn := range component {
			if unicode.IsUpper(rn) {
				componentRunes = append(componentRunes, '!')
			}
			componentRunes = append(componentRunes, unicode.ToLower(rn))
		}
		if len(componentRunes) != len(component) {
			splitModule[i] = string(componentRunes)
		}
	}
	depPath := filepath.Join(append([]string{modcache}, splitModule...)...)
	depPath += "@" + d.Version
	ents, err := os.ReadDir(depPath)
	if err != nil {
		// TODO: go mod tidy the target directory / file so go actually downloads the dependencies?
		return "", fmt.Errorf("failed to read go mod cache: %w", err)
	}
	var licenseFilePath string
	for _, ent := range ents {
		switch {
		case strings.Contains(ent.Name(), "COPYING"):
			fallthrough
		case strings.Contains(ent.Name(), "LICENSE"):
			licenseFilePath = filepath.Join(depPath, ent.Name())
		}
	}
	// TODO: fall back to other methods if there is no license file:
	// a. is it embedded in the readme?
	// b. is it embedded in the source code files themselves?
	if licenseFilePath == "" {
		return "", ErrNoLicenseFile
	}
	// TODO: a dependency may have multiple licenses for different portions of itself
	licenseFile, err := os.Open(licenseFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to open license file: %w", err)
	}
	defer licenseFile.Close()
	const tooBigForALicenseFile = 100000
	licenseContent, err := ioutil.ReadAll(io.LimitReader(licenseFile, tooBigForALicenseFile))
	if err != nil {
		return "", fmt.Errorf("failed to read license file: %w", err)
	}
	coverage := licensecheck.Scan(licenseContent)
	if coverage.Percent < 90 {
		return "", ErrNoConfidentLicense
	}
	return License(coverage.Match[0].ID), nil
}

var ErrNoLicenseFile = fmt.Errorf("no license file found")

var ErrNoConfidentLicense = fmt.Errorf("no confident license found in file")
