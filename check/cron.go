package check

import (
	"fmt"

	dbops "github.com/jacewalker/ip-monitor/db"
	"github.com/jacewalker/ip-monitor/notifications"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

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
			// dbops.DeleteFromDatabase(db, monitor)
			dbops.DeleteCheck(db, &monitor)
			dbops.SaveToDatabase(db, &newMonitor)
			if !notifications.SendEmailNotification(newMonitor) {
				log.Warn().Msg("Unable to send email notification")
				continue
			}
		} else {
			log.Info().Msg("Ports have not changed.")
		}

	}
}
