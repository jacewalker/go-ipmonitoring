package dbops

import (
	"gorm.io/gorm"
)

type Check struct {
	gorm.Model
	Address   string
	Addresses string
	OpenPorts string
	Label     string
	ScanType  string
	Error     string
	Email     string
}
