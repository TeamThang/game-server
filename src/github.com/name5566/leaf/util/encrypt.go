package util

import (
	"fmt"
	"crypto/md5"
	"encoding/hex"
	"github.com/satori/go.uuid"
	"encoding/base64"
	"strings"
)

// 获取字符串的md5
func GetMD5(s string) string {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(s))
	cipherStr := md5Ctx.Sum(nil)
	return hex.EncodeToString(cipherStr)
}

// 生成uuid
func GetUUID() (string, error) {
	u1, err := uuid.NewV4()
	if err != nil {
		return "", fmt.Errorf("uuid v4 gerator wrong: %s", err)
	}
	return u1.String(), nil
}

// 生成uuid, 没有"-"
func GetNumUUID() (string, error) {
	u1, err := uuid.NewV4()
	if err != nil {
		return "", fmt.Errorf("uuid v4 without '-' gerator wrong: %s", err)
	}
	ret := strings.Replace(u1.String(), "-", "", -1)
	return ret, nil
}

// 生成v1版本uuid, 没有"-"
func GetNumUUIDV1() (string, error) {
	u1, err := uuid.NewV1()
	if err != nil {
		return "", fmt.Errorf("uuid v1 gerator wrong: %s", err)
	}
	ret := strings.Replace(u1.String(), "-", "", -1)
	return ret, nil
}

func EncodeB64(message string) (retour string) {
	base64Text := make([]byte, base64.StdEncoding.EncodedLen(len(message)))
	base64.StdEncoding.Encode(base64Text, []byte(message))
	return string(base64Text)
}

func DecodeB64(message string) (retour string, err error) {
	base64Text := make([]byte, base64.StdEncoding.DecodedLen(len(message)))
	if l, err := base64.StdEncoding.Decode(base64Text, []byte(message)); err == nil{
		retour = string(base64Text[:l])
	}
	return
}