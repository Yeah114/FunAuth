# FunAuth 接口说明（精简版）

仅包含接口定义：路径、方法、请求头、请求体、响应、状态码。

公共前缀：`/api`
## /api/open（Open API，需要提供 Cookie）

所有端点均需请求头：`Authorization: cookie:<cookie>`。

### GET /api/open/g79/user_detail
- 返回：`{"success": true, "user": {...}}`

### GET /api/open/g79/rental_search?name=<kw>
- 返回：`{"success": true, "result": {...}}`

### GET /api/open/g79/lobby_room?id=<room_id>
- 返回：`{"success": true, "result": {...}}`

### GET /api/open/g79/rental_available?sort_type=0&order_type=0&offset=0
- 返回：`{"success": true, "result": {...}}`

### GET /api/open/g79/rental_details?id=<server_id>
- 返回：`{"success": true, "result": {...}}`

### GET /api/open/g79/user_settings
- 返回：`{"success": true, "result": {...}}`

### GET /api/open/g79/user_search?kw=<name_or_mail>&type=1&limit=10
- 返回：`{"success": true, "result": {...}}`

### GET /api/open/g79/download_info?item_id=<id>
- 返回：`{"success": true, "result": {...}}`


## GET /api/new

- 用途：
  - Bearer 模式：生成一次性 UUID（调试辅助）
  - Cookie 模式：若请求头含 `Authorization: cookie:<cookie>`，返回 `ok`，表示后续可直接用该 cookie 授权
- 响应：
  - Bearer 模式：`text/plain`（示例：`8f2c3d8a-2d6d-4b66-9a0e-3a1c5f2a0a31`）
  - Cookie 模式：`ok`

## POST /api/phoenix/login

- 请求头（二选一）：
  - `Authorization: Bearer <token>`
  - `Authorization: cookie:<cookie>`
- 请求体：
```json
{
  "login_token": "可选",
  "username": "可选",
  "password": "可选",
  "server_code": "房间号(19位)或租赁服名",
  "server_passcode": "入服口令",
  "client_public_key": "客户端公钥（可选）"
}
```
- 成功响应：
```json
{
  "success": true,
  "growth_level": 0,
  "token": "原样回显login_token（如传入）",
  "respond_to": "",
  "ip_address": "host:port",
  "chainInfo": "存在client_public_key时返回"
}
```
- 失败状态码：
  - 400：请求体不合法
  - 401：缺少/无效 Authorization
  - 503：上游 G79 客户端初始化失败

## POST /api/phoenix/transfer_check_num

- 请求体：
```json
{ "data": "[\"<mcpHex>\",\"<val>\",<unique_id>]" }
```
- 成功响应：
```json
{ "success": true, "value": "[\"<valm>\",\"<sign>\",false,[],\"\",\"\",3,\"<tmpsNum>\"]" }
```
- 失败响应：
```json
{ "success": false, "message": "bad data | bad mcp hex | pattern not found" }
```

## GET /api/phoenix/transfer_start_type

- 请求头（二选一）：
  - `Authorization: Bearer <token>`
  - `Authorization: cookie:<cookie>`
- 查询：
  - `content=<hex>`：G79 HTTP 加密十六进制
- 成功响应：
```json
{ "success": true, "data": "<hex>" }
```
- 失败响应：
```json
{ "success": false, "message": "authorization is required | authorization is invalid | bad hex | encrypt failed" }
```


