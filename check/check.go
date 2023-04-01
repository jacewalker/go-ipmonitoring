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
	"gorm.io/gorm"
)

func ParseCheck(c *gin.Context) dbops.Check {
	check := dbops.Check{}
	// introduce IP validation and subnet parsing
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
	log.Info().Msg(ch.OpenPorts)
	return nil
}

func DailyPortCheck(db *gorm.DB) {
	fmt.Printf("\n")
	log.Info().Msgf("Starting daily check...")
	monitors, err := dbops.GetAllFromDatabase(db)
	if err != nil {
		log.Info().Msgf("Error getting monitors from database:", err)
	}

	for _, monitor := range monitors {
		fmt.Printf("\n")
		log.Info().Msgf("Checking %s", monitor.Address)
		oldOpenPorts, err := dbops.StringToPorts(monitor.OpenPorts)
		if err != nil {
			log.Info().Msgf("Error converting open ports string to slice:", err)
			continue
		}

		newMonitor := monitor

		err = ScanPorts(&newMonitor)
		if err != nil {
			log.Info().Msgf("Error scanning ports:", err)
			continue
		}

		newOpenPorts, _ := dbops.StringToPorts(newMonitor.OpenPorts)

		_, totalDiff := dbops.GetOpenPortDifferences(oldOpenPorts, newOpenPorts)
		if len(totalDiff) != 0 {
			log.Info().Msg("Ports have changed.")
			dbops.DeleteFromDatabase(db, monitor)
			dbops.SaveToDatabase(db, newMonitor)
			if !notifications.SendEmailNotification(newMonitor) {
				log.Warn().Msg("Unable to send email notification")
				continue
			}
		} else {
			log.Info().Msg("Ports have not changed.")
		}

	}
}
