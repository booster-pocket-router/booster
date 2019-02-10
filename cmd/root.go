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
	verbose     bool
	externalLog bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "booster",
	Short: "",
	Long:  ``,

	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Setup logger
		setupLogger(verbose, externalLog)

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
	rootCmd.PersistentFlags().BoolVar(&externalLog, "external-log", false, "If set, assumes that the loggin is handled by a third party entity")
}

func setupLogger(verbose bool, external bool) {
	level := log.InfoLevel
	if verbose {
		log.SetLevel("debug")
		level = log.DebugLevel
	}
	if external {
		log.SetOutput(nil)                     // disable "local" logging
		log.Register(newExternalLogger(level)) // enable "remote" (snapcraft's daemon handled logger usually) logging
	}
}

type externalLogger struct {
	defaultLogger log.Logger
	level         log.Level
}

func newExternalLogger(level log.Level) *externalLogger {
	return &externalLogger{
		level:         level,
		defaultLogger: stdLog.New(os.Stderr, "", 0), // Do not add date/time information
	}
}

func (l *externalLogger) Log(level log.Level, msg string) {
	if level < l.level {
		return
	}

	l.defaultLogger.Println(msg)
}

func (l *externalLogger) Flush() {
}
