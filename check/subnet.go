package check

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	dbops "github.com/jacewalker/ip-monitor/db"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

func ScanSubnet(db *gorm.DB, activeCheck dbops.Check) {
	fmt.Println("Subnet provided.", activeCheck.Addresses)
	ipAddressSlice := strings.Split(activeCheck.Addresses, ",")

	var wg sync.WaitGroup
	maxConcurrency := 30
	sem := make(chan struct{}, maxConcurrency)

	for _, ip := range ipAddressSlice {
		sem <- struct{}{}
		wg.Add(1)
		fmt.Printf("Working on %s...\n", ip)
		newactiveCheck := dbops.Check{}
		newactiveCheck.Address = ip
		newactiveCheck.Label = activeCheck.Label
		newactiveCheck.Email = activeCheck.Email

		go func() {
			if err := ScanPorts(&newactiveCheck); err != nil {
				log.Error().Msg("unable to scan ports or host is down")
			} else {
				dbops.SaveToDatabase(db, &newactiveCheck)
			}
			wg.Done()
			<-sem
		}()
	}
	wg.Wait()
}

func parseSubnet(c *gin.Context, subnet *string) (addresses string, err error) {
	ip, ipnet, err := net.ParseCIDR(*subnet)
	if err != nil {
		return "", errors.New("not a subnet")
	}

	log.Info().Msg("Subnet received")

	var ips string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); incIP(ip) {
		ips += ip.String() + ","
	}

	return ips, nil
}

// Increment IP
func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
