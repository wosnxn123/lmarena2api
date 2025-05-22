package yescaptcha

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"lmarena2api/common/env"
	"net/http"
	"time"
)

var YescaptchaClientKey = env.String("YESCAPTCHA_CLIENT_KEY", "")
var YescaptchaProxyUrl = env.String("YESCAPTCHA_PROXY_URL", "")

const (
	// YesCaptcha API配置
	createTaskURL = "https://api.yescaptcha.com/createTask"
	getResultURL  = "https://api.yescaptcha.com/getTaskResult"

	// Turnstile配置
	WebsiteURL = "https://beta.lmarena.ai/"
	websiteKey = "0x4AAAAAAA65vWDmG-O_lPtT"
)

// 创建任务请求结构
type CreateTaskRequest struct {
	ClientKey string `json:"clientKey"`
	Task      Task   `json:"task"`
}

type Task struct {
	Type       string `json:"type"`
	WebsiteURL string `json:"websiteURL"`
	WebsiteKey string `json:"websiteKey"`
	UserAgent  string `json:"userAgent,omitempty"`
	WaitLoad   bool   `json:"waitLoad,omitempty"`
	Proxy      string `json:"proxy,omitempty"`
}

// 创建任务响应结构
type CreateTaskResponse struct {
	ErrorID          int    `json:"errorId"`
	ErrorCode        string `json:"errorCode"`
	ErrorDescription string `json:"errorDescription"`
	TaskID           string `json:"taskId"`
	Status           string `json:"status"`
	Error            string `json:"error,omitempty"`
}

// 获取结果请求结构
type GetResultRequest struct {
	ClientKey string `json:"clientKey"`
	TaskID    string `json:"taskId"`
}

// 获取结果响应结构
type GetResultResponse struct {
	ErrorID          int       `json:"errorId"`
	ErrorCode        string    `json:"errorCode"`
	ErrorDescription string    `json:"errorDescription"`
	Status           string    `json:"status"`
	Solution         *Solution `json:"solution"`
	Error            string    `json:"error,omitempty"`
}

type Solution struct {
	Token              string            `json:"token"`
	UserAgent          string            `json:"userAgent"`
	RequestHeaders     map[string]string `json:"request_headers,omitempty"`
	Headers            map[string]string `json:"headers,omitempty"`
	Cookies            map[string]string `json:"cookies,omitempty"`
	URL                string            `json:"url,omitempty"`
	Status             int               `json:"status,omitempty"`
	GRecaptchaResponse string            `json:"gRecaptchaResponse,omitempty"`
}

func CreateTask() (string, error) {
	// 构建请求
	reqData := CreateTaskRequest{
		ClientKey: YescaptchaClientKey,
		Task: Task{
			Type:       "TurnstileTaskProxylessM1",
			WebsiteURL: WebsiteURL,
			WebsiteKey: websiteKey,
		},
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return "", err
	}

	// 发送请求
	resp, err := http.Post(createTaskURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 解析响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result CreateTaskResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	// 检查错误
	if result.ErrorID != 0 {
		return "", fmt.Errorf("API错误: %s - %s", result.ErrorCode, result.ErrorDescription)
	}

	return result.TaskID, nil
}

func GetTaskResult(taskID string) (string, error) {
	// 构建请求
	reqData := GetResultRequest{
		ClientKey: YescaptchaClientKey,
		TaskID:    taskID,
	}

	// 最多尝试20次，每次间隔3秒
	for i := 0; i < 20; i++ {
		jsonData, err := json.Marshal(reqData)
		if err != nil {
			return "", err
		}

		// 发送请求
		resp, err := http.Post(getResultURL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			return "", err
		}

		// 解析响应
		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return "", err
		}

		var result GetResultResponse
		if err := json.Unmarshal(body, &result); err != nil {
			return "", err
		}

		// 检查错误
		if result.ErrorID != 0 {
			return "", fmt.Errorf("API错误: %s - %s", result.ErrorCode, result.ErrorDescription)
		}

		// 检查状态
		if result.Status == "ready" {
			return result.Solution.Token, nil
		}

		//logger.SysLog("[YESCAPTCHA]任务正在处理中，3秒后重试...")
		time.Sleep(3 * time.Second)
	}

	return "", fmt.Errorf("等待超时，未能获取结果")
}

// GetCFClearance 获取Cloudflare的cf_clearance值
func GetCFClearance(targetURL string, proxy string) (string, error) {
	if YescaptchaClientKey == "" {
		return "", fmt.Errorf("YesCaptcha客户端密钥未设置")
	}

	// 创建CloudFlare绕过任务
	createTaskReq := CreateTaskRequest{
		ClientKey: YescaptchaClientKey,
		Task: Task{
			Type:       "CloudFlareTaskS2",
			UserAgent:  "",
			WaitLoad:   true,
			WebsiteURL: targetURL,
			Proxy:      proxy,
		},
	}

	jsonData, err := json.Marshal(createTaskReq)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(createTaskURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var createResult CreateTaskResponse
	if err := json.Unmarshal(body, &createResult); err != nil {
		return "", err
	}

	if createResult.Status == "error" || createResult.TaskID == "" {
		return "", fmt.Errorf("创建任务失败: %s", createResult.Error)
	}

	// 获取任务结果
	taskID := createResult.TaskID

	// 最多尝试30次，每次间隔3秒
	for i := 0; i < 30; i++ {
		time.Sleep(3 * time.Second)

		getTaskReq := GetResultRequest{
			ClientKey: YescaptchaClientKey,
			TaskID:    taskID,
		}

		jsonData, err := json.Marshal(getTaskReq)
		if err != nil {
			return "", err
		}

		resp, err := http.Post(getResultURL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			return "", err
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return "", err
		}

		var result GetResultResponse
		if err := json.Unmarshal(body, &result); err != nil {
			return "", err
		}

		if result.Status == "processing" {
			continue
		} else if result.Status == "ready" {
			fmt.Println(result.Solution)
			// 从cookies中提取cf_clearance
			if clearance, ok := result.Solution.Cookies["cf_clearance"]; ok {
				return clearance, nil
			}
			return "", fmt.Errorf("未找到cf_clearance值")
		} else {
			return "", fmt.Errorf("任务失败: %s", result.Error)
		}
	}

	return "", fmt.Errorf("等待超时，未能获取结果")
}
