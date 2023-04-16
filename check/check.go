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

	check.Label = c.PostForm("label")
	check.Email = c.PostForm("email")

	// Parse the input as a Subnet
	addresses, err := parseSubnet(c, &input)
	fmt.Println("Addresses: ", addresses)
	if err != nil {
		log.Info().Msg("Parse subnet failed. Checking for IPv4.")
	} else {
		check.Addresses = addresses
		check.ScanType = "subnet"
		dbops.SaveToDatabase(db, &check)
		return check
	}

	// Parse the input as an IP
	ip := net.ParseIP(input)
	if ip != nil {
		check.Address = ip.String()
		check.ScanType = "ip"
		dbops.SaveToDatabase(db, &check)
		return check
	} else {
		log.Info().Msg("IPv4 check failed. Checking FQDN.")
	}

	// Parse the input as a FQDN
	ips, err := net.LookupIP(input)
	if err != nil {
		log.Warn().Msg("malformed input")
	} else if len(ips) > 0 {
		check.Address = ips[0].To4().String()
		check.Hostname = input
		check.ScanType = "fqdn"
		dbops.SaveToDatabase(db, &check)
		return check
	}

	dbops.SaveToDatabase(db, &check)
	return check
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
	ch.PortScanCompleted = true
	log.Info().Msg(ch.OpenPorts)
	return nil
}
