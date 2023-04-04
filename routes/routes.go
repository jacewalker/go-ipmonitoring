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

// Parse the input as either a net.IP or net.CIDR then call the respective function to scan the IP address(es).
func AddRoute(c *gin.Context) {
	activeCheck := check.ParseCheck(c)

	switch activeCheck.ScanType {
	case "subnet":
		go check.ScanSubnet(db, activeCheck)
	case "ip":
		go check.ScanIP(db, activeCheck)
	default:
		log.Error().Msg("Missing scan type.")
	}

	MonitorsRoute(c)
}

// Get all checks from the database then render the home page.
func MonitorsRoute(c *gin.Context) {
	allChecks, _ := dbops.GetAllFromDatabase(db)
	ips := make(map[string]string)

	for _, check := range allChecks {
		ips[check.Address] = check.OpenPorts
	}

	c.HTML(http.StatusOK, "home.html", gin.H{
		"checks": allChecks,
	})
}

func DeleteCheckRoute(c *gin.Context) {
	address := c.Param("ipaddr")
	fmt.Println("Deleting... ", address)

	ch, err := dbops.LookupCheck(db, address)
	if err != nil {
		log.Warn().Msgf("unable to lookup check, error: %s", err)
		c.Redirect(http.StatusTemporaryRedirect, "/")
	}
	if dbops.DeleteCheck(db, ch) {
		c.Redirect(http.StatusTemporaryRedirect, "/")
	} else {
		c.HTML(http.StatusOK, "home.html", gin.H{
			"error": "Unable to delete check. Try again",
		})
	}
}
