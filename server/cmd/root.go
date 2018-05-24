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
	"github.com/20zinnm/entity"
	"github.com/20zinnm/spac/server/movement"
	"github.com/20zinnm/spac/server/perceiving"
	"github.com/20zinnm/spac/server/networking"
	"github.com/20zinnm/spac/server/health"
	"github.com/20zinnm/spac/server/shooting"
	"log"
	"github.com/20zinnm/spac/common/net"
	commonPhysics "github.com/20zinnm/spac/common/world"
	"github.com/20zinnm/spac/server/physics"
	"github.com/pkg/profile"
)

var cfgFile string
var prof bool
var (
	worldRadius float64 = 10000
	addr                = ":8080"
	upgrader            = &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	tick time.Duration
	//logger *zap.Logger
)

var rootCmd = &cobra.Command{
	Use:   "server",
	Short: "Run a spac server",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if prof {
			defer profile.Start(profile.ProfilePath("."), profile.TraceProfile).Stop()
		}
		var manager entity.Manager
		world := &commonPhysics.World{Space: commonPhysics.NewSpace()}
		manager.AddSystem(health.New(world)) // every 50 ms
		manager.AddSystem(movement.New(world))
		manager.AddSystem(shooting.New(&manager, world))
		manager.AddSystem(physics.New(&manager, world, worldRadius))
		manager.AddSystem(perceiving.New(world))
		netwk := networking.New(&manager, world, worldRadius)
		manager.AddSystem(netwk)
		go func() {
			ticker := time.NewTicker(tick)
			defer ticker.Stop()
			start := time.Now()
			for t := range ticker.C {
				delta := t.Sub(start).Seconds()
				start = t
				manager.Update(delta)
			}
			fmt.Println("game stopped")

		}()
		fmt.Println("game started")
		http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				fmt.Println("error upgrading connection", err)
				return
			}
			netwk.Add(net.Websocket(conn))
		}))
		log.Fatal(http.ListenAndServe(addr, nil))
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
	rootCmd.Flags().StringVar(&addr, "addr", ":8080", "Accept incoming requests at this address.")
	rootCmd.Flags().Float64Var(&worldRadius, "radius", 10000, "Radius of the game world.")
	rootCmd.Flags().DurationVarP(&tick, "tick", "t", time.Millisecond*20, "Duration of a game tick")
	rootCmd.Flags().BoolVar(&prof, "profile", false, "Enable performance profiling.")
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
