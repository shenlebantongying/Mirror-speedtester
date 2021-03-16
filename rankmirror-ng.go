// SPDX-License-Identifier:  BSD-3-Clause

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

// Architecture note
// To simplify implementation, reduce maintenance cost and ease of life,
// This program just invoke existing commands like curl, ping, host, then parse the result...

// Steps
// 1. Parse mirror-list.json -> load into `mirrorDB` -> parse into a convenient `mirrorList`
// 2. Iterate `mirrorlist` to get all urls, then perform each tests for each url.
// 3. Aggregate

// Add new tests result:
// 1. Add a new field to the end of `type Mirror struct`
// 2. make a function to obtain the field
// 3. invoke the function during step [2.]

func main() {

	// [1.]
	var mirrorDB MirrorDB

	f, err := ioutil.ReadFile("./mirror-list.json")
	check(err, "Cannot read /etc/os-re lease")
	err = json.Unmarshal(f, &mirrorDB)
	check(err, "mirror-list.json might be invalid")

	mirrorList := getMyMirrorList(mirrorDB, getSystemName())

	// [2.]

	for i, m := range mirrorList {
		fmt.Printf("Testing [%v/%v] %s \r", i+1, len(mirrorList), m.Name)
		mirrorList[i].DownloadSpeed = getAverageDownloadSpeed(m.Url + m.TestFile)
		mirrorList[i].Ping = getAveragePing(m.BaseUrl)
		fmt.Printf(EraseLine)
	}

	// [3.]

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

//**********************************************************************************************************************
// Utils

func check(e error, s string) {
	if e != nil {
		println(s)
		panic(e)
	}
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

type Mirror struct {
	Name          string
	BaseUrl       string
	Url           string
	TestFile      string
	DownloadSpeed float64
	Ping          float64
}

// ANSI escape codes
const (
	Esc       = "\033["
	EraseLine = Esc + "K"
	Left      = Esc + "1D"
)

//**********************************************************************************************************************

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

//**********************************************************************************************************************

func getAveragePing(url string) float64 {
	var cmd strings.Builder
	cmd.WriteString("ping -c 3 -q ")
	cmd.WriteString(url)

	output, err := exec.Command("/bin/sh", "-c", cmd.String()).Output()
	if err != nil {
		return 9999.99
	}

	// Code below are parsing this:
	//ping -c 5 -q google.com
	//lc->
	//0 -> PING google.com (172.217.1.174) 56(84) bytes of data.
	//1 ->
	//2 -> --- google.com ping statistics ---
	//3 -> 5 packets transmitted, _4_ received, 20% packet loss, time 4005ms
	//4 -> rtt min/avg/max/mdev = 37.594/_37.950_/38.302/0.270 ms

	lc := 0 //line counter for scanner
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		if lc == 3 {
			received := strings.Split(scanner.Text(), " ")[3]
			if received == "0" {
				return 9999.99
			}
		} else if lc == 4 {
			received := strings.Split(scanner.Text(), "/")[4]
			avgRTT, err := strconv.ParseFloat(received, 32)
			check(err, "Cannot parser float for ping")
			return avgRTT
		}
		lc++

	}
	return 9999.99
}

func BytesToKiBs(n int) float64 {
	//1 KiB/s = 1024 Bytes/s
	return float64(n) / 1024.0
}

func getAverageDownloadSpeed(url string) float64 {
	var cmd strings.Builder
	// &{speed_download} -> Bytes per second
	cmd.WriteString("curl -s -w \"%{speed_download}\" -o /dev/null -L ")
	cmd.WriteString(url)

	output, err := exec.Command("/bin/sh", "-c", cmd.String()).Output()
	check(err, "No curl found on the system")

	if string(output) == "" {
		return 0
	}

	downSpeedBytes, err := strconv.Atoi(string(output))
	check(err, "curl return format error")

	return BytesToKiBs(downSpeedBytes)
}

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

//**********************************************************************************************************************

func NewOsRelease() *osRelease {
	return &osRelease{id: "", versionId: 0}
}
