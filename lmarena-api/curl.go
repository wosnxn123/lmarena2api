package lmarena_api

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"lmarena2api/cycletls"
	"os/exec"
	"strings"
	"sync"
)

// 首先定义一个新的函数，使用curl命令执行SSE请求并返回与原函数相同类型的通道
// CurlSSE 使用curl命令执行SSE请求并返回响应通道
func CurlSSE(url string, options cycletls.Options) (<-chan cycletls.SSEResponse, error) {
	// 创建一个通道用于发送SSE响应
	sseChan := make(chan cycletls.SSEResponse)

	// 构建curl命令所需参数
	headers := options.Headers
	data := options.Body

	// 从headers中提取cookie
	cookie := ""
	if cookieVal, exists := headers["cookie"]; exists {
		cookie = cookieVal
		// 从headers中删除cookie，因为curl命令会使用-b参数
		delete(headers, "cookie")
	}

	// 创建上下文，可用于取消操作
	ctx, cancel := context.WithCancel(context.Background())

	// 在goroutine中执行curl命令
	go func() {
		defer close(sseChan) // 确保在函数结束时关闭通道
		defer cancel()       // 确保取消上下文

		// 构建基本命令
		args := []string{
			"-N",          // 禁用缓冲
			"--no-buffer", // 禁用输出缓冲
			url,
		}

		// 添加headers
		for key, value := range headers {
			args = append(args, "-H", fmt.Sprintf("%s: %s", key, value))
		}

		// 添加cookies
		if cookie != "" {
			args = append(args, "-b", cookie)
		}

		// 添加数据
		if data != "" {
			args = append(args, "--data-raw", data)
		}

		// 创建命令
		cmd := exec.CommandContext(ctx, "curl", args...)

		// 获取输出管道
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			sseChan <- cycletls.SSEResponse{
				Status: 500,
				Data:   fmt.Sprintf("创建stdout管道失败: %v", err),
				Done:   true,
			}
			return
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			sseChan <- cycletls.SSEResponse{
				Status: 500,
				Data:   fmt.Sprintf("创建stderr管道失败: %v", err),
				Done:   true,
			}
			return
		}

		// 启动命令
		if err := cmd.Start(); err != nil {
			sseChan <- cycletls.SSEResponse{
				Status: 500,
				Data:   fmt.Sprintf("启动命令失败: %v", err),
				Done:   true,
			}
			return
		}

		// 处理标准错误（仅记录，不发送到通道）
		go func() {
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				select {
				case <-ctx.Done():
					return
				default:
					// 可以记录错误，但不发送到通道
					// fmt.Fprintf(os.Stderr, "STDERR: %s\n", scanner.Text())
				}
			}
		}()

		// 处理标准输出 - 使用更低级别的读取方式以避免缓冲
		reader := bufio.NewReader(stdout)
		var line string
		//var err error

		for {
			select {
			case <-ctx.Done():
				return
			default:
				// 逐行读取，不使用Scanner以避免可能的缓冲
				line, err = reader.ReadString('\n')
				if err != nil {
					if err == io.EOF {
						// 正常结束
						break
					}
					// 其他错误
					sseChan <- cycletls.SSEResponse{
						Status: 500,
						Data:   fmt.Sprintf("读取stdout出错: %v", err),
						Done:   true,
					}
					return
				}

				// 去除行尾的换行符
				line = strings.TrimRight(line, "\r\n")
				if line == "" {
					continue
				}

				// 直接发送每一行数据
				sseChan <- cycletls.SSEResponse{
					Status: 200,
					Data:   line,
					Done:   false,
				}

				// 立即刷新，确保数据被发送
				// 这里不需要特别操作，因为通道发送是同步的
			}

			if err == io.EOF {
				break
			}
		}

		// 等待命令完成
		if err := cmd.Wait(); err != nil {
			// 只有在非EOF错误时才发送错误消息
			if err != io.EOF {
				sseChan <- cycletls.SSEResponse{
					Status: 500,
					Data:   fmt.Sprintf("命令执行错误: %v", err),
					Done:   true,
				}
				return
			}
		}

		// 发送完成标记
		sseChan <- cycletls.SSEResponse{
			Status: 200,
			Data:   "[DONE]",
			Done:   true,
		}
	}()

	return sseChan, nil
}

