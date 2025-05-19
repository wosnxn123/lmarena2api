package check

import (
	"lmarena2api/common/config"
	logger "lmarena2api/common/loggger"
)

func CheckEnvVariable() {
	logger.SysLog("environment variable checking...")

	if config.LACookie == "" {
		logger.FatalLog("环境变量 LA_COOKIE 未设置")
	}

	logger.SysLog("environment variable check passed.")
}
