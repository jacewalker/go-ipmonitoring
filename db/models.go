package dbops

import (
	"gorm.io/gorm"
)

type Check struct {
	gorm.Model
	Address           string
	Addresses         string
	Hostname          string
	OpenPorts         string
	PortScanCompleted bool
	Label             string
	ScanType          string
	Online            bool
	OfflineCount      int
	Error             string
	// NotificationMethod string
	Email string
	// Pushover           string
}
