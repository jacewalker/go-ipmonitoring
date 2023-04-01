package routes

import (
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/gin-gonic/gin"
	"github.com/jacewalker/ip-monitor/check"
	dbops "github.com/jacewalker/ip-monitor/db"
)

var db = dbops.Init()

func HomeRoute(c *gin.Context) {
	c.HTML(http.StatusOK, "home.html", nil)
}

func AddRoute(c *gin.Context) {
	activeIP := check.ParseCheck(c)
	go func() {
		if err := check.ScanPorts(&activeIP); err != nil {
			log.Error().Msg("unable to scan ports")
		}
		dbops.SaveToDatabase(db, activeIP)
	}()

	c.HTML(http.StatusOK, "home.html", gin.H{
		"status": "added",
		"ports":  activeIP.OpenPorts,
	})
}

func MonitorsRoute(c *gin.Context) {
	allChecks, _ := dbops.GetAllFromDatabase(db)
	ips := make(map[string]string)

	for _, check := range allChecks {
		ips[check.Address] = check.OpenPorts
	}

	fmt.Println("Printing IPs...")
	fmt.Println(ips)

	c.HTML(http.StatusOK, "monitors.html", gin.H{
		"ips": ips,
	})
}
