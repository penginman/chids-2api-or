package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"orchids-api/internal/clerk"
	"orchids-api/internal/register"
)

// AccountPushData 推送到远程 API 的账号数据
type AccountPushData struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	ClientCookie string `json:"client_cookie"`
	ClientUat    string `json:"client_uat"`
	SessionID    string `json:"session_id"`
	UserID       string `json:"user_id"`
	ProjectID    string `json:"project_id"`
	AgentMode    string `json:"agent_mode"`
	Weight       int    `json:"weight"`
	Enabled      bool   `json:"enabled"`
}

func main() {
	// 命令行参数
	count := flag.Int("count", 5, "注册数量")
	workers := flag.Int("workers", 2, "并发线程数")
	headless := flag.Bool("headless", false, "是否使用无头模式")
	pushURL := flag.String("push-url", "", "推送 API 地址 (可从环境变量 PUSH_API_URL 读取)")
	pushUser := flag.String("push-user", "", "推送 API 用户名 (可从环境变量 PUSH_API_USER 读取)")
	pushPass := flag.String("push-pass", "", "推送 API 密码 (可从环境变量 PUSH_API_PASS 读取)")

	flag.Parse()

	// 从环境变量读取配置(优先级高)
	if *pushURL == "" {
		*pushURL = os.Getenv("PUSH_API_URL")
	}
	if *pushUser == "" {
		*pushUser = os.Getenv("PUSH_API_USER")
	}
	if *pushPass == "" {
		*pushPass = os.Getenv("PUSH_API_PASS")
	}

	log.Printf("========================================")
	log.Printf("自动注册与推送工具启动")
	log.Printf("========================================")
	log.Printf("配置信息:")
	log.Printf("  注册数量: %d", *count)
	log.Printf("  并发线程: %d", *workers)
	log.Printf("  无头模式: %v", *headless)
	log.Printf("  推送地址: %s", maskURL(*pushURL))
	log.Printf("  推送用户: %s", maskString(*pushUser))
	log.Printf("========================================")

	// 验证必需配置
	if *pushURL == "" {
		log.Fatal("错误: 必须提供推送 API 地址 (--push-url 或环境变量 PUSH_API_URL)")
	}
	if *pushUser == "" || *pushPass == "" {
		log.Fatal("错误: 必须提供推送 API 认证信息 (--push-user/--push-pass 或环境变量)")
	}

	// 创建注册服务
	registerService := register.New()

	// 执行批量注册
	log.Printf("\n开始批量注册...")
	batchResult := registerService.BatchRegister(*count, *workers, *headless)

	if batchResult == nil {
		log.Fatal("批量注册失败: 返回结果为空")
	}

	log.Printf("\n========================================")
	log.Printf("批量注册完成!")
	log.Printf("========================================")
	log.Printf("总数: %d", batchResult.Total)
	log.Printf("成功: %d", batchResult.Success)
	log.Printf("失败: %d", batchResult.Failed)
	log.Printf("耗时: %s", batchResult.Duration)
	log.Printf("========================================\n")

	// 推送成功的账号到远程 API
	if batchResult.Success > 0 {
		log.Printf("开始推送注册账号到远程 API...")
		pushResults := pushAccounts(batchResult.Results, *pushURL, *pushUser, *pushPass)

		log.Printf("\n========================================")
		log.Printf("推送完成!")
		log.Printf("========================================")
		log.Printf("推送总数: %d", pushResults.Total)
		log.Printf("推送成功: %d", pushResults.Success)
		log.Printf("推送失败: %d", pushResults.Failed)
		log.Printf("========================================\n")

		// 输出详细结果
		if len(pushResults.Details) > 0 {
			log.Printf("详细结果:")
			for i, detail := range pushResults.Details {
				status := "✓"
				if !detail.Success {
					status = "✗"
				}
				log.Printf("  [%d] %s %s - %s", i+1, status, detail.Email, detail.Message)
			}
		}

		// 如果有推送失败的,退出码为1
		if pushResults.Failed > 0 {
			os.Exit(1)
		}
	} else {
		log.Printf("没有成功注册的账号,跳过推送")
		os.Exit(1)
	}

	log.Printf("\n所有任务完成!")
}

// PushResults 推送结果统计
type PushResults struct {
	Total   int
	Success int
	Failed  int
	Details []PushDetail
}

// PushDetail 单个推送详情
type PushDetail struct {
	Email   string
	Success bool
	Message string
}

// pushAccounts 推送账号到远程 API
func pushAccounts(results []*register.RegisterResult, apiURL, username, password string) *PushResults {
	pushResults := &PushResults{
		Details: make([]PushDetail, 0),
	}

	for i, result := range results {
		if result == nil {
			log.Printf("[推送 %d/%d] 跳过空结果", i+1, len(results))
			continue
		}

		pushResults.Total++

		// 只推送成功的账号
		if !result.Success || result.ClientCookie == "" {
			detail := PushDetail{
				Email:   result.Email,
				Success: false,
				Message: fmt.Sprintf("注册失败,跳过推送: %s", result.Error),
			}
			pushResults.Details = append(pushResults.Details, detail)
			pushResults.Failed++
			log.Printf("[推送 %d/%d] 跳过失败的注册: %s", i+1, len(results), result.Email)
			continue
		}

		// 获取完整账号信息
		log.Printf("[推送 %d/%d] 获取账号信息: %s", i+1, len(results), result.Email)
		info, err := clerk.FetchAccountInfo(result.ClientCookie)
		if err != nil {
			detail := PushDetail{
				Email:   result.Email,
				Success: false,
				Message: fmt.Sprintf("获取账号信息失败: %v", err),
			}
			pushResults.Details = append(pushResults.Details, detail)
			pushResults.Failed++
			log.Printf("[推送 %d/%d] 获取账号信息失败: %v", i+1, len(results), err)
			continue
		}

		// 构造推送数据
		emailParts := strings.Split(result.Email, "@")
		accountData := AccountPushData{
			Name:         "Auto-" + emailParts[0],
			Email:        info.Email,
			ClientCookie: info.ClientCookie,
			ClientUat:    info.ClientUat,
			SessionID:    info.SessionID,
			UserID:       info.UserID,
			ProjectID:    info.ProjectID,
			AgentMode:    "claude-opus-4.5",
			Weight:       1,
			Enabled:      true,
		}

		// 推送到远程 API
		success, message := pushToAPI(accountData, apiURL, username, password)
		detail := PushDetail{
			Email:   result.Email,
			Success: success,
			Message: message,
		}
		pushResults.Details = append(pushResults.Details, detail)

		if success {
			pushResults.Success++
			log.Printf("[推送 %d/%d] 成功: %s", i+1, len(results), result.Email)
		} else {
			pushResults.Failed++
			log.Printf("[推送 %d/%d] 失败: %s - %s", i+1, len(results), result.Email, message)
		}

		// 避免请求过快
		time.Sleep(500 * time.Millisecond)
	}

	return pushResults
}

// pushToAPI 推送单个账号到 API
func pushToAPI(account AccountPushData, apiURL, username, password string) (bool, string) {
	// 构造 JSON 请求体
	jsonData, err := json.Marshal(account)
	if err != nil {
		return false, fmt.Sprintf("JSON 编码失败: %v", err)
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return false, fmt.Sprintf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 添加 Basic Auth
	auth := username + ":" + password
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Set("Authorization", "Basic "+encodedAuth)

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Sprintf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, _ := io.ReadAll(resp.Body)

	// 检查状态码
	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
		return true, fmt.Sprintf("HTTP %d - 推送成功", resp.StatusCode)
	}

	return false, fmt.Sprintf("HTTP %d - %s", resp.StatusCode, string(body))
}

// maskURL 遮蔽 URL 中的敏感信息
func maskURL(url string) string {
	if url == "" {
		return "(未配置)"
	}
	// 只显示协议和主机名
	if idx := strings.Index(url, "://"); idx > 0 {
		if idx2 := strings.Index(url[idx+3:], "/"); idx2 > 0 {
			return url[:idx+3+idx2] + "/***"
		}
	}
	return url
}

// maskString 遮蔽字符串
func maskString(s string) string {
	if s == "" {
		return "(未配置)"
	}
	if len(s) <= 2 {
		return "***"
	}
	return s[:1] + "***" + s[len(s)-1:]
}
