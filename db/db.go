package dbops

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Declaring env variables for the database connection.
var (
	DB_HOST string = os.Getenv("DB_HOST")
	DB_PORT string = os.Getenv("DB_PORT")
	DB_NAME string = os.Getenv("DB_NAME")
	DB_USER string = os.Getenv("DB_USER")
	DB_PASS string = os.Getenv("DB_PASS")
)

// Confirm the env variables can be read, otherwise exit.
func init() {
	godotenv.Load()
	var ok bool
	DB_HOST, ok = os.LookupEnv("DB_HOST")
	if !ok {
		log.Fatal().Msg("DB_HOST environment variable not set")
	}
	DB_PORT, ok = os.LookupEnv("DB_PORT")
	if !ok {
		log.Fatal().Msg("DB_PORT environment variable not set")
	}
	DB_NAME, ok = os.LookupEnv("DB_NAME")
	if !ok {
		log.Fatal().Msg("DB_NAME environment variable not set")
	}
	DB_USER, ok = os.LookupEnv("DB_USER")
	if !ok {
		log.Fatal().Msg("DB_USER environment variable not set")
	}
	DB_PASS, ok = os.LookupEnv("DB_PASS")
	if !ok {
		log.Fatal().Msg("DB_PASS environment variable not set")
	}
}

// Connect to the database and complete first migration.
func Init() *gorm.DB {
	db, err := createDatabaseIfNotExist()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	err = db.AutoMigrate(&Check{})
	if err != nil {
		log.Error().Msgf("Unable to complete database migration:", err)
	} else {
		log.Info().Msgf("Database migration completed.")
	}

	return db
}

func createDatabaseIfNotExist() (*gorm.DB, error) {
	// Connect to the server
	dsn := fmt.Sprintf("host=%s user=%s password=%s port=%s sslmode=disable TimeZone=Australia/Melbourne", DB_HOST, DB_USER, DB_PASS, DB_PORT)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Warn().Msg(db.Error.Error())
		return nil, errors.New("cannot connect to database server")
	}

	// Check if the database exists
	var count int64
	db.Raw("SELECT COUNT(*) FROM pg_database WHERE datname = ?", DB_NAME).Scan(&count)

	// If the database does not exist:
	if count == 0 {
		// Create the database
		result := db.Exec(fmt.Sprintf("CREATE DATABASE %s", DB_NAME))
		if result.Error != nil {
			log.Fatal().Msg(result.Error.Error())
		}
	}
	dsn2 := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Australia/Melbourne", DB_HOST, DB_USER, DB_PASS, DB_NAME, DB_PORT)
	db2, err := gorm.Open(postgres.Open(dsn2), &gorm.Config{})
	if err != nil {
		log.Warn().Msg(db.Error.Error())
		return nil, errors.New("cannot connect to database server")
	}
	return db2, nil
}

func SaveToDatabase(db *gorm.DB, ch *Check) {
	log.Info().Msg("Saving message transaction to the database.")

	var portsSave string

	if len(ch.OpenPorts) == 0 {
		if ch.PortScanCompleted {
			portsSave = "None"
		} else {
			portsSave = "Pending scan..."
		}

	} else {
		portsSave = ch.OpenPorts
	}

	hostnameSave := ""
	if len(ch.Hostname) == 0 {
		hostnameSave = ""
	} else {
		hostnameSave = ch.Hostname
	}

	db.Save(&Check{
		Address:           ch.Address,
		OpenPorts:         portsSave,
		Label:             ch.Label,
		Email:             ch.Email,
		Online:            ch.Online,
		Hostname:          hostnameSave,
		PortScanCompleted: ch.PortScanCompleted,
		OfflineCount:      ch.OfflineCount,
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
	result := db.Where("address = ?", ch.Address).Delete(&ch)
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
