package lmarena_api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"lmarena2api/common"
	"lmarena2api/common/config"
	logger "lmarena2api/common/loggger"
	"lmarena2api/cycletls"
	"os/exec"
	"regexp"
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
	//logger.SysError(fmt.Sprintf("output: %v", output))
	return "", fmt.Errorf("未找到arena-auth-prod-v1 cookie")
}

func MakeStreamChatRequest(c *gin.Context, client cycletls.CycleTLS, jsonData []byte, cookie string, modelInfo common.ModelInfo) (<-chan cycletls.SSEResponse, error) {
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

func MakeSignUpRequest(token string) (string, error) {
	// 构建请求数据
	requestData := fmt.Sprintf(`{"turnstile_token":"%s"}`, token)

	// 构建curl命令
	cmd := exec.Command("curl",
		"https://beta.lmarena.ai/api/sign-up",
		"-H", "accept: */*",
		"-H", "accept-language: en-US,en;q=0.9",
		"-H", "cookie: cf_clearance=20VPnkBAX4ekFnvhZAC6JJMR27HmPUqbjqTrj_N5RYE-1747743176-1.2.1.1-QC67qVWttXKkZs3RaGtGRs4xgzylmNjJFa2tbq2ZPqDqsmVQUPom4lB.vkDCImUwjzCKQi93eteDYgPaU7ntpnrVW08e3rQVJlpu42HWambeMrLa7.YRjhddbx8o5Fjq6NJ2tqBI_kiiCbB_r5kAEe_mjmFbhc6w46QLdwcLdKl4GyMTGektNXpKabYPWhCIB40wZf31cWyzq6akRGSoCIRiHP8UvDHkaTJnGNDBbA4uGU8zZC6gYT.kw7D_MLhpBLjZgGhEnONQMmr0L.Ci_XGEltfj8HbJUtwuFqSjvXD3H7ZmBYWMICImqtjNN28jbFhllGBLElhxHDaSPPF3MB5YtFPUvGIerqQAbRxAzk_VKGCGnsYiBFm7zlcur5pi;",
		"-H", "content-type: text/plain;charset=UTF-8",
		"-H", "origin: https://beta.lmarena.ai",
		"-H", "priority: u=1, i",
		"-H", "referer: https://beta.lmarena.ai/",
		"-H", "sec-ch-ua: \"Google Chrome\";v=\"135\", \"Not-A.Brand\";v=\"8\", \"Chromium\";v=\"135\"",
		"-H", "sec-ch-ua-mobile: ?0",
		"-H", "sec-ch-ua-platform: \"macOS\"",
		"-H", "sec-fetch-dest: empty",
		"-H", "sec-fetch-mode: cors",
		"-H", "sec-fetch-site: same-origin",
		"-H", "user-agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36",
		"--data-raw", requestData,
		"-s") // 添加-s参数使curl静默输出，不显示进度信息

	// 执行命令并获取输出
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("执行curl命令失败: %v, 输出: %s", err, string(output))
	}

	// 提取JSON部分（假设响应是一个完整的JSON对象）
	jsonStr := string(output)

	// 检查是否为有效的JSON
	var jsonObj interface{}
	if err := json.Unmarshal([]byte(jsonStr), &jsonObj); err != nil {
		return "", fmt.Errorf("解析JSON失败: %v", err)
	}

	// 将JSON转换为Base64
	base64Str := base64.StdEncoding.EncodeToString([]byte(jsonStr))

	return base64Str, nil
}
