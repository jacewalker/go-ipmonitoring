package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jacewalker/ip-monitor/check"
	dbops "github.com/jacewalker/ip-monitor/db"
	"github.com/jacewalker/ip-monitor/routes"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env")
	r := gin.Default()
	r.Static("/src", "./src")
	r.LoadHTMLGlob("./views/*.html")

	db := dbops.Init()
	dailyTicker := time.NewTicker(24 * time.Hour)
	go func() {
		for {
			select {
			case <-dailyTicker.C:
				check.DailyPortCheck(db)
			}
		}
	}()

	r.GET("/", routes.MonitorsRoute)
	r.POST("/add", routes.AddRoute)
	r.GET("/delete/:ipaddr", routes.DeleteCheckRoute)

	r.Run(":8080")
}
