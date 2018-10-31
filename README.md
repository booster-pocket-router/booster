# booster
Network interface balancer

[![GoDoc](https://godoc.org/github.com/booster-proj/booster?status.svg)](https://godoc.org/github.com/booster-proj/booster)

## Abstract
While more and more people today have a fast Internet connection, there are plenty of other people that do not. The aim of this project is to create a solution that combines multiple Internet access points (such as Wifi or mobile devices) into one single faster Internet connection, that it is easy to use, and fast to configure.

## Installation & Usage
At the moment it is only possible to build `booster` from source. You'll need to [install go](https://golang.org/doc/install) then.  
(Windows is not yet supported)
  
Afterwards, type this commands into your command line:  
``` bash
git clone https://github.com/booster-proj/booster.git # Clones this repository into your current dir
make test # Runs tests
make # Creates the bin/booster executable into the current directory
```   

`booster` needs to retrieve the network interfaces that provide a network connection to work. When it starts, it retrieves them (it is also possible to filter the network interfaces by name above the other default filers, using the option `iname`. On macOS I always set it to "en"). Afterwards it spawns a proxy serverusing the protocol specified by the `proto` flag (just use **socks5**), which will fetch the data from the sources provided, according to some strategy. At the moment we have only implemented a naive round robin fashion.  
  
##### Run:
``` bash
bin/booster --help
```   
For help.  
##### Session:
Setup:  
I plug my iPhone 5s (with tethering enabled, iOS 12) into my MacBook Pro (macOS 10.14),
Run:
``` bash
bin/booster -iname=en -proto=socks5
```   
Last:  
 - System Preferences > Network > Advanced... > Proxies  
 - select: SOCKS Proxy, localhost : { port from previous command's output, by default **1080** }   
 - https://www.speedtest.net 🤓

