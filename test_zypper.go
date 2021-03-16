package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

var (
	zypperCmd string
)

func init() {
	cacheDir, e := os.UserCacheDir()
	check(e, "~./.cache/ is inaccessible")
	zypperCmd = "sudo zypper --no-gpg-checks --root=" + filepath.Join(cacheDir, "tumbleweed") + " "
}

func zypperRefresh() {
	c := zypperCmd + "refresh"
	fmt.Println(c)
	_, err := exec.Command("/bin/sh", "-c", c).Output()
	check(err, "Failed: "+c)
}

func zypperRemoveRepos() {
	c := zypperCmd + "rr -a"
	fmt.Println(c)
	_, err := exec.Command("/bin/sh", "-c", c).Output()
	check(err, "Failed: "+c)
}

func zypperAddRepo(url string) {
	c := zypperCmd + "ar -c " + url + " Tumbleweed"
	fmt.Println(c)
	_, err := exec.Command("/bin/sh", "-c", c).Output()
	check(err, "Failed: "+c)

}

func zypperInNano() {
	c := zypperCmd + "install --no-recommends --dry-run --download-only nano"
	fmt.Println(c)
	_, err := exec.Command("/bin/sh", "-c", c).Output()
	check(err, "Failed: "+c)
}

func zypperSingleRepo(url string) {
	zypperRemoveRepos()
	zypperAddRepo(url)
}
