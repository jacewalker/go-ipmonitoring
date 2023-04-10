package dbops

import (
	"gorm.io/gorm"
)

// type MonitoringType struct {
// 	Uptime        bool
// 	OpenPort      bool
// 	Vulnerability bool
// }

type Check struct {
	gorm.Model
	Address   string
	Addresses string
	OpenPorts string
	Label     string
	ScanType  string
	// Monitoring MonitoringType
	Error string
	Email string
}
