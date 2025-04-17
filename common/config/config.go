package config

import (
	"alexsidebar2api/common/env"
	google_api "alexsidebar2api/google-api"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"
)

var BackendSecret = os.Getenv("BACKEND_SECRET")
var ASCookie = os.Getenv("AS_COOKIE")
var IpBlackList = strings.Split(os.Getenv("IP_BLACK_LIST"), ",")
var ProxyUrl = env.String("PROXY_URL", "")
var ChineseChatEnabled = env.Bool("CHINESE_CHAT_ENABLED", true)
var ApiSecret = os.Getenv("API_SECRET")
var ApiSecrets = strings.Split(os.Getenv("API_SECRET"), ",")
var UserAgent = env.String("USER_AGENT", "AlexSideBar/191 CFNetwork/1568.300.101 Darwin/24.2.0")

var RateLimitCookieLockDuration = env.Int("RATE_LIMIT_COOKIE_LOCK_DURATION", 10*60)

// 隐藏思考过程
var ReasoningHide = env.Int("REASONING_HIDE", 0)

// 前置message
var PRE_MESSAGES_JSON = env.String("PRE_MESSAGES_JSON", "")

// 路由前缀
var RoutePrefix = env.String("ROUTE_PREFIX", "")
var SwaggerEnable = os.Getenv("SWAGGER_ENABLE")
var BackendApiEnable = env.Int("BACKEND_API_ENABLE", 1)

var DebugEnabled = os.Getenv("DEBUG") == "true"

var RateLimitKeyExpirationDuration = 20 * time.Minute

var RequestOutTimeDuration = 5 * time.Minute

var (
	RequestRateLimitNum            = env.Int("REQUEST_RATE_LIMIT", 60)
	RequestRateLimitDuration int64 = 1 * 60
)

type RateLimitCookie struct {
	ExpirationTime time.Time // 过期时间
}

var (
	rateLimitCookies sync.Map // 使用 sync.Map 管理限速 Cookie
)

func AddRateLimitCookie(cookie string, expirationTime time.Time) {
	rateLimitCookies.Store(cookie, RateLimitCookie{
		ExpirationTime: expirationTime,
	})
	//fmt.Printf("Storing cookie: %s with value: %+v\n", cookie, RateLimitCookie{ExpirationTime: expirationTime})
}

type ASTokenInfo struct {
	//ApiKey       string
	RefreshToken string
	AccessToken  string
}

var (
	ASTokenMap   = map[string]ASTokenInfo{}
	ASCookies    []string   // 存储所有的 cookies
	cookiesMutex sync.Mutex // 保护 ASCookies 的互斥锁
)

func InitASCookies() ([]string, error) {
	cookiesMutex.Lock()
	defer cookiesMutex.Unlock()

	ASCookies = []string{}

	// 从环境变量中读取 AS_COOKIE 并拆分为切片
	cookieStr := os.Getenv("AS_COOKIE")
	if cookieStr != "" {

		for _, cookie := range strings.Split(cookieStr, ",") {
			cookie = strings.TrimSpace(cookie)
			request := google_api.RefreshTokenRequest{
				RefreshToken: cookie,
			}

			response, err := google_api.GetFirebaseToken(request)
			if err != nil {
				return nil, fmt.Errorf("GetFirebaseToken err %v , Req: %v", err, request)
			}
			ASTokenMap[cookie] = ASTokenInfo{
				RefreshToken: response.RefreshToken,
				AccessToken:  response.AccessToken,
			}
			ASTokenMap[cookie] = ASTokenInfo{
				//ApiKey:       split[0],
				RefreshToken: response.RefreshToken,
				AccessToken:  response.AccessToken,
			}
			ASCookies = append(ASCookies, cookie)
		}
	}
	return ASCookies, nil
}

type CookieManager struct {
	Cookies      []string
	currentIndex int
	mu           sync.Mutex
}

// GetASCookies 获取 ASCookies 的副本
func GetASCookies() []string {
	//cookiesMutex.Lock()
	//defer cookiesMutex.Unlock()

	// 返回 ASCookies 的副本，避免外部直接修改
	cookiesCopy := make([]string, len(ASCookies))
	copy(cookiesCopy, ASCookies)
	return cookiesCopy
}

func NewCookieManager() *CookieManager {
	var validCookies []string
	// 遍历 ASCookies
	for _, cookie := range GetASCookies() {
		cookie = strings.TrimSpace(cookie)
		if cookie == "" {
			continue // 忽略空字符串
		}

		// 检查是否在 RateLimitCookies 中
		if value, ok := rateLimitCookies.Load(cookie); ok {
			rateLimitCookie, ok := value.(RateLimitCookie) // 正确转换为 RateLimitCookie
			if !ok {
				continue
			}
			if rateLimitCookie.ExpirationTime.After(time.Now()) {
				// 如果未过期，忽略该 cookie
				continue
			} else {
				// 如果已过期，从 RateLimitCookies 中删除
				rateLimitCookies.Delete(cookie)
			}
		}

		// 添加到有效 cookie 列表
		validCookies = append(validCookies, cookie)
	}

	return &CookieManager{
		Cookies:      validCookies,
		currentIndex: 0,
	}
}

func (cm *CookieManager) GetRandomCookie() (string, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if len(cm.Cookies) == 0 {
		return "", errors.New("no cookies available")
	}

	// 生成随机索引
	randomIndex := rand.Intn(len(cm.Cookies))
	// 更新当前索引
	cm.currentIndex = randomIndex

	return cm.Cookies[randomIndex], nil
}

func (cm *CookieManager) GetNextCookie() (string, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if len(cm.Cookies) == 0 {
		return "", errors.New("no cookies available")
	}

	cm.currentIndex = (cm.currentIndex + 1) % len(cm.Cookies)
	return cm.Cookies[cm.currentIndex], nil
}

// RemoveCookie 删除指定的 cookie（支持并发）
func RemoveCookie(cookieToRemove string) {
	cookiesMutex.Lock()
	defer cookiesMutex.Unlock()

	// 创建一个新的切片，过滤掉需要删除的 cookie
	var newCookies []string
	for _, cookie := range GetASCookies() {
		if cookie != cookieToRemove {
			newCookies = append(newCookies, cookie)
		}
	}

	// 更新 GSCookies
	ASCookies = newCookies
}
