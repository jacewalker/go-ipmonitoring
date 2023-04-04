package dbops

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Init() *gorm.DB {
	dbName := "checks.db"

	if _, err := os.Stat(dbName); errors.Is(err, os.ErrNotExist) {
		log.Info().Msgf(dbName, "doesn't exist. Creating file...")
		if _, err := os.Create(dbName); err != nil {
			log.Error().Msgf("Unable to create files.db:", err)
		}
	} else {
		log.Info().Msgf(dbName, "already exists. Continuing using existing database.")
	}

	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		log.Error().Msgf("Unable to connect to the database:", err)
	} else {
		log.Info().Msgf("Database Initialised Successfully")
	}
	err = db.AutoMigrate(&Check{})
	if err != nil {
		log.Error().Msgf("Unable to complete database migration:", err)
	} else {
		log.Info().Msgf("Database migration completed.")
	}
	return db
}

func SaveToDatabase(db *gorm.DB, ch *Check) {
	log.Info().Msg("Saving message transaction to the database.")

	var portsSave string

	if len(ch.OpenPorts) == 0 {
		portsSave = "None"
	} else {
		portsSave = ch.OpenPorts
	}

	// if ch.Error != "" {

	// }

	db.Create(&Check{
		Address:   ch.Address,
		OpenPorts: portsSave,
		Label:     ch.Label,
		Email:     ch.Email,
	})
}

func GetAllFromDatabase(db *gorm.DB) ([]Check, error) {
	var monitors []Check
	results := db.Find(&monitors)
	if results.Error != nil {
		log.Warn().Msg(results.Error.Error())
		return nil, results.Error
	}
	return monitors, nil
}

func PortsToString(ch *Check, ports []int) {
	openPortsStr := []string{}
	for _, port := range ports {
		openPortsStr = append(openPortsStr, strconv.Itoa(port))
	}
	portsStr := strings.Join(openPortsStr, ",")

	ch.OpenPorts = portsStr
}

func StringToPorts(portsStr string) ([]int, error) {
	portsStr = strings.TrimSpace(portsStr)
	if portsStr == "" {
		return nil, nil
	}

	portStrings := strings.Split(portsStr, ",")
	ports := make([]int, len(portStrings))
	for i, portString := range portStrings {
		port, err := strconv.Atoi(portString)
		if err != nil {
			return nil, err
		}
		ports[i] = port
	}
	return ports, nil
}

func GetOpenPortDifferences(oldSlice []int, newSlice []int) (totalOpen []int, diffOpen []int) {
	totalOpenPorts := make([]int, 0)
	diffOpenPorts := make([]int, 0)

	// Check for ports that are in oldSlice but not newSlice
	for _, port := range oldSlice {
		found := false
		for _, newPort := range newSlice {
			if port == newPort {
				found = true
				totalOpenPorts = append(totalOpenPorts, port)
				break
			}
		}
		if !found {
			diffOpenPorts = append(diffOpenPorts, port)
		}
	}

	// Check for ports that are in newSlice but not oldSlice
	for _, port := range newSlice {
		found := false
		for _, oldPort := range oldSlice {
			if port == oldPort {
				found = true
				break
			}
		}
		if !found {
			diffOpenPorts = append(diffOpenPorts, port)
			totalOpenPorts = append(totalOpenPorts, port)
		}
	}

	return totalOpenPorts, diffOpenPorts
}

func LookupCheck(db *gorm.DB, addr string) (*Check, error) {
	var check Check
	result := db.Where("address = ?", addr).First(&check)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("no check found with address %s", addr)
		}
		return nil, result.Error
	}
	return &check, nil
}

func DeleteCheck(db *gorm.DB, ch *Check) bool {
	result := db.Delete(&ch)
	if result.Error != nil {
		log.Warn().Msg(result.Error.Error())
		return false
	}
	if result.RowsAffected == 0 {
		log.Warn().Msgf("No records deleted for check with ID %d", ch.ID)
		return false
	}
	log.Info().Msgf("Deleted %d records for check with ID %d", result.RowsAffected, ch.ID)
	return true
}
