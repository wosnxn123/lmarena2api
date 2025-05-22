package job

import (
	logger "lmarena2api/common/loggger"
	"lmarena2api/cycletls"
	"time"
)

func UpdateCookieTokenTask() {
	client := cycletls.Init()
	defer safeClose(client)
	for {
		logger.SysLog("lmarena2api Scheduled UpdateCookieTokenTask Task Job Start!")

		//for _, cookie := range config.NewCookieManager().Cookies {
		//	cookieBase64 := strings.Replace(cookie, "base64-", "", 1)
		//
		//	jsonData, err := base64.StdEncoding.DecodeString(cookieBase64)
		//	if err != nil {
		//		logger.SysError(fmt.Sprintf("Error decoding base64:", err))
		//		continue
		//	}
		//
		//	var tokenResponse map[string]interface{}
		//	json.Unmarshal(jsonData, &tokenResponse)
		//
		//	// Get expires_at timestamp
		//	expiresAt := tokenResponse["expires_at"].(float64)
		//
		//	// Get current time
		//	currentTime := time.Now().Unix()
		//
		//	// Calculate time until expiration (in seconds)
		//	timeUntilExpiration := expiresAt - float64(currentTime)
		//
		//	// Check if expiration is within 10 minutes (600 seconds)
		//	if timeUntilExpiration <= 600 {
		//
		//		//clearance, err := yescaptcha.GetCFClearance(yescaptcha.WebsiteURL, yescaptcha.YescaptchaProxyUrl)
		//		//if err != nil {
		//		//	fmt.Println("获取cf_clearance失败:", err)
		//		//	return
		//		//}
		//		//
		//		clearance := "cf_clearance=kSV2AOzEw5fjn7DIt5ZR7kq5T2XzbDRYX1rvpGOE84U-1747887554-1.2.1.1-GUZEW_12KBG22rCAonZJrE1z2kNyZh7DHoP_3hX62TlELu3yBPdRZKWkr0IZPNZRfWKPi19WIuSu8o3IOILaJ9des.l8SFUqSZL_YvxXQw4HIjfGnyuTjPuaa8fYtytrnmF2hMsEQvSq2XlbQk_9BMCU2ET284LsP0hDZIc2iaf1fbmwoxlS_8vonwdKezB38XrGuQfwO48V1xeAwzhOBH.hfT0TOJG1fvrYqXNl8zTuWV_ntde6lvqSRIqgMqXAFTdqWOSTy2DhJvYUIb9G4lt7fAb7zKQtG6hTE8ZhhXfIAqi4uO3CmvWXWjJ2wyoxQnX4hfGbvJ02Z0EjAPCXg76GwdhKdlVbyStEbvinOnQ"
		//		task, err := yescaptcha.CreateTask()
		//		if err != nil {
		//			fmt.Println("创建任务失败:", err)
		//			return
		//		}
		//		token, err := yescaptcha.GetTaskResult(task)
		//		if err != nil {
		//			fmt.Println("获取结果失败:", err)
		//			return
		//		}
		//		//clearance := "_ga=GA1.1.1934472180.1747794847;cf_clearance=givAzRfZL_LbEDfxiwFC6bDUTgnH.wpHWjqId3I7tEQ-1747794847-1.2.1.1-GlI2RAx24iClZckPa_dknQ4MA3OLSAREeytnoXJcV.IrvwShYDTnVjh2LrBNAzkMFt67WHOeVDU0KSdHs.HpcOJxOIqOivSVQGrgCYfAPjj3_ts37mO5NpXnh5qi1wPuizjuhLjg_gCF3AZuysdtrQRCpkKEdQi_5.HmbREliyazhjyoHxyHxrhGlqjsfchNvmqWdtwKw.aGDPHjgNrIbw22wn0ch.9n67zQEYea53hcDE7ZWisibrhZkx4hqfJKfuSS9te71zI9aYihkmP27JmWU3FC70cwkFFkVXlBbA89R35pcOcI8gqnu7PfO17R9EM8xcsxHsMXXD.3NdnvX1fTXTqOP5HGnZS48JeiYN4;vercel-experiment-uuid=ceq9ARHZgDRUHQ1Xa9Nf1;ph_phc_LG7IJbVJqBsk584rbcKca0D5lV2vHguiijDrVji7yDM_posthog=%7B%22distinct_id%22%3A%220196f0b0-e401-7e47-90d2-3c4a048fe312%22%2C%22%24sesid%22%3A%5B1747794874396%2C%220196f0b0-e3fe-79b9-a510-7d8da7db089c%22%2C1747794846718%5D%2C%22%24initial_person_info%22%3A%7B%22r%22%3A%22%24direct%22%2C%22u%22%3A%22https%3A%2F%2Fbeta.lmarena.ai%2F%22%7D%7D;_ga_72FK1TMV06=GS2.1.s1747794847$o1$g1$t1747794874$j0$l0$h0;cf_chl_rc_m=1"
		//		resp, err := lmarena_api.MakeSignUpRequest(token, clearance)
		//		if err != nil {
		//			fmt.Println("API请求失败:", err)
		//			return
		//		}
		//
		//		newCookie := "base64-" + resp
		//
		//		config.LATokenMap = map[string]config.LATokenInfo{}
		//		config.LATokenMap[cookie] = config.LATokenInfo{
		//			NewCookie: newCookie,
		//			//		// 其他字段如果需要的话也可以设置
		//		}
		//		logger.SysLog(fmt.Sprintf("Cookie is no valid,need to update %s -> %s", cookie, newCookie))
		//	} else {
		//		config.LATokenMap[cookie] = config.LATokenInfo{
		//			NewCookie: cookie,
		//			//		// 其他字段如果需要的话也可以设置
		//		}
		//		logger.SysLog(fmt.Sprintf("Cookie is still valid, no need to update %s", cookie))
		//	}
		//
		//	//// 使用GetAuthToken获取新的token
		//	//newToken, err := kilo_api.GetAuthToken(nil, cookie) // 传入nil因为这里不需要gin上下文
		//	//if err != nil {
		//	//	logger.SysError(fmt.Sprintf("GetAuthToken err for cookie %s: %v", cookie, err))
		//	//	// 更新LATokenMap
		//	//	config.LATokenMap[cookie] = config.LATokenInfo{
		//	//		NewCookie: cookie,
		//	//		// 其他字段如果需要的话也可以设置
		//	//	}
		//	//} else {
		//	//	logger.SysLog(fmt.Sprintf("Updated token for cookie %s", cookie))
		//	//	// 更新LATokenMap
		//	//	config.LATokenMap[cookie] = config.LATokenInfo{
		//	//		NewCookie: newToken,
		//	//		// 其他字段如果需要的话也可以设置
		//	//	}
		//	//}
		//
		//}

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
