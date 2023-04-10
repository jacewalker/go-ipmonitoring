package check

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/Ullaakut/nmap/v2"
	"github.com/gin-gonic/gin"
	dbops "github.com/jacewalker/ip-monitor/db"
	"github.com/rs/zerolog/log"
)

func ParseCheck(c *gin.Context) dbops.Check {
	var check dbops.Check
	var input string = c.PostForm("ipaddr")
	var label string = c.PostForm("label")
	var email string = c.PostForm("email")

	// if c.Request.FormValue("uptime-option") == "checked" {
	// 	check.Monitoring = dbops.MonitoringType{
	// 		Uptime: true,
	// 	}
	// } else if c.Request.FormValue("openport-option") == "checked" {
	// 	check.Monitoring = dbops.MonitoringType{
	// 		OpenPort: true,
	// 	}
	// } else if c.Request.FormValue("vulnerability-option") == "checked" {
	// 	check.Monitoring = dbops.MonitoringType{
	// 		Vulnerability: true,
	// 	}
	// }
	// fmt.Println("Monitors selected: ", check.Monitoring)

	// Parse the input as a Subnet
	addresses, err := parseSubnet(c, &input)
	fmt.Println("Addresses: ", addresses)
	if err != nil {
		log.Info().Msg("Parse subnet failed. Proceeding as single IP.")
	} else {
		check.Addresses = addresses
		check.Label = label
		check.ScanType = "subnet"
		check.Email = email
		return check
	}

	// Parse the input as an IP
	ip := net.ParseIP(input)
	if ip != nil {
		check.Address = ip.String()
		check.Label = label
		check.ScanType = "ip"
		check.Email = email
		return check
	} else {
		check.Error = fmt.Sprintln("malformed input")
		return check
	}
}

func ScanPorts(ch *dbops.Check) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	openPorts := []int{}

	scanner, err := nmap.NewScanner(
		nmap.WithTargets(ch.Address),
		nmap.WithMostCommonPorts(1000),
		nmap.WithContext(ctx),
	)
	if err != nil {
		log.Info().Msgf("Unable to create scanner:", err)
		return errors.New("unable to create scanner")
	}

	results, warnings, err := scanner.Run()
	if err != nil {
		log.Info().Msgf("Unable to run scanner:", err)
		return errors.New("unable to run scanner")
	}

	if warnings != nil {
		log.Printf("Warnings: \n %v", warnings)
	}

	// if host is down, set error
	// if len(results.Hosts) == 0 {
	// 	ch.Error = "host down"
	// 	return errors.New("host down")
	// }

	for _, host := range results.Hosts {
		if len(host.Ports) > 0 && host.Status.State == "up" {
			for _, port := range host.Ports {
				if port.State.String() == "open" {
					openPorts = append(openPorts, int(port.ID))
				}
			}
		}
	}

	dbops.PortsToString(ch, openPorts)
	log.Info().Msg(ch.OpenPorts)
	return nil
}
