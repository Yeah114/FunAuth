package handlers

import (
	"fmt"
	"net/http"

	"github.com/Yeah114/FunAuth/auth"
	"github.com/gin-gonic/gin"
)

func RegisterPhoenixTanLobbyTransferServerRoute(api *gin.RouterGroup) {
	api.POST("/phoenix/tan_lobby_transfer_server", func(c *gin.Context) {
		raknetServers, websocketServers, err := auth.TransferServerList()
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "error_info": fmt.Sprintf("TransferServerList: %v", err)})
			return
		}
		c.JSON(http.StatusOK, TanLobbyTransferServersResponse{
			Success:          true,
			ErrorInfo:        "",
			RaknetServers:    raknetServers,
			WebsocketServers: websocketServers,
		})
	})
}
