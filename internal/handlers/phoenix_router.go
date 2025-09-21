package handlers

import "github.com/gin-gonic/gin"

// 汇总注册，调用各自文件中的注册函数
func RegisterPhoenixRoutes(api *gin.RouterGroup) {
	RegisterPhoenixLoginRoute(api)
	RegisterPhoenixTransferCheckNumRoute(api)
	RegisterPhoenixTransferStartTypeRoute(api)

	RegisterPhoenixTanLobbyLoginRoute(api)
	RegisterPhoenixTanLobbyCreateRoute(api)
	RegisterPhoenixTanLobbyTransferServerRoute(api)
}
