# booster
Network interface balancer

[![GoDoc](https://godoc.org/github.com/booster-proj/booster?status.svg)](https://godoc.org/github.com/booster-proj/booster)

## Abstract
While more and more people today have a fast Internet connection, there are plenty of other people that do not. The aim of this project is to create a solution that combines multiple Internet access points (such as Wifi or mobile devices) into one single faster Internet connection, that it is easy to use, and fast to configure.

## Installation & Usage
At the moment it is only possible to build `booster` from source. You'll need to [install go](https://golang.org/doc/install) then.  
(Windows is not yet supported)
  
Afterwards, type this commands into your command line:
`git clone https://github.com/booster-proj/booster.git` # Clones this repository into your current directory  
`make test` # Runs tests  
`make` # Creates the bin/booster executable into the current directory  
