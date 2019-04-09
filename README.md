# booster
[![GoDoc](https://godoc.org/github.com/booster-proj/booster?status.svg)](https://godoc.org/github.com/booster-proj/booster)
[![Build Status](https://travis-ci.org/booster-proj/booster.svg?branch=master)](https://travis-ci.org/booster-proj/booster)
[![Go Report Card](https://goreportcard.com/badge/github.com/booster-proj/booster)](https://goreportcard.com/report/github.com/booster-proj/booster)
[![Release](https://img.shields.io/github/release/booster-proj/booster.svg)](https://github.com/booster-proj/booster/releases/latest)
[![License: GPL v3](https://img.shields.io/badge/License-GPL%20v3-blue.svg)](http://www.gnu.org/licenses/gpl-3.0)

## Abstract
While more and more people today have a fast Internet connection, there are plenty of other people that do not. The aim of this project is to create a solution that combines multiple Internet access points (LTE, ADSL) into one single tunable network connection.

## Who might be interested in this project?
We're trying to solve one by one some real usecases, either things that came up to our mind or requested features from the community. If you think that you have a problem that `booster` may solve, you're highly encouraged to either contact us (booster at keepinmind dot info) or to [file a new feature request](https://github.com/booster-proj/booster/issues/new?template=feature_request.md)!

#### Gamers
Having lag or jiffer problems (e.g. ping that is not constant over time, check [this](https://www.speedtest.net/help) out for clarifications). With `booster` we want to "reserve" a slice of the overall network channel for the game beign played to provide a smooth gaming experience, while using the rest of the bonded network connection for the other actions, such as [Window's background auto-updates](https://answers.microsoft.com/en-us/windows/forum/all/how-do-i-stop-windows-10-from-ruining-my-gaming/227e3fbe-88b1-46ba-bfdd-38b71e17607e), or maybe watching a movie over the network.

Follow the progress on [this](https://github.com/booster-proj/booster/issues/41) issue.

#### Travellers/Poor ADLS owners
Having problems downloading/uploading multimedia over the Internet. For example when you find yourself at your friend's place, you want to watch a movie together but the ADSL at his/her home is too slow. With `booster` we can bond the ADSL, both your and your friend's LTE networks, apply restrictions (something that internally we call policies) on how the different sources are drained, and provide a faster network access point.

`booster` already shows benefits for solving this usecase: without `booster` our WIFI's download speed reaches [**~34Mbps**](https://www.speedtest.net/result/7783615417), but **with booster**, using both @philip's and my phone's mobile network connection we managed to obtain [**155Mbps**](https://www.speedtest.net/result/7777990270)! :tada:

Follow this usecase's progress on [this](https://github.com/booster-proj/booster/issues/42) issue.

#### Developers & designers
That want to get involved! You're very welcome! :D

## How does it work?


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
Recap: when `booster` spawns, it identifies the network interfaces available in the system that provide an active internet connection. It then starts a **socks5** proxy server. According to some particular **strategy** (still not configurable), and a set of **policies** (configurable), the server is able to **distribute** the incoming network traffic across the network interfaces collected.

Note that `booster` runs as daemon when installed through `snap`, otherwise you'll have to start it manually:
``` bash
bin/booster
```
Note: get help with the `--help` flag.

Once started, `booster` can be remotely controller through its public HTTP Json API. The documentation is available in the [Wiki](https://github.com/booster-proj/booster/wiki/API-Documentation).

## Highscores
First things first: this section is reserved for the highest download speed that we managed to obtain in our office. Without `booster` our WIFI's download speed reaches [**~34Mbps**](https://www.speedtest.net/result/7783615417), but **with booster**... :tada:  
<p align="center">
   <a href="https://www.speedtest.net/result/7777990270"><img src="https://www.speedtest.net/result/7777990270.png"/></a>
</p>

