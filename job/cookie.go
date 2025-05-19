package job

import (
	"fmt"
	"lmarena2api/common/config"
	logger "lmarena2api/common/loggger"
	"lmarena2api/cycletls"
	kilo_api "lmarena2api/lmarena-api"
	"time"
)

func UpdateCookieTokenTask() {
	client := cycletls.Init()
	defer safeClose(client)
	for {
		logger.SysLog("lmarena2api Scheduled UpdateCookieTokenTask Task Job Start!")

		for _, cookie := range config.NewCookieManager().Cookies {
			// 使用GetAuthToken获取新的token
			newToken, err := kilo_api.GetAuthToken(nil, cookie) // 传入nil因为这里不需要gin上下文
			if err != nil {
				logger.SysError(fmt.Sprintf("GetAuthToken err for cookie %s: %v", cookie, err))
				//continue
				newToken = cookie
			}

			// 更新LATokenMap
			config.LATokenMap[cookie] = config.LATokenInfo{
				NewCookie: newToken,
				// 其他字段如果需要的话也可以设置
			}

			logger.SysLog(fmt.Sprintf("Updated token for cookie %s", cookie))
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
