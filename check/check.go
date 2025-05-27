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

	if config.CfClearance == "" {
		logger.FatalLog("环境变量 CF_CLEARANCE 未设置")
	}

	logger.SysLog("environment variable check passed.")
}
