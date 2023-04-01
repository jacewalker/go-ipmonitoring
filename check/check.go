package check

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Ullaakut/nmap/v2"
	"github.com/gin-gonic/gin"
	dbops "github.com/jacewalker/ip-monitor/db"
	"github.com/jacewalker/ip-monitor/notifications"
	"github.com/rs/zerolog/log"
)

func ParseCheck(c *gin.Context) dbops.Check {
	check := dbops.Check{}
	check.Address = c.PostForm("ipaddr")

	return check
}

func ScanPorts(ch *dbops.Check) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	openPorts := []int{}

	scanner, err := nmap.NewScanner(
		nmap.WithTargets(ch.Address),
		nmap.WithMostCommonPorts(50),
		nmap.WithContext(ctx),
	)
	if err != nil {
		fmt.Println("Unable to create scanner:", err)
		return errors.New("unable to create scanner")
	}

	results, warnings, err := scanner.Run()
	if err != nil {
		fmt.Println("Unable to run scanner:", err)
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
	fmt.Println(ch.OpenPorts)
	return nil
}

func DailyCheck() {
	fmt.Println("Starting daily check...")
	db := dbops.Init()
	monitors, err := dbops.GetAllFromDatabase(db)
	if err != nil {
		fmt.Println("Error getting monitors from database:", err)
	}

	for _, monitor := range monitors {
		oldOpenPorts, err := dbops.StringToPorts(monitor.OpenPorts)
		if err != nil {
			fmt.Println("Error converting open ports string to slice:", err)
			continue
		}

		newMonitor := monitor

		err = ScanPorts(&newMonitor)
		if err != nil {
			fmt.Println("Error scanning ports:", err)
			continue
		}

		newOpenPorts, _ := dbops.StringToPorts(newMonitor.OpenPorts)

		_, totalDiff := dbops.GetOpenPortDifferences(oldOpenPorts, newOpenPorts)
		if len(totalDiff) != 0 {
			dbops.DeleteFromDatabase(db, monitor)
			dbops.SaveToDatabase(db, newMonitor)
			if !notifications.SendEmailNotification(newMonitor) {
				log.Warn().Msg("Unable to send email notification")
				continue
			}
		}

	}
}
