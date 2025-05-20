package job

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"lmarena2api/common/config"
	logger "lmarena2api/common/loggger"
	"lmarena2api/cycletls"
	lmarena_api "lmarena2api/lmarena-api"
	"lmarena2api/yescaptcha"
	"strings"
	"time"
)

func UpdateCookieTokenTask() {
	client := cycletls.Init()
	defer safeClose(client)
	for {
		logger.SysLog("lmarena2api Scheduled UpdateCookieTokenTask Task Job Start!")

		for _, cookie := range config.NewCookieManager().Cookies {
			cookieBase64 := strings.Replace(cookie, "base64-", "", 1)

			jsonData, err := base64.StdEncoding.DecodeString(cookieBase64)
			if err != nil {
				logger.SysError(fmt.Sprintf("Error decoding base64:", err))
				continue
			}

			var tokenResponse map[string]interface{}
			json.Unmarshal([]byte(jsonData), &tokenResponse)

			// Get expires_at timestamp
			expiresAt := tokenResponse["expires_at"].(float64)

			// Get current time
			currentTime := time.Now().Unix()

			// Calculate time until expiration (in seconds)
			timeUntilExpiration := expiresAt - float64(currentTime)

			// Check if expiration is within 10 minutes (600 seconds)
			if timeUntilExpiration <= 600 {
				task, err := yescaptcha.CreateTask()
				if err != nil {
					fmt.Println("创建任务失败:", err)
					return
				}
				token, err := yescaptcha.GetTaskResult(task)
				if err != nil {
					fmt.Println("获取结果失败:", err)
					return
				}
				resp, err := lmarena_api.MakeSignUpRequest(token)
				if err != nil {
					fmt.Println("API请求失败:", err)
					return
				}

				newCookie := "base64-" + resp

				config.LATokenMap = map[string]config.LATokenInfo{}
				config.LATokenMap[cookie] = config.LATokenInfo{
					NewCookie: newCookie,
					//		// 其他字段如果需要的话也可以设置
				}
				logger.SysLog(fmt.Sprintf("Cookie is no valid,need to update %s -> %s", cookie, newCookie))
			} else {
				config.LATokenMap[cookie] = config.LATokenInfo{
					NewCookie: cookie,
					//		// 其他字段如果需要的话也可以设置
				}
				logger.SysLog(fmt.Sprintf("Cookie is still valid, no need to update %s", cookie))
			}

			//// 使用GetAuthToken获取新的token
			//newToken, err := kilo_api.GetAuthToken(nil, cookie) // 传入nil因为这里不需要gin上下文
			//if err != nil {
			//	logger.SysError(fmt.Sprintf("GetAuthToken err for cookie %s: %v", cookie, err))
			//	// 更新LATokenMap
			//	config.LATokenMap[cookie] = config.LATokenInfo{
			//		NewCookie: cookie,
			//		// 其他字段如果需要的话也可以设置
			//	}
			//} else {
			//	logger.SysLog(fmt.Sprintf("Updated token for cookie %s", cookie))
			//	// 更新LATokenMap
			//	config.LATokenMap[cookie] = config.LATokenInfo{
			//		NewCookie: newToken,
			//		// 其他字段如果需要的话也可以设置
			//	}
			//}

		}

		logger.SysLog("lmarena2api Scheduled UpdateCookieTokenTask Task Job End!")

		now := time.Now()
		remainder := now.Minute() % 10
		minutesToAdd := 10 - remainder
		if remainder == 0 {
			minutesToAdd = 10
		}
		next := now.Add(time.Duration(minutesToAdd) * time.Minute)
		next = time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), next.Minute(), 0, 0, next.Location())
		time.Sleep(next.Sub(now))
	}
}
func safeClose(client cycletls.CycleTLS) {
	if client.ReqChan != nil {
		close(client.ReqChan)
	}
	if client.RespChan != nil {
		close(client.RespChan)
	}
}
