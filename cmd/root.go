// Copyright © 2019 KIM KeepInMind GmbH/srl
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"fmt"
	stdLog "log"
	"os"

	"github.com/spf13/cobra"
	"upspin.io/log"
)

var (
	// Log configuration
	verbose  bool
	cleanLog bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "booster",
	Short: "Handle a booster server",
	Long: `Use booster to start a server that will allow to build a powerfull multihomed system.
Use its SOCKS5 proxy to pipe your network traffic though booster's balancing techniques.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Setup logger
		setupLogger(verbose, cleanLog)

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "If set, makes the logger print also debug messages")
	rootCmd.PersistentFlags().BoolVar(&cleanLog, "clean-log", false, "If set, assumes that the loggin is handled by a third party entity")
}

func setupLogger(verbose bool, clean bool) {
	level := log.InfoLevel
	if verbose {
		log.SetLevel("debug")
		level = log.DebugLevel
	}
	if clean {
		log.SetOutput(nil)             // disable "local" logging
		log.Register(newLogger(level)) // enable "remote" (snapcraft's daemon handled logger usually) logging
	}
}

type logger struct {
	defaultLogger log.Logger
	level         log.Level
}

func newLogger(level log.Level) *logger {
	return &logger{
		level:         level,
		defaultLogger: stdLog.New(os.Stderr, "", 0), // Do not add date/time information
	}
}

func (l *logger) Log(level log.Level, msg string) {
	if level < l.level {
		return
	}

	l.defaultLogger.Println(msg)
}

func (l *logger) Flush() {
}
