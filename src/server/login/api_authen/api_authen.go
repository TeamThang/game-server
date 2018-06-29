package api

import (
	"fmt"
	"github.com/name5566/leaf/db/redis"
	rredis "github.com/gomodule/redigo/redis"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/util"
	"strconv"
	"strings"
	"time"
	"sort"
)

// AccessKey采用k-v存储redis中
const AccessKeyFmt string = "ACCESSKEY:%s"    // 存入redis的key格式,TOKEN:uuid
const KeyValFmt string = "%s_%s_%d"           // 存入redis的value格式， {userID}_{secretKey}_{timestamp}
const AccessSetFmt string = "ACCESSKEYSET:%s" // 记录用户的accessKey,  {userID}_{AccessKeySet}

// 设置accessKey到redis
func SetAccessKey(userID uint) (map [string] string, error) {
	ret := make(map[string] string)
	accessKey, err := genAccessKey()
	if err != nil {
		return ret, fmt.Errorf("accessKey generate err: %s", err)
	}
	secretKey, err := genSecretKey()
	if err != nil {
		return ret, fmt.Errorf("secretKey generate err: %s", err)
	}
	redisKey := fmt.Sprintf(AccessKeyFmt, accessKey) // 存入redis时格式化
	redisVal := fmt.Sprintf(KeyValFmt, strconv.Itoa(int(userID)), secretKey, time.Now().Unix())
	log.Debug("redis access key: %s, redis access value: %s", redisKey, redisVal)

	_, err = redis.Do("set", redisKey, redisVal)
	if err != nil {
		return ret, fmt.Errorf("accessKey set 1 err: %s", err)
	}
	setKey := fmt.Sprintf(AccessSetFmt, userID)
	_, err = redis.Do("sadd", setKey, accessKey)
	if err != nil {
		return ret, fmt.Errorf("accessKey set 2 err: %s", err)
	}
	ret["AccessKey"] = accessKey
	ret["SecretKey"] = secretKey
	return ret, nil
}

// 从redis中读取AccessKey的值
// 获取对应的userID
func GetAccessKey(accessKey string) (userID uint, secretKey string, timeStamp string, err error) {
	redisKey := fmt.Sprintf(AccessKeyFmt, accessKey) // redis格式化
	res, err := rredis.String(redis.Do("get", redisKey))
	if err != nil {
		if err.Error() == "redigo: nil returned" {
			err = fmt.Errorf("accessKey is not right")
			return
		}
		err = fmt.Errorf("accessKey get by id 1 err: %s", err)
		return
	}

	value := strings.Split(res, "_")
	userIDInt, err := strconv.Atoi(value[0])
	userID = uint(userIDInt)
	secretKey = value[1]
	timeStamp = value[2]
	if err != nil {
		err = fmt.Errorf("accessKey get from id 2 err: %s", err)
		return
	}
	return
}

// 从redis删除对应的AccessKey
func DelAccessKey(accessKey string) (error) {
	redisKey := fmt.Sprintf(AccessKeyFmt, accessKey) // redis格式化
	res, err := rredis.String(redis.Do("get", redisKey))
	if err != nil {
		if err.Error() == "redigo: nil returned" {
			return fmt.Errorf("accessKey is not right")
		}
		return err
	}
	_, err = redis.Do("del", redisKey)
	if err != nil {
		return fmt.Errorf("accessKey delete 1 err: %s", err)
	}
	value := strings.Split(res, "_")
	userIDInt, err := strconv.Atoi(value[0])
	userID := uint(userIDInt)
	setKey := fmt.Sprintf(AccessSetFmt, userID)
	_, err = redis.Do("srem", setKey, accessKey)
	if err != nil {
		return fmt.Errorf("accessKey delete 2 err: %s", err)
	}
	return nil
}

type ApiKey struct {
	AccessKey string
	SecretKey string
	CreatedAt string
	createdAt int
}

type ApiKeys []ApiKey
// 根据创建时间逆序排列
type SortByCreatedAt struct{ ApiKeys }

func (p SortByCreatedAt) Less(i, j int) bool {
	return p.ApiKeys[i].createdAt > p.ApiKeys[j].createdAt
}

func (p SortByCreatedAt) Len() int {
	return len(p.ApiKeys)
}

func (p SortByCreatedAt) Swap(i, j int) {
	p.ApiKeys[i], p.ApiKeys[j] = p.ApiKeys[j], p.ApiKeys[i]
}

// 通过userID获取对应的accessKey和secretKey
func GetUserAccessKeys(userID uint) (apiAuthens *ApiKeys, err error) {
	setKey := fmt.Sprintf(AccessSetFmt, userID)
	resList, e := rredis.Strings(redis.Do("SMEMBERS", setKey))
	if e != nil {
		err = fmt.Errorf("accessKey get 1 err: %s", e)
		return
	}
	apiKeys := ApiKeys{}
	for _, accessKey := range resList {
		_, secretKey, timeStamp, e := GetAccessKey(accessKey)
		if e != nil {
			err = fmt.Errorf("accessKey get 2 err: %s", e)
			return
		}
		timeStampInt, e := strconv.Atoi(timeStamp)
		if e != nil {
			err = fmt.Errorf("accessKey time stamp err: %s, accessKey: %s", e, accessKey)
			return
		}
		t := time.Unix(int64(timeStampInt), 0)

		apiKeys = append(apiKeys, ApiKey{
			AccessKey: accessKey,
			SecretKey: secretKey,
			CreatedAt: t.Format("2006/01/02 15:04:05"),
			createdAt: timeStampInt,
		})
	}
	sort.Sort(SortByCreatedAt{apiKeys})
	apiAuthens = &apiKeys
	return
}

// 从redis清除对应userID的accessKey
func CleanUserAccessKey(userID uint) (error) {
	setKey := fmt.Sprintf(AccessSetFmt, userID)
	resList, err := rredis.Strings(redis.Do("SMEMBERS", setKey))
	if err != nil {
		return fmt.Errorf("accessKey clean 1 err: %s", err)
	}
	for _, accessKey := range resList {
		_, err := redis.Do("del", accessKey)
		if err != nil {
			return fmt.Errorf("accessKey clean 2 err:  %s", err)
		}
	}
	_, err = redis.Do("del", setKey)
	if err != nil {
		return fmt.Errorf("accessKey clean 2 err:  %s", err)
	}
	return nil
}

// 生成AccessKey
func genAccessKey() (string, error) {
	uuid, err := util.GetNumUUID()
	if err != nil {
		return "", err
	}
	return uuid, nil
}

// 生成SecretKey
func genSecretKey() (string, error) {
	uuid, err := util.GetNumUUIDV1()
	if err != nil {
		return "", err
	}
	return uuid, nil
}
