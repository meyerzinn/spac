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
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"time"
	"github.com/gorilla/websocket"
	"net/http"
	"github.com/jakecoffman/cp"
	"github.com/20zinnm/spac/common/physics/world"
	"github.com/20zinnm/entity"
	"github.com/20zinnm/spac/server/movement"
	"github.com/20zinnm/spac/server/perceiving"
	"github.com/20zinnm/spac/server/networking"
	"github.com/20zinnm/spac/server/health"
	"github.com/20zinnm/spac/common/physics"
	"github.com/20zinnm/spac/server/shooting"
	"log"
	"github.com/20zinnm/spac/common/net"
)

var cfgFile string
var (
	worldRadius float64 = 10000
	addr                = ":8080"
	tick                = time.Millisecond * 20
	upgrader            = &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	//logger *zap.Logger
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "server",
	Short: "Run a spac server",
	Long:  ``,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		var manager entity.Manager
		// set up space
		space := cp.NewSpace()
		// > rules
		space.SetDamping(0.3)
		space.SetGravity(cp.Vector{0, 0})

		world := world.New(space)
		manager.AddSystem(health.New(space))
		manager.AddSystem(movement.New(world))
		manager.AddSystem(shooting.New(&manager, world))
		manager.AddSystem(physics.New(&manager, world))
		manager.AddSystem(perceiving.New(world))
		netwk := networking.New(&manager, world)
		manager.AddSystem(netwk)
		go func() {
			var ticks int64
			ticker := time.NewTicker(tick)
			var skipped int64
			start := time.Now()
			for t := range ticker.C {
				select {
				case <-ticker.C: // if the next tick is immediately available (means we're lagging)
					skipped++
					break
				default:
					if skipped > 0 {
						fmt.Printf("skipping %d ticks; is the server lagging?", skipped)
						ticks += skipped
						skipped = 0
					}
					delta := t.Sub(start).Seconds()
					start = t
					manager.Update(delta)
					ticks++
				}
			}
		}()
		fmt.Println("game started")
		http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				fmt.Println("error upgrading connection", err)
				return
			}
			netwk.Add(net.Websocket(conn))
			fmt.Println("client connected")
		}))
		log.Fatal(http.ListenAndServe(addr, nil))
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
	cobra.OnInitialize(initConfig)
	rootCmd.Flags().StringVar(&addr, "addr", ":8080", "Accept incoming requests at this address.")
	rootCmd.Flags().DurationVarP(&tick, "tick", "t", 20*time.Millisecond, "Duration of a game tick.")
	rootCmd.Flags().Float64Var(&worldRadius, "radius", 10000, "Radius of the game world.")
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	//rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cmd.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".cmd" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".cmd")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
