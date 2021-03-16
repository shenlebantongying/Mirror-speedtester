package main

import (
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

func getAverageDownloadSpeed(url string, wg *sync.WaitGroup, mirrorList []Mirror, index int) {
	defer wg.Done()

	var cmd strings.Builder
	// &{speed_download} -> Bytes per second
	cmd.WriteString("curl -s -w \"%{speed_download}\" -o /dev/null -L ")
	cmd.WriteString(url)
	output, err := exec.Command("/bin/sh", "-c", cmd.String()).Output()
	check(err, "Curl Fatal: "+cmd.String())

	if string(output) == "" {
		mirrorList[index].DownloadSpeed = 0
	}

	downSpeedBytes, err := strconv.Atoi(string(output))
	check(err, "curl return format error")

	mirrorList[index].DownloadSpeed = BytesToKiBs(downSpeedBytes)
	//return BytesToKiBs(downSpeedBytes)
}

func BytesToKiBs(n int) float64 {
	//1 KiB/s = 1024 Bytes/s
	return float64(n) / 1024.0
}
