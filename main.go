package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jacewalker/ip-monitor/check"
	"github.com/jacewalker/ip-monitor/routes"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env")
	r := gin.Default()
	r.LoadHTMLGlob("./views/*.html")

	// Run the DailyCheck function once per day
	dailyTicker := time.NewTicker(20 * time.Second)
	go func() {
		for {
			select {
			case <-dailyTicker.C:
				check.DailyCheck()
			}
		}
	}()

	r.GET("/", routes.HomeRoute)
	r.POST("/add", routes.AddRoute)
	r.GET("/show", routes.MonitorsRoute)

	r.Run(":8080")
}
