package check

import (
	"fmt"

	dbops "github.com/jacewalker/ip-monitor/db"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

func ScanIP(db *gorm.DB, activeCheck dbops.Check) {
	fmt.Println("Scanning IPv4:", activeCheck.Address)
	if err := ScanPorts(&activeCheck); err != nil {
		log.Error().Msg("unable to scan ports or host is down")
	} else {
		dbops.DeleteCheck(db, &activeCheck)
		activeCheck.PortScanCompleted = true
		dbops.SaveToDatabase(db, &activeCheck)
	}
}
