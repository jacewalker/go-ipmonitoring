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
	Error             string
	Email             string
}
