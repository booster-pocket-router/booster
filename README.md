<p align="center">
   <img src="https://raw.githubusercontent.com/booster-proj/booster/master/assets/banner.png" alt="Booster" width="200" />
</p>

[![Release](https://img.shields.io/github/release/booster-proj/booster.svg)](https://github.com/booster-proj/booster/releases/latest)
[![GoDoc](https://godoc.org/github.com/booster-proj/booster?status.svg)](https://godoc.org/github.com/booster-proj/booster)
[![Build Status](https://travis-ci.org/booster-proj/booster.svg?branch=master)](https://travis-ci.org/booster-proj/booster)
[![Go Report Card](https://goreportcard.com/badge/github.com/booster-proj/booster)](https://goreportcard.com/report/github.com/booster-proj/booster)

## Abstract
While more and more people today have a fast Internet connection, there are plenty of other people that do not. The aim of this project is to create a solution that combines multiple Internet access points (such as Wifi or mobile devices) into one single faster Internet connection, that it is easy to use, and fast to configure.

## Highscores
First things first: this section is reserved for the highest download speed that we managed to obtain in our office. Without `booster` our WIFI's download speed reaches [**~34Mbps**](https://www.speedtest.net/result/7783615417), but **with booster**... :tada:  
<p align="center">
   <a href="https://www.speedtest.net/result/7777990270"><img src="https://www.speedtest.net/result/7777990270.png"/></a>
</p>

## Installation
*(Windows is not yet supported)*
#### Binary
Pick your [release](https://github.com/booster-proj/booster/releases).
#### Snap
[![Get it from the Snap Store](https://snapcraft.io/static/images/badges/en/snap-store-black.svg)](https://snapcraft.io/booster)  
Note: at the moment `booster` is not able to bind to an interface that points to an Apple device without root privileges. To overcome the issue install the snap as root.
You can always inspect the logs using:
``` bash
snap logs booster -f
```

#### From source
First [install go](https://golang.org/doc/install), then type this commands into your command line:   
``` bash
git clone https://github.com/booster-proj/booster.git && cd booster # Clone
make test # Test
make # Build
```
## Usage
Recap: when `booster` spawns, it identifies the network interfaces available in the system that provide an active internet connection. It then starts a proxy server that speaks either **socks5** or **http**. According to some particular **strategy** (still not configurable), and a set of **policies** (configurable), the server is able to **distribute** the incoming network traffic across the network interfaces collected.

Note that `booster` runs as daemon when installed through `snap`, otherwise you'll have to start it manually:
``` bash
bin/booster
```
Note: get help with the `--help` flag.

Once started, `booster` can be remotely controller through its public HTTP Json API. The documentation is available in the [Wiki](https://github.com/booster-proj/booster/wiki/API-Documentation).
