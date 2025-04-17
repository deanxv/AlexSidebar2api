package common

import "time"

var StartTime = time.Now().Unix() // unit: second
var Version = "v1.0.3"            // this hard coding will be replaced automatically when building, no need to manually change

type ModelInfo struct {
	Model     string
	MaxTokens int
}

// 创建映射表（假设用 model 名称作为 key）
var ModelRegistry = map[string]ModelInfo{
	"claude-3-7-sonnet":          {"agent_sonnet_37", 100000},
	"claude-3-7-sonnet-thinking": {"agent_sonnet_37", 100000},
	"claude-3-5-sonnet":          {"agent_sonnet", 100000},
	"deepseek-r1":                {"agent_deepseek_r1", 100000},
	"deepseek-v3":                {"deepseek_v3", 100000},
	"o3-mini":                    {"agent_o3_mini", 100000},
	"gpt-4o":                     {"gpt4o", 100000},
	"o1":                         {"o1", 100000},
	//"gemini-2.0":        {"gemini-2.0", 100000},
}

// 通过 model 名称查询的方法
func GetModelInfo(modelName string) (ModelInfo, bool) {
	info, exists := ModelRegistry[modelName]
	return info, exists
}

func GetModelList() []string {
	var modelList []string
	for k := range ModelRegistry {
		modelList = append(modelList, k)
	}
	return modelList
}
