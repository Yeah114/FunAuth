package auth

import (
	"github.com/Yeah114/g79client/utils"
)

// TransferStartType 解密 contentHex，拼接 uid 到明文前缀，再加密返回。
func TransferStartType(uid, contentHex string) (string, error) {
	plain, err := utils.G79HttpDecrypt(contentHex)
	if err != nil {
		return "", err
	}
	merged := uid + plain
	enc, err := utils.G79HttpEncrypt(merged)
	if err != nil {
		return "", err
	}
	return enc, nil
}
