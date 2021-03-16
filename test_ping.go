package main

import (
	"bufio"
	"bytes"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

func getAveragePing(url string, wg *sync.WaitGroup, mirrorList []Mirror, index int) {
	defer wg.Done()

	var cmd strings.Builder
	cmd.WriteString("ping -c 3 -q ")
	cmd.WriteString(url)

	output, err := exec.Command("/bin/sh", "-c", cmd.String()).Output()
	if err != nil {
		mirrorList[index].Ping = 9999.99
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
				mirrorList[index].Ping = 9999.99
				return
			}
		} else if lc == 4 {
			received := strings.Split(scanner.Text(), "/")[4]
			avgRTT, err := strconv.ParseFloat(received, 32)
			check(err, "Cannot parser float for ping")

			mirrorList[index].Ping = avgRTT
			return
		}
		lc++

	}
	mirrorList[index].Ping = 9999.99
}
