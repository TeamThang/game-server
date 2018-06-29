package util

import (
	"time"
	"fmt"
	"github.com/name5566/leaf/log"
)

// 获取maxAge秒之后的时间
func GetExpiredTime(maxAge int) (expiredTime time.Time) {
	current := time.Now()
	maxAgeStr := fmt.Sprintf("%ds", maxAge)
	ss, err := time.ParseDuration(maxAgeStr)
	if err != nil {
		log.Error("%s is not right time duration" , maxAgeStr)
	}
	expiredTime = current.Add(ss)
	return
}
