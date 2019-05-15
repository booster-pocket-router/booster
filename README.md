# booster
[![GoDoc](https://godoc.org/github.com/booster-proj/booster?status.svg)](https://godoc.org/github.com/booster-proj/booster)
[![Build Status](https://travis-ci.org/booster-proj/booster.svg?branch=master)](https://travis-ci.org/booster-proj/booster)
[![Go Report Card](https://goreportcard.com/badge/github.com/booster-proj/booster)](https://goreportcard.com/report/github.com/booster-proj/booster)
[![Release](https://img.shields.io/github/release/booster-proj/booster.svg)](https://github.com/booster-proj/booster/releases/latest)
[![booster](https://snapcraft.io/booster/badge.svg)](https://snapcraft.io/booster)

## Abstract
I would like to see a world where I can **tune** my devices network flow with no effort. I would like also to be able to communicate to my programs the things that I know, instread of making them figure out things *only* on their own. **What is the issu here?** tbc


## Who might be interested in this project?
We're trying to solve one by one some real usecases, either things that came up to our mind or requested features from the community. If you think that you have a problem that `booster` may solve, you're highly encouraged to either contact us (booster@keepinmind.info) or to [file a new feature request](https://github.com/booster-proj/booster/issues/new?template=feature_request.md)!

#### Gamers
Having lag or jitter problems (e.g. ping that is not constant over time, check [this](https://www.speedtest.net/help) out for clarifications). With `booster` we want to "reserve" a slice of the overall network channel for the game beign played to provide a smooth gaming experience, while using the rest of the bonded network connection for the other actions, such as [Window's background auto-updates](https://answers.microsoft.com/en-us/windows/forum/all/how-do-i-stop-windows-10-from-ruining-my-gaming/227e3fbe-88b1-46ba-bfdd-38b71e17607e), or maybe watching a movie over the network (issue [#41](https://github.com/booster-proj/booster/issues/41)).

#### Travellers/Slow ADLS owners
Having problems downloading/uploading data over the Internet. For example when you find yourself at your friend's place, you want to watch a movie together but the ADSL at his/her home is too slow. With `booster` we can bond the ADSL, both your and your friend's LTE networks, apply rules on how the different sources are drained, and provide a faster network access point.

`booster` already shows benefits for solving this usecase: without `booster` our offices WIFI's download speed reaches [**~34Mbps**](https://www.speedtest.net/result/7783615417), but **with booster**, using both @philip's and my phone's mobile network connection we managed to obtain [**~155Mbps**](https://www.speedtest.net/result/7777990270)! :tada: (issue [#42](https://github.com/booster-proj/booster/issues/42)).

#### Creative people
That want to get involved, have some feedback, know something that might be helpful.. in any case you're very welcome! 😊
