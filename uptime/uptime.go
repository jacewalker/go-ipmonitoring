package uptime

import (
	"fmt"
	"time"

	dbops "github.com/jacewalker/ip-monitor/db"
	probing "github.com/prometheus-community/pro-bing"
	"github.com/rs/zerolog/log"
)

var db = dbops.Init()

func SendICMPRequest(ch *dbops.Check) {
	var status bool

	pinger, err := probing.NewPinger(ch.Address)
	if err != nil {
		log.Warn().Msg(err.Error())
	}

	pinger.SetPrivileged(true)
	pinger.Count = 2
	pinger.Timeout = 3 * time.Second
	fmt.Printf("Sending ICMP to %s\n", ch.Address)
	err = pinger.Run()
	if err != nil {
		log.Warn().Msgf(err.Error())
	}

	if pinger.Statistics().PacketsRecv == 0 {
		status = false
	} else {
		status = true
	}

	if ch.Online == status {
		fmt.Println("No changes, not updating the database.")
	} else {
		fmt.Println("Current ICMP status: ", status)
		fmt.Println("Previous ICMP Response: ", ch.Online)
		fmt.Println("Updating the database.")
		ch.Online = status
		if dbops.DeleteCheck(db, ch) {
			dbops.SaveToDatabase(db, ch)
		}
	}

}
