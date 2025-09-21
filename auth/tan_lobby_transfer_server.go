package auth

import (
	"fmt"

	"github.com/Yeah114/g79client"
)

func TransferServerList() ([]string, []string, error) {
	servers, err := g79client.GetGlobalTransferServers()
	if err != nil {
		return nil, nil, err
	}
	var raknetServers []string
	var websocketServers []string
	for _, server := range servers {
		for _, port := range server.Ports {
			raknetServers = append(raknetServers, fmt.Sprintf("%s:%d", server.IP, port))
		}
		websocketServers = append(websocketServers, fmt.Sprintf("%s:%d", server.IP, server.SignalWebPort.Int64()))
	}
	return raknetServers, websocketServers, nil
}
