// SPDX-License-Identifier:  BSD-3-Clause

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"
	"sync"
)

// Architecture note
// To simplify implementation, reduce maintenance cost and ease of life,
// This program just invoke existing commands like curl, ping, host, then parse the result...

// Steps
// 1. Parse mirror-list.json -> load into `mirrorDB` -> parse into a convenient `mirrorList`
// 2. Iterate `mirrorList` to get all urls, and trigger goroutines for each tests.
// 3. Aggregate

// Add new tests result:
// 0. Create a file as test_[name].go
// 1. Add a new field to the end of `type Mirror struct`
// 2. make a function to obtain the field
// 3. invoke the function during step [2.]

// Test function parameters:
// func(Parameters,     =>
//      wg,             => global sync.WaitGroup
//      mirrorList,     => list to hold results
//      index           => pass an index to let the function know its position at mirrorList
//     ) void {}        => No return, since we store them directly into mirrorList

func main() {

	// [0.] CLI parse & variable init
	//******************************************************************************************************************
	var mirrorListPath,
		distro string

	flag.StringVar(&distro, "distro", "", "Manually set distro name ")
	flag.StringVar(&mirrorListPath, "mirrolist", "./mirror-list.json", "Set path to mirror-list.json")
	flag.Parse()

	if distro == "" {
		distro = getSystemName()
	}

	// [1.]
	//******************************************************************************************************************
	var mirrorDB MirrorDB

	f, err := ioutil.ReadFile(mirrorListPath)
	check(err, "Cannot read ./mirror-list.json")
	err = json.Unmarshal(f, &mirrorDB)
	check(err, "mirror-list.json might be invalid")

	mirrorList := getMyMirrorList(mirrorDB, distro)

	// [2.]
	//******************************************************************************************************************
	wg := new(sync.WaitGroup) // Global wait groups for every tests

	for i, m := range mirrorList {
		wg.Add(1)
		go getAverageDownloadSpeed(m.Url+m.TestFile, wg, mirrorList, i)
		wg.Add(1)
		go getAveragePing(m.BaseUrl, wg, mirrorList, i)
	}

	wg.Wait()
	// [3.]
	//******************************************************************************************************************
	sort.Slice(mirrorList[:], func(i, j int) bool {
		return mirrorList[i].DownloadSpeed > mirrorList[j].DownloadSpeed
	})
	fmt.Println("[ID] #name #downlaod speed # ping #url")
	for i, m := range mirrorList {
		if m.Ping == 9999.99 {
			fmt.Printf("[%2v] %s | %3.2f KiB/s| %s | https:// %s  \n", i+1, m.Name, m.DownloadSpeed, "NaN", m.Url)
		} else {
			fmt.Printf("[%2v] %s | %3.2f KiB/s | %3.2f ms | https:// %s \n", i+1, m.Name, m.DownloadSpeed, m.Ping, m.Url)
		}
	}
}

//[Data structures]
//**********************************************************************************************************************

type Mirror struct {
	Name          string
	BaseUrl       string
	Url           string
	TestFile      string
	DownloadSpeed float64
	Ping          float64
}
type osRelease struct {
	// https://www.freedesktop.org/software/systemd/man/os-release.html
	// Note that ArchLinux doesn't have VERSION_ID
	id        string  // ID : opensuse-tumbleweed | opensuse
	versionId float64 // VERSION_ID : 20210311    | 15.2
}

// MirrorDB mirror-list.json mapping
type MirrorDB []_Mirror
type _Mirror struct {
	Name    string `json:"name"`
	Ip      string `json:"ip"`
	Url     string `json:"url"`
	Mapping []struct {
		Distro string   `json:"distro"`
		Path   []string `json:"path"`
	} `json:"mapping"`
}

// ANSI escape codes
const (
	Esc       = "\033["
	EraseLine = Esc + "K"
	Left      = Esc + "1D"
)

//[MirrorList processing]
//**********************************************************************************************************************
func getMyMirrorList(mirrordb MirrorDB, distro string) []Mirror {
	// "distro" format:
	// -> regular  distros: ID-VERSION_ID (e.g. opensuse-15.3)
	// -> rolling releases: ID            (e.g. "opensuse-tumbleweed" or "arch")

	var mirrorList []Mirror
	var urlBuilder strings.Builder

	for _, mir := range mirrordb {
		for _, mapping := range mir.Mapping {
			if mapping.Distro == distro {

				urlBuilder.WriteString(mir.Url)
				urlBuilder.WriteString(mapping.Path[0])

				// Note that even we are keep appending
				// Go seems to have a quite efficient slice
				// TODO: compare efficiencies of slice and list.List for massive appending
				mirrorList = append(
					mirrorList,
					Mirror{Name: mir.Name,
						BaseUrl:       mir.Url,
						Url:           urlBuilder.String(),
						TestFile:      mapping.Path[1],
						DownloadSpeed: 0})
				urlBuilder.Reset()
			}
		}
	}
	return mirrorList
}

// [Utils]
//**********************************************************************************************************************

func check(e error, s string) {
	if e != nil {
		println(s)
		panic(e)
	}
}

func NewOsRelease() *osRelease {
	return &osRelease{id: "", versionId: 0}
}

func getSystemName() string {
	f, err := ioutil.ReadFile("/etc/os-release")
	check(err, "Cannot read /etc/os-release")
	scanner := bufio.NewScanner(bytes.NewReader(f))

	for scanner.Scan() {
		var arr = strings.Split(scanner.Text(), "=")
		if arr[0] == "ID" {
			return arr[1][1 : len(arr[1])-1]
		}
	}
	// XDG says ID is default to linux
	//https://www.freedesktop.org/software/systemd/man/os-release.html
	return "linux"
}
