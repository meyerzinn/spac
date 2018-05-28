// Copyright © 2018 Meyer Zinn <meyerzinn@gmail.com>
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
	"github.com/20zinnm/spac/server/server"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/profile"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"time"
)

var cfgFile string
var prof bool
var (
	debug       bool
	worldRadius float64
	addr        string
	tick        time.Duration
	//logger *zap.Logger
)

var rootCmd = &cobra.Command{
	Use:   "server",
	Short: "Run a spac server",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if prof {
			defer profile.Start(profile.ProfilePath(".")).Stop()
		}
		var options []server.Option
		if debug {
			options = append(options, server.Debug())
		}
		options = append(options, server.TickRate(tick), server.BindAddress(addr), server.WorldRadius(worldRadius))
		server.Start(options...)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.Flags().StringVar(&addr, "addr", ":7722", "Accept incoming requests at this address.")
	rootCmd.Flags().Float64Var(&worldRadius, "radius", 10000, "Radius of the game world.")
	rootCmd.Flags().DurationVarP(&tick, "tick", "t", time.Second/20, "Duration of a game tick")
	rootCmd.Flags().BoolVar(&prof, "profile", false, "Enable performance profiling.")
	rootCmd.Flags().BoolVar(&debug, "debug", false, "Enable debugging endpoint.")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		viper.AddConfigPath(home)
		viper.SetConfigName(".cmd")
	}
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
