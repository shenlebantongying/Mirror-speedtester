# Rankmirror-NG [WIP]
[![Go Report Card](https://goreportcard.com/badge/github.com/shenlebantongying/rankmirror-ng)](https://goreportcard.com/report/github.com/shenlebantongying/rankmirror-ng)

A utility written in `Go` to test mirrors quality for openSUSE and possibly various linux distros .This program provides similar functionalities as Archlinux's [rankmirror](https://git.archlinux.org/pacman-contrib.git/tree/src/rankmirrors.sh.in) and [reflector](https://xyne.archlinux.ca/projects/reflector/) or Debian/Ubuntu's [netselect-apt](https://packages.debian.org/stretch/netselect-apt).

This repo also maintain a mirror list for openSUSE.

# Usage

![Peek](https://user-images.githubusercontent.com/20123683/111223124-1bfae600-85b3-11eb-948f-0ca547c9f66f.gif)
![image](https://user-images.githubusercontent.com/20123683/111234072-2bcef600-85c4-11eb-82e4-6543f6d62da0.png)

_Only support openSUSE-Tumbleweed for now_

## Add more mirrors for testing 

mirror-list.json format

Mapping:
"distro" use the variables from `/etc/os-releases` as id
* regular distros: ID-VERSION_ID
* rolling releases: ID
* "path" => path to the repo and download test file

Example
```json
{
  "name": "tuna",
  "ip": "101.6.8.193",
  "url": "mirrors.tuna.tsinghua.edu.cn",
  "mapping": [
    {
      "distro": "arch",
      "path": ["/archlinux/core/os/x86_64/", "core.files"]
    },
    {
      "distro": "debian-8",
      "path": ["/debian/", "extrafiles"]
    },
    {
      "distro": "opensuse-15.2",
      "path": ["/opensuse/distribution/leap/15.2/", "repo/oss/INDEX.gz"]
    },
    {
      "distro": "opensuse-tumbleweed",
      "path": ["/opensuse/tumbleweed/", "repo/oss/INDEX.gz"]
    }
  ]
}
```

# Roadmap

+ [ ] concurrency to speed up tests
+ [ ] Merge more from [marguerite/rankmirror-ng](https://github.com/marguerite/rankmirror-ng)
+ [ ] Automatically change mirror via zypper (and more)
+ [ ] complete cli interface
+ [ ] probing https, http, ftp
+ [ ] probing ipv6 & ipv4
+ [ ] colorful output

# Acknowledgement

This program borrowed a lot of codes from [marguerite/rankmirror-ng](https://github.com/marguerite/rankmirror-ng).
