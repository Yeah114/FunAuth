package auth

import (
	"context"
	"fmt"
	"os"

	g79 "github.com/Yeah114/g79client"
	"github.com/Yeah114/unmcpk"
)

// TransferCheckNum 调用 unmcpk 模块生成校验值
func TransferCheckNum(ctx context.Context, isPC bool, data, engineVersion, patchVersion string) (string, error) {
	if engineVersion == "" {
		engineVersion = g79.EngineVersion
	}
	if patchVersion == "" {
		latestVersion, err := g79.GetGlobalG79LatestVersion()
		if err != nil {
			return "", fmt.Errorf("get latest version failed")
		}
		patchVersion = latestVersion
	}

	python3Path := os.Getenv("FUNAUTH_PYTHON3")
	value, err := unmcpk.GenerateTransferCheckNum(isPC, data, engineVersion, patchVersion, python3Path)
	if err != nil {
		return "", err
	}
	return value, nil
}
