package token

import (
	"fmt"
	"github.com/name5566/leaf/db/redis"
	rredis "github.com/gomodule/redigo/redis"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/util"
	"strconv"
	"strings"
)

// token采用k-v存储redis中，设置过期时间
// token为uuid, 值为User.ID
const TokenKeyFmt string = "QUANTITY_TOKEN:%s" // 存入redis的key格式,TOKEN:uuid
const TokenValFmt string = "%s_%s"             // 存入redis的value格式， {userID}_{token from login server}

// 设置session到redis
// loginName: 登录名; reqRight: 登录鉴权信息; duration: session超时时间
func SetSessionID(userID uint, duration uint, tokenLogin string) (string, error) {
	token, err := genToken()
	if err != nil {
		return "", fmt.Errorf("token generate err: %s", err)
	}
	tokenVal := fmt.Sprintf(TokenValFmt, strconv.Itoa(int(userID)), tokenLogin)
	log.Debug("token: %s, val: %s", token, tokenVal)
	tokenKey := fmt.Sprintf(TokenKeyFmt, token) // 存入redis时格式化
	_, err = redis.Do("set", tokenKey, tokenVal)
	if err != nil {
		return "", fmt.Errorf("token set 1 err: %s", err)
	}
	_, err = redis.Do("expire", tokenKey, duration)
	if err != nil {
		return "", fmt.Errorf("token set 2 err: %s", err)
	}
	return token, nil
}

// 从redis中读取token的值
// 获取对应的userID
func GetTokenValByID(token string) (userID uint, loginToken string, maxAge uint, err error) {
	tokenKey := fmt.Sprintf(TokenKeyFmt, token) // redis格式化
	res, err := rredis.String(redis.Do("get", tokenKey))
	if err != nil {
		if err.Error() == "redigo: nil returned" {
			err = fmt.Errorf("unlogin")
			return
		}
		err = fmt.Errorf("token val get by id err: %s", err)
		return
	}
	maxAgeRes, err := rredis.Int(redis.Do("pttl", tokenKey))
	if err != nil {
		err = fmt.Errorf("token [%s] is not set expired time", tokenKey)
		return
	}
	maxAge = uint(maxAgeRes / 1000) // 毫秒转换为秒

	value := strings.Split(res, "_")
	userIDInt, err := strconv.Atoi(value[0])
	userID = uint(userIDInt)
	loginToken = value[1]
	if err != nil {
		err = fmt.Errorf("token val get from id err: %s", err)
		return
	}
	return
}

// 根据当前系统token获取登陆服务的userID和token
func GetLoginTokenByID(token string) (loginUserID uint, loginToken string, err error) {
	tokenKey := fmt.Sprintf(TokenKeyFmt, token)
	res, err := rredis.String(redis.Do("get", tokenKey))
	if err != nil {
		if err.Error() == "redigo: nil returned" {
			err = fmt.Errorf("unlogin")
			return
		}
		err = fmt.Errorf("token val get by id err: %s", err)
		return
	}

	value := strings.Split(res, "_")
	userIDInt, err := strconv.Atoi(value[0])
	loginUserID = uint(userIDInt)
	loginToken = value[1]
	if err != nil {
		err = fmt.Errorf("token val get from id err: %s", err)
		return
	}
	return
}

// 从redis删除对应的session
func DelSessionID(token string) (error) {
	tokenKey := fmt.Sprintf(TokenKeyFmt, token) // 存入redis时格式化
	_, err := rredis.String(redis.Do("get", tokenKey))
	if err != nil {
		if err.Error() == "redigo: nil returned" {
			return fmt.Errorf("unlogin")
		}
		return err
	}
	_, err = redis.Do("del", tokenKey)
	if err != nil {
		return fmt.Errorf("token delete err: %s", err)
	}
	return nil
}

// 从redis清除对应userID的session
func CleanSessionID(userID uint) (error) {
	sessionQueryKey := fmt.Sprintf(TokenKeyFmt, "*")
	tokens, err := rredis.Strings(redis.Do("keys", sessionQueryKey))
	if err != nil {
		return fmt.Errorf("token clean err: %s", err)
	}
	for _, tokenKey := range tokens {
		res, err := rredis.String(redis.Do("get", tokenKey))
		if err != nil {
			return fmt.Errorf("token clean err:  %s", err)
		}
		value := strings.Split(res, "_")
		uID, err := strconv.Atoi(value[0])
		if err != nil {
			return fmt.Errorf("token val get from id err: %s", err)
		}
		if (uint(uID) == userID) {
			_, err := redis.Do("del", tokenKey)
			if err != nil {
				return fmt.Errorf("token clean err:  %s", err)
			}
		}
	}
	return nil
}

// 生成token
func genToken() (string, error) {
	uuid, err := util.GetUUID()
	if err != nil {
		return "", err
	}
	return uuid, nil
}

// 检查是否登陆
func CheckLogin(token string) bool {
	userID, _, _, err := GetTokenValByID(token)
	if err != nil {
		log.Error("check login failed: %v\n", err)
		return false
	}
	if userID == 0 {
		return false
	}
	return true
}
