package lmarena_api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"lmarena2api/common"
	"lmarena2api/common/config"
	logger "lmarena2api/common/loggger"
	"lmarena2api/cycletls"
	"os/exec"
	"regexp"
	"strings"
)

const (
// baseURL            = "https://kilocode.ai"
// chatEndpoint       = baseURL + "/api/claude/v1/messages"
// openRouterEndpoint = baseURL + "/api/openrouter/chat/completions"
)

func GetAuthToken(c *gin.Context, cookie string) (string, error) {
	cmd := exec.Command("curl", "-i", "https://beta.lmarena.ai/api/refresh",
		"-X", "POST",
		"-H", "accept: */*",
		"-H", "accept-language: zh-CN,zh;q=0.9",
		"-H", "content-length: 0",
		"-H", "content-type: application/json",
		"-b", "arena-auth-prod-v1="+cookie,
		"-H", "origin: https://beta.lmarena.ai",
		"-H", "priority: u=1, i",
		"-H", "referer: https://beta.lmarena.ai/c/81abf456-d419-456f-bd11-0fb8093fd7c9",
		"-H", "sec-ch-ua: \"Chromium\";v=\"136\", \"Google Chrome\";v=\"136\", \"Not.A/Brand\";v=\"99\"",
		"-H", "sec-ch-ua-full-version: 136.0.1613.16",
		"-H", "sec-ch-ua-full-version-list: \"Chromium\";v=\"136.0.1613.16\", \"Google Chrome\";v=\"136.0.1613.16\", \"Not.A/Brand\";v=\"99.0.0.0\"",
		"-H", "sec-ch-ua-mobile: ?0",
		"-H", "sec-ch-ua-platform: \"macOS\"",
		"-H", "sec-fetch-dest: empty",
		"-H", "sec-fetch-mode: cors",
		"-H", "sec-fetch-site: same-origin",
		"-H", "user-agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("执行curl命令失败: %v", err)
	}

	response := string(output)
	re := regexp.MustCompile(`set-cookie: arena-auth-prod-v1=([^;]+)`)
	matches := re.FindStringSubmatch(response)

	if len(matches) > 1 {
		return matches[1], nil
	}
	logger.SysLog(fmt.Sprintf("output: %v", output))
	return "", fmt.Errorf("未找到arena-auth-prod-v1 cookie")
}

func MakeStreamChatRequest(c *gin.Context, client cycletls.CycleTLS, jsonData []byte, cookie string, modelInfo common.ModelInfo) (<-chan cycletls.SSEResponse, error) {
	split := strings.Split(cookie, "=")
	if len(split) >= 2 {
		cookie = split[0]
	}

	tokenInfo, ok := config.LATokenMap[cookie]
	if !ok {
		return nil, fmt.Errorf("cookie not found in ASTokenMap")
	}

	headers := map[string]string{
		"accept":          "*/*",
		"accept-language": "zh-CN,zh;q=0.9",
		"content-type":    "application/json",
		"origin":          "https://beta.lmarena.ai",
		"referer":         "https://beta.lmarena.ai/",
		"supabase-jwt":    tokenInfo.NewCookie,
		"user-agent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36",
	}

	options := cycletls.Options{
		Timeout: 10 * 60 * 60,
		Proxy:   config.ProxyUrl, // 在每个请求中设置代理
		Body:    string(jsonData),
		Method:  "POST",
		Headers: headers,
	}

	logger.Debug(c.Request.Context(), fmt.Sprintf("cookie: %v", cookie))

	sseChan, err := client.DoSSE("https://arena-api-stable.vercel.app/evaluation", options, "POST")
	if err != nil {
		logger.Errorf(c, "Failed to make stream request: %v", err)
		return nil, fmt.Errorf("Failed to make stream request: %v", err)
	}
	return sseChan, nil
}
