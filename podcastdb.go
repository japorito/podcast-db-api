package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/japorito/podcast-db-api/conf"
	"net/http"
	// "golang.org/x/net/context"
	// "google.golang.org/api/sheets/v4"
)

func parseFlags(confPath *string) {
	const (
		defaultConf = "/etc/podcastdb.conf"
		confUsage   = "Specify the path to the machine level configuration file."
	)

	flag.StringVar(confPath, "configuration", defaultConf, confUsage)
	flag.StringVar(confPath, "c", defaultConf, confUsage)

	flag.Parse()
}

func main() {
	var confPath string

	parseFlags(&confPath)

	if machine, err := conf.SetPrimaryConfiguration(confPath); err == nil {
		if machine.GetBool("production") {
			gin.SetMode(gin.ReleaseMode)
		} else {
			gin.SetMode(gin.DebugMode)
		}

		r := gin.Default()

		r.GET("/hello", func(c *gin.Context) {
			if messages, err := conf.GetConfiguration("strings-en"); err == nil {
				c.JSON(http.StatusOK, gin.H{"message": messages.GetString("helloworld-message")})
			} else {
				fmt.Errorf("%v\nError reading configuration. Exiting...", err)
				return
			}
		})

		r.Run(fmt.Sprintf("%v:%v", machine.GetString("ip"), machine.GetString("port")))
	} else {
		fmt.Errorf("%v\nError reading configuration. Exiting...", err)

		return
	}
}