// ResponseCollector 用于收集响应内容
type ResponseCollector struct {
	mu      sync.Mutex
	content strings.Builder
	debug   bool // 调试模式标志
}

// AddContent 线程安全地添加内容
func (rc *ResponseCollector) AddContent(text string) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.content.WriteString(text)
}

// GetContent 线程安全地获取累积的内容
func (rc *ResponseCollector) GetContent() string {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	return rc.content.String()
}

// LogDebug 输出调试信息
func (rc *ResponseCollector) LogDebug(format string, args ...interface{}) {
	if rc.debug {
		rc.mu.Lock()
		defer rc.mu.Unlock()
		fmt.Printf("[DEBUG] "+format+"\n", args...)
	}
}

// StreamProcessor 处理流式响应
type StreamProcessor struct {
	collector *ResponseCollector
	callback  func(content string)
	mu        sync.Mutex
}

// NewStreamProcessor 创建新的流处理器
func NewStreamProcessor(callback func(content string), debug bool) *StreamProcessor {
	return &StreamProcessor{
		collector: &ResponseCollector{debug: debug},
		callback:  callback,
	}
}

// ProcessLine 处理单行输出
func (sp *StreamProcessor) ProcessLine(line string) {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	// 记录原始行用于调试
	sp.collector.LogDebug("收到原始行: %s", line)

	// 根据前缀处理不同类型的输出
	if strings.HasPrefix(line, "af:") {
		sp.collector.LogDebug("识别到消息头: %s", line)
		sp.collector.AddContent(line + "\n")
	} else if strings.HasPrefix(line, "a0:") {
		// 提取实际内容（去掉a0:前缀和引号）
		content := strings.Trim(line[3:], "\"")
		sp.collector.LogDebug("提取内容: %s", content)
		sp.collector.AddContent(line + "\n")
		if sp.callback != nil {
			sp.callback(content)
		}
	} else if strings.HasPrefix(line, "ae:") || strings.HasPrefix(line, "ad:") {
		sp.collector.LogDebug("识别到消息尾: %s", line)
		sp.collector.AddContent(line + "\n")
	} else {
		// 处理没有特定前缀的行
		sp.collector.LogDebug("未识别前缀的行: %s", line)
		sp.collector.AddContent(line + "\n")
	}
}

// GetCollectedContent 获取收集的内容
func (sp *StreamProcessor) GetCollectedContent() string {
	return sp.collector.GetContent()
}

// ExecuteCurlWithContext 使用上下文执行curl命令
func ExecuteCurlWithContext(ctx context.Context, url string, headers map[string]string, cookies string, data string, processor *StreamProcessor) error {
	// 构建基本命令
	args := []string{url}

	// 添加headers
	for key, value := range headers {
		args = append(args, "-H", fmt.Sprintf("%s: %s", key, value))
	}

	// 添加cookies
	if cookies != "" {
		args = append(args, "-b", cookies)
	}

	// 添加数据
	if data != "" {
		args = append(args, "--data-raw", data)
	}

	// 创建命令
	cmd := exec.CommandContext(ctx, "curl", args...)

	// 获取输出管道
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("创建stdout管道失败: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("创建stderr管道失败: %v", err)
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动命令失败: %v", err)
	}

	// 创建等待组
	var wg sync.WaitGroup
	wg.Add(2) // 一个用于stdout，一个用于stderr

	// 处理标准输出
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
				line := scanner.Text()
				processor.ProcessLine(line)
			}
		}
		if err := scanner.Err(); err != nil {
			processor.collector.LogDebug("扫描stdout出错: %v", err)
		}
	}()

	// 处理标准错误
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
				line := scanner.Text()
				processor.collector.LogDebug("STDERR: %s", line)
			}
		}
		if err := scanner.Err(); err != nil {
			processor.collector.LogDebug("扫描stderr出错: %v", err)
		}
	}()

	// 等待所有goroutine完成
	go func() {
		wg.Wait()
	}()

	// 等待命令完成
	return cmd.Wait()
}
