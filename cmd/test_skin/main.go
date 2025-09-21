package main

import (
	"fmt"
	"log"

	"github.com/Yeah114/g79client"
)

func main() {
	client, err := g79client.NewClient()
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}

	if err := client.AuthenticateWithCookie(`{"sauth_json":"{\"gameid\":\"x19\",\"login_channel\":\"netease\",\"app_channel\":\"netease\",\"platform\":\"pc\",\"sdkuid\":\"aibgqyxlsghz3lqo\",\"sessionid\":\"1-eyJzIjogIjBmM3F3aGF2NDFrdHg5b2ZmdGhwbnQxcW15cW1udnEzIiwgIm9kaSI6ICJhbWF3cXlxYWF3b29iNWcyLWQiLCAic2kiOiAiNTkzNjZiM2NhMDJkMzMxMGQ1ZmI1ZWExYjUwMGMyNGI5YjkwNzI2MiIsICJ1IjogImFpYmdxeXhsc2doejNscW8iLCAidCI6IDIsICJnX2kiOiAiYWVjZnJ4b2R5cWFhYWFqcCJ9\",\"sdk_version\":\"3.9.0\",\"udid\":\"0j7iwgxzd0y4kaga1z1sxj7tnqn8mxcs\",\"deviceid\":\"amawqyqaawoob5g2-d\",\"aim_info\":\"{\\\"aim\\\":\\\"127.0.0.1\\\",\\\"country\\\":\\\"CN\\\",\\\"tz\\\":\\\"+0800\\\",\\\"tzid\\\":\\\"\\\"}\",\"client_login_sn\":\"4318DF932C410EBEE8F2164D0F5ECED6\",\"gas_token\":\"\",\"source_platform\":\"pc\",\"ip\":\"127.0.0.1\"}"}`); err != nil {
		log.Fatalf("认证失败: %v", err)
	}

	userSettingList, err := client.GetUserSettingList()
	if err != nil {
		log.Fatalf("获取用户设置列表失败: %v", err)
	}
	fmt.Println(userSettingList)
}