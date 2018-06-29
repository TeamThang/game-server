package http

import (
	"net/http"
	"server/login/api_authen"
	"fmt"
	"server/api/html"
)

var HttpServeMux = http.NewServeMux()

func init() {
	initHtmlHttp()
}

func initHtmlHttp() {
	HttpServeMux.HandleFunc("/login", html.ShowLoginTest)
}

type AuthKey struct {
	AccessKey string
	SecretKey string
}

// 验证用户
// accessKey:对应指定用户， sercretKey: 验证是否有效
func httpAuthenMiddleWare (handler http.Handler) http.Handler {
	ourFunc := func(w http.ResponseWriter, r *http.Request) {
		var authKey AuthKey
		authKey.AccessKey = r.Header.Get("AccessKey")
		authKey.SecretKey = r.Header.Get("SecretKey")

		if _, secretKey, _, err := api.GetAccessKey(authKey.AccessKey); err == nil {
			if secretKey == authKey.SecretKey {  // secretkey当前只做匹配验证
				handler.ServeHTTP(w, r)
			} else {
				w.WriteHeader(403)
				w.Write([]byte(`{"message": "SecretKey is not match","status": 403}`))
			}
		} else {
			w.WriteHeader(403)
			w.Write([]byte(fmt.Sprintf(`{"message": "%s","status": 403}`, err.Error())))
		}
	}
	return http.HandlerFunc(ourFunc)
}