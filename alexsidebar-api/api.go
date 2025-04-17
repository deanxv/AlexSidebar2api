package alexsidebar_api

import (
	"alexsidebar2api/common"
	"alexsidebar2api/common/config"
	logger "alexsidebar2api/common/loggger"
	"alexsidebar2api/cycletls"
	"fmt"
	"github.com/gin-gonic/gin"
)

const (
	baseURL      = "https://api.alexcodes.app"
	chatEndpoint = baseURL + "/call_assistant5"
)

func MakeStreamChatRequest(c *gin.Context, client cycletls.CycleTLS, jsonData []byte, cookie string) (<-chan cycletls.SSEResponse, error) {
	//split := strings.Split(cookie, "=")
	tokenInfo, ok := config.ASTokenMap[cookie]
	if !ok {
		return nil, fmt.Errorf("cookie not found in ASTokenMap")
	}

	options := cycletls.Options{
		Timeout: 10 * 60 * 60,
		Proxy:   config.ProxyUrl, // 在每个请求中设置代理
		Body:    string(jsonData),
		Method:  "POST",
		Headers: map[string]string{
			"User-Agent":       config.UserAgent,
			"Content-Type":     "application/json",
			"app-version":      "2.5.7",
			"app-build-number": "191",
			"auth":             tokenInfo.AccessToken,
			"accept-language":  "zh-CN,zh-Hans;q=0.9",
			"local-id":         common.GenerateSerialNumber(10),
		},
	}

	logger.Debug(c.Request.Context(), fmt.Sprintf("cookie: %v", cookie))

	logger.Debug(c, fmt.Sprintf("%s", options))

	sseChan, err := client.DoSSE(chatEndpoint, options, "POST")
	if err != nil {
		logger.Errorf(c, "Failed to make stream request: %v", err)
		return nil, fmt.Errorf("failed to make stream request: %v", err)
	}
	return sseChan, nil
}
