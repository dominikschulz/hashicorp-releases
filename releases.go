package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"

	"github.com/alexflint/go-arg"
	"github.com/blang/semver"
)

var releaseURL = "https://releases.hashicorp.com/"

type Build struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Filename string `json:"filename"`
	URL      string `json:"url"`
}

type Version struct {
	Name       string  `json:"name"`
	Version    string  `json:"version"`
	Shasums    string  `json:"shasums"`
	ShasumsSig string  `json:"shasums_signature"`
	Builds     []Build `json:"builds"`
}

func (v *Version) Build() *Build {
	for _, b := range v.Builds {
		if b.OS == "linux" && b.Arch == "amd64" {
			return &b
		}
	}
	return nil
}

type Release struct {
	Name     string             `json:"name"`
	Versions map[string]Version `json:"versions"`
	versions semver.Versions
}

func (r *Release) LatestRelease(includePre bool) *Version {
	if r.versions == nil {
		vs := make(semver.Versions, len(r.Versions))
		for v := range r.Versions {
			vp, err := semver.Make(v)
			if err != nil {
				fmt.Printf("Failed to parse version %s: %s", v, err)
				continue
			}
			if len(vp.Pre) > 0 {
				//fmt.Printf("Skipping pre-release: %s", vp.String())
				continue
			}
			vs = append(vs, vp)
		}
		sort.Sort(vs)
		r.versions = vs
	}
	if len(r.versions) < 1 {
		return nil
	}
	v := r.Versions[r.versions[len(r.versions)-1].String()]
	return &v
}

var args struct {
	Product string `arg:"required"`
	Version string
	URL     bool
}

func main() {
	arg.MustParse(&args)

	r, err := fetchRelease(args.Product)
	if err != nil {
		fmt.Printf("Failed to fetch releases for %s: %s", args.Product, err)
		os.Exit(1)
	}

	v := r.LatestRelease(false)
	b := v.Build()
	if args.URL && b != nil {
		fmt.Printf("%s\n", b.URL)
	}
	if v.Version != args.Version {
		os.Exit(1)
	}
	os.Exit(0)
}

func fetchRelease(product string) (*Release, error) {
	url := releaseURL + product + "/index.json"
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch from %s: %s", url, err)
	}
	defer resp.Body.Close()

	var r Release
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}
