package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jacewalker/ip-monitor/check"
	dbops "github.com/jacewalker/ip-monitor/db"
	"github.com/jacewalker/ip-monitor/routes"
	"github.com/rs/zerolog/log"

	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Warn().Msg("Unable to load .env file")
	}
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

	uptimeTicker := time.NewTicker(60 * time.Second)
	go func() {
		for {
			select {
			case <-uptimeTicker.C:
				check.UptimeCheck(db)
			}
		}
	}()

	r.GET("/", routes.MonitorsRoute)
	r.POST("/add", routes.AddRoute)
	r.GET("/delete/:ipaddr", routes.DeleteCheckRoute)

	r.Run(":80")
}
