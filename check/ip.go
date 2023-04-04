package check

import (
	"fmt"

	dbops "github.com/jacewalker/ip-monitor/db"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

func ScanIP(db *gorm.DB, activeCheck dbops.Check) {
	fmt.Println("IP has been provided!!!")
	go func() {
		if err := ScanPorts(&activeCheck); err != nil {
			log.Error().Msg("unable to scan ports or host is down")
		} else {
			dbops.SaveToDatabase(db, &activeCheck)
		}
	}()
}
