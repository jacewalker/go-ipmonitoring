package dbops

import (
	"gorm.io/gorm"
)

type Check struct {
	gorm.Model
	Address   string
	OpenPorts string
}
