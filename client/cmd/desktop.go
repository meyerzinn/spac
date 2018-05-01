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
	"github.com/spf13/cobra"
	"github.com/faiface/pixel/pixelgl"
	"github.com/jakecoffman/cp"
	"github.com/20zinnm/spac/common/physics/world"
	"github.com/20zinnm/spac/common/physics"
	"os"
	"os/signal"
	"time"
	"github.com/20zinnm/entity"
	"github.com/faiface/pixel"
	"github.com/20zinnm/spac/client/rendering"
	"github.com/20zinnm/spac/client/networking"
)

var (
	host = "localhost:8080"
)

// desktopCmd represents the desktop command
var desktopCmd = &cobra.Command{
	Use:   "desktop",
	Short: "Run the spac desktop client",
	//Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		pixelgl.Run(func() {
			cfg := pixelgl.WindowConfig{
				Title:  "spac",
				Bounds: pixel.R(0, 0, 1024, 768),
				VSync:  true,
			}
			win, err := pixelgl.NewWindow(cfg)
			if err != nil {
				panic(err)
			}

			var manager entity.Manager
			manager.AddSystem(rendering.New(win))
			manager.AddSystem(networking.New(&manager, win, host))
			done := make(chan struct{})
			interrupt := make(chan os.Signal, 1)
			signal.Notify(interrupt, os.Interrupt)
			start := time.Now()
			for {
				select {
				case <-interrupt:
					close(done)
					return
				default:
					t := time.Now()
					delta := t.Sub(start).Seconds()
					start = t
					manager.Update(delta)
				}
			}
		})
	},
}

func init() {
	rootCmd.AddCommand(desktopCmd)

	desktopCmd.Flags().StringVar(&host, "host", "localhost:8080", "Host to connect to.")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// desktopCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// desktopCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
