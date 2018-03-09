// Copyright © 2018 Mateusz Dyminski <dyminski@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/mateuszdyminski/cache/engine"
)

var peers []string
var port int

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cache",
	Short: "Dead simple distributed cache with REST based API",
	Long:  `TBD`,
	Run: func(cmd *cobra.Command, args []string) {
		eng, err := engine.NewEngine(peers, port)
		if err != nil {
			panic(err)
		}

		eng.Start()
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
	rootCmd.PersistentFlags().StringSliceVarP(&peers, "peers", "n", nil, "list of peers to synchronise the state with. eg. [localhost:5001, localhost:5002]")
	rootCmd.PersistentFlags().IntVarP(&port, "port", "p", 5555, "http port to start cache with. eg. 5555")
	rootCmd.MarkPersistentFlagRequired("peers")
	rootCmd.MarkPersistentFlagRequired("port")
}
