package define

import (
	"runtime"

	"code-platform/config"
)

func InitBasePath() string {
	basePathMap := config.Workspace.GetStringMapString("base_path")
	if runtime.GOOS == "windows" {
		// 开发环境可能为Windows
		return basePathMap["windows"]
	}
	return basePathMap["linux"]
}
