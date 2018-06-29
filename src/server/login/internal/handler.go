package internal

import (
	"reflect"
	"fmt"
	"net/http"
	"strings"
	"server/msg"
	"io/ioutil"
	"encoding/json"
	"github.com/name5566/leaf/conf"
	"github.com/name5566/leaf/util"
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/db/postgre"
	"github.com/name5566/leaf/db/postgre/model"
	tk "github.com/name5566/leaf/db/redis/token"
	"github.com/name5566/leaf/gate/user"
	"server/login/api_authen"
	"server/msg/account"
)

func handleMsg(m interface{}, h interface{}) {
	skeleton.RegisterChanRPC(reflect.TypeOf(m), h)
}

func init() {
	handleMsg(&account.ApiKeyCreate{}, HandleApiKeyCreate)
	handleMsg(&account.ApiKeyQuery{}, HandleApiKeyQuery)
	handleMsg(&account.ApiKeyDelete{}, HandleApiKeyDelete)
}

type ReqID struct {
	UserID  uint
	RightID uint
	EmailID uint
	Code    string
	Token   string
}

type LoginServerReq struct {
	Uri    string
	Method string
}

// websocket消息对应Restful接口格式
var logServerMap = map[string]LoginServerReq{

	"UserCreate": {Uri: "/v1/account", Method: "POST"},
	"UserQuery":  {Uri: "/v1/account/%d", Method: "GET"},
	"UserUpdate": {Uri: "/v1/account/%d", Method: "PUT"},
	"UserDelete": {Uri: "/v1/account/%d", Method: "DELETE"},

	"RightCreate": {Uri: "/v1/right", Method: "POST"},
	"RightQuery":  {Uri: "/v1/right", Method: "GET"},
	"RightUpdate": {Uri: "/v1/right/%d", Method: "PUT"},
	"RightDelete": {Uri: "/v1/right/%d", Method: "DELETE"},

	"RightBind":      {Uri: "/v1/bind_auth", Method: "POST"},
	"RightUnBind":    {Uri: "/v1/bind_auth/user/%d/right/%d", Method: "DELETE"},
	"BindRightQuery": {Uri: "/v1/bind_auth/user/%d", Method: "GET"},

	"Login":       {Uri: "/v1/login", Method: "POST"},
	"Logout":      {Uri: "/v1/logout", Method: "GET"},
	"GetUserInfo": {Uri: "/v1/get_user_info", Method: "GET"},

	"VerifyEmailSend":  {Uri: "/v1/verify/email/user/%d", Method: "GET"},
	"VerifyEmailCheck": {Uri: "/v1/verify/email/user/%d/code/%s", Method: "PUT"},

	"NotifyEmailCreate": {Uri: "/v1/notify/email", Method: "POST"},
	"NotifyEmailQuery":  {Uri: "/v1/notify/email/user/%d", Method: "GET"},
	"NotifyEmailSend":   {Uri: "/v1/notify/email/user/%d", Method: "POST"},
	"NotifyEmailDelete": {Uri: "/v1/notify/email/%d", Method: "DELETE"},
	"NotifyEmailSub":    {Uri: "/v1/notify/email/%d/user/%d/sub", Method: "PUT"},
	"NotifyEmailUnSub":  {Uri: "/v1/notify/email/%d/user/%d/unsub", Method: "PUT"},

	"PwdChange":         {Uri: "/v1/pwd/user/%d", Method: "PUT"},
	"EmailRestPwdSend":  {Uri: "/v1/pwd", Method: "POST"},
	"EmailRestPwdCheck": {Uri: "/v1/pwd/user/%d/code/%s", Method: "GET"},
}

// 构造Restful请求的uri
func make_uri(msgID string, req json.RawMessage, reqID ReqID) (uri string, err error) {
	logServerReq := logServerMap[msgID]
	switch msgID {
	case "UserQuery", "UserUpdate", "UserDelete", "BindRightQuery", "VerifyEmailSend", "NotifyEmailQuery", "NotifyEmailSend", "PwdChange": // uri中只包含userID
		err := json.Unmarshal(req, &reqID)
		if err != nil {
			return "", err
		}
		uri = fmt.Sprintf(logServerReq.Uri, reqID.UserID)
		return uri, nil
	case "RightUpdate", "RightDelete": // uri中只包含rightID
		err := json.Unmarshal(req, &reqID)
		if err != nil {
			return "", err
		}
		uri = fmt.Sprintf(logServerReq.Uri, reqID.RightID)
		return uri, nil
	case "NotifyEmailDelete": // uri中只包含emailID
		err := json.Unmarshal(req, &reqID)
		if err != nil {
			return "", err
		}
		uri = fmt.Sprintf(logServerReq.Uri, reqID.EmailID)
		return uri, nil
	case "RightUnBind": // uri中包含userID和rightID
		err := json.Unmarshal(req, &reqID)
		if err != nil {
			return "", err
		}
		uri = fmt.Sprintf(logServerReq.Uri, reqID.UserID, reqID.RightID)
		return uri, nil
	case "VerifyEmailCheck", "EmailRestPwdCheck": // uri中包含userID和Code
		err := json.Unmarshal(req, &reqID)
		if err != nil {
			return "", err
		}
		uri = fmt.Sprintf(logServerReq.Uri, reqID.UserID, reqID.Code)
		return uri, nil
	case "NotifyEmailSub", "NotifyEmailUnSub": // uri中包含emailID和userID
		err := json.Unmarshal(req, &reqID)
		if err != nil {
			return "", err
		}
		uri = fmt.Sprintf(logServerReq.Uri, reqID.EmailID, reqID.UserID)
		return uri, nil
	default:
		return logServerReq.Uri, nil
	}
}

// 解析登陆服务器返回的body
func parseResBody(statusCode int, body []byte, otherData map[string]interface{}) (*msg.Response) {
	var m map[string]json.RawMessage
	json.Unmarshal(body, &m)
	var message interface{}
	data := make(map[string]interface{})
	for k, v := range m {
		if k == "message" {
			message = v
		} else {
			if k != "status" {
				data[k] = v
			}
		}
	}
	if otherData != nil {
		for k, v := range otherData {
			data[k] = v
		}
	}
	return &msg.Response{Status: statusCode, Message: message, Data: data}
}

// 构造返回消息
// {msgID: {}}
func makeResponse(msgID string, msgRes *msg.Response) *map[string]interface{} {

	resMsg := make(map[string]interface{})
	resMsg[msgID] = map[string]interface{}{
		"status":  msgRes.Status,
		"message": msgRes.Message,
		"data":    msgRes.Data,
	}
	return &resMsg
}

// 登陆handler
// 转发请求消息到登陆服务
func HandleLogin(args []interface{}) {
	a, userID, token, err := GetUserInfo(args[2])
	if err != nil && string(err.Error()) != "unlogin" {
		a.WriteMsg(&msg.Response{Status: 400, Message: string(err.Error())})
		return
	}

	// 解析请求参数
	msgID, ok := args[0].(string)
	if !ok {
		a.WriteMsg(&msg.Response{Status: 400, Message: "req format is not right"})
		return
	}
	logServerReq := logServerMap[msgID]
	fmt.Println(msgID, ok)
	req, ok := args[1].(json.RawMessage)
	fmt.Println(req, ok)
	if !ok {
		msgRes := &msg.Response{Status: 400, Message: "req format is not right"}
		a.WriteMsg(makeResponse(msgID, msgRes))
		return
	}
	var reqJson ReqID
	err = json.Unmarshal(req, &reqJson)
	if err != nil {
		msgRes := &msg.Response{Status: 400, Message: string(err.Error())}
		a.WriteMsg(makeResponse(msgID, msgRes))
		return
	}
	var reqID ReqID
	var loginToken string
	if userID != 0 && token != "" { // token存在则写入
		reqID.UserID, loginToken, err = tk.GetLoginTokenByID(token)
		if err != nil {
			msgRes := &msg.Response{Status: 400, Message: string(err.Error())}
			a.WriteMsg(makeResponse(msgID, msgRes))
			return
		}
		log.Debug("token: %s, login token: %s", token, loginToken)
		req = addUserID(reqID.UserID, req)
	}

	// 请求登陆服务器
	uri, err := make_uri(msgID, req, reqID)
	if err != nil {
		msgRes := &msg.Response{Status: 400, Message: string(err.Error())}
		a.WriteMsg(makeResponse(msgID, msgRes))
		return
	}
	log.Debug("req send to login server: %s, uri: %s", req, uri)
	res, err := httpToLoginServer(logServerReq.Method, uri, string(req), loginToken)
	if err != nil {
		msgRes := &msg.Response{Status: 500, Message: string(err.Error())}
		a.WriteMsg(makeResponse(msgID, msgRes))
		return
	}
	// 返回结果分类处理
	if msgID == "Login" { // 登陆流程处理
		loginResponse(&a, msgID, res)
		return
	}
	if msgID == "Logout" { // 注销流程处理
		logoutResponse(&a, msgID, res, token)
		return
	}
	if msgID == "UserCreate" { // 注册流程处理
		registResponse(&a, msgID, res)
		return
	}
	if msgID == "UserDelete" { // 销户流程处理
		unRegistResponse(&a, msgID, res, userID)
		return
	}
	// 默认返回结果构造
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		msgRes := &msg.Response{Status: 500, Message: string(err.Error())}
		a.WriteMsg(makeResponse(msgID, msgRes))
		return
	}
	msgRes := parseResBody(res.StatusCode, body, nil)
	a.WriteMsg(makeResponse(msgID, msgRes))
}

// http请求到登陆服务器
func httpToLoginServer(method string, uri string, param string, sessionID string) (res *http.Response, err error) {
	client := &http.Client{}
	fmt.Println(util.UrlJoin(conf.Config.LoginServer, uri))
	req, err := http.NewRequest(method, util.UrlJoin(conf.Config.LoginServer, uri),
		strings.NewReader(param))
	if err != nil {
		return
	}
	req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
	res, err = client.Do(req)
	if err != nil {
		return
	}
	return
}

// 登陆流程处理
func loginResponse(a *gate.Agent, msgID string, res *http.Response) {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		msgRes := &msg.Response{Status: 500, Message: string(err.Error())}
		(*a).WriteMsg(makeResponse(msgID, msgRes))
		return
	}
	if res.StatusCode != 201 { // 登陆服务器返回失败
		msgRes := parseResBody(res.StatusCode, body, nil)
		(*a).WriteMsg(makeResponse(msgID, msgRes))
		return
	}
	cookies := res.Cookies()
	tokenCookie := getCookie(cookies, "session_id") // 登陆流程需要解析登陆服务器返回cookie中的token("session_id")并存储
	log.Debug("login server token: %s, max-age: %d\n", tokenCookie.Value, tokenCookie.MaxAge)
	var reqID ReqID
	err = json.Unmarshal(body, &reqID)
	if err != nil {
		msgRes := &msg.Response{Status: 500, Message: string(err.Error())}
		(*a).WriteMsg(makeResponse(msgID, msgRes))
		return
	}
	log.Debug("UserID: %d\n", reqID.UserID)
	if reqID.UserID == 0 {
		msgRes := &msg.Response{Status: 500, Message: fmt.Sprint("call login server failed")}
		(*a).WriteMsg(makeResponse(msgID, msgRes))
		return
	}
	log.Debug("token MaxAge: %d, Expires: %d\n", tokenCookie.MaxAge, tokenCookie.Expires)
	currentToken, err := tk.SetSessionID(reqID.UserID, uint(tokenCookie.MaxAge), tokenCookie.Value) // 生成当前服务的token，并存储登陆服务器token
	if err != nil {
		msgRes := &msg.Response{Status: 500, Message: string(err.Error())}
		(*a).WriteMsg(makeResponse(msgID, msgRes))
		return
	}
	err = synUser(reqID.UserID) // 同步用户，可能登陆服务器的用户未通过当前服务注册
	if err != nil {
		msgRes := &msg.Response{Status: 500, Message: string(err.Error())}
		(*a).WriteMsg(makeResponse(msgID, msgRes))
		return
	}

	loginData := map[string]interface{}{"token": currentToken, "MaxAge": tokenCookie.MaxAge}
	(*a).SetUserData(user.UserData{
		UserID:  reqID.UserID,
		Token:   currentToken,
		Expired: util.GetExpiredTime(tokenCookie.MaxAge),
	})
	msgRes := parseResBody(201, body, loginData)
	(*a).WriteMsg(makeResponse(msgID, msgRes))
}

// 登出流程处理
func logoutResponse(a *gate.Agent, msgID string, res *http.Response, token string) {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		msgRes := &msg.Response{Status: 500, Message: string(err.Error())}
		(*a).WriteMsg(makeResponse(msgID, msgRes))
		return
	}
	if res.StatusCode != 200 { // 登陆服务器返回失败
		msgRes := parseResBody(res.StatusCode, body, nil)
		(*a).WriteMsg(makeResponse(msgID, msgRes))
		return
	}
	err = tk.DelSessionID(token) // 清空redis中session
	if err != nil {
		msgRes := &msg.Response{Status: 500, Message: string(err.Error())}
		(*a).WriteMsg(makeResponse(msgID, msgRes))
		return
	}
	(*a).SetUserData(nil) // 清空内存中UserData
	msgRes := parseResBody(200, body, nil)
	(*a).WriteMsg(makeResponse(msgID, msgRes))
}

// 注册流程处理
func registResponse(a *gate.Agent, msgID string, res *http.Response) {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		msgRes := &msg.Response{Status: 500, Message: string(err.Error())}
		(*a).WriteMsg(makeResponse(msgID, msgRes))
		return
	}
	if res.StatusCode != 201 { // 登陆服务器返回失败
		msgRes := parseResBody(res.StatusCode, body, nil)
		(*a).WriteMsg(makeResponse(msgID, msgRes))
		return
	}
	var reqID ReqID
	err = json.Unmarshal(body, &reqID)
	if err != nil {
		msgRes := &msg.Response{Status: 500, Message: string(err.Error())}
		(*a).WriteMsg(makeResponse(msgID, msgRes))
		return
	}
	log.Debug("UserID: %d\n", reqID.UserID)
	if reqID.UserID == 0 {
		msgRes := &msg.Response{Status: 500, Message: fmt.Sprint("call login server failed")}
		(*a).WriteMsg(makeResponse(msgID, msgRes))
		return
	}
	err = synUser(reqID.UserID)
	if err != nil {
		msgRes := &msg.Response{Status: 500, Message: string(err.Error())}
		(*a).WriteMsg(makeResponse(msgID, msgRes))
		return
	}
	msgRes := parseResBody(201, body, nil)
	(*a).WriteMsg(makeResponse(msgID, msgRes))
}

// 销户流程处理
func unRegistResponse(a *gate.Agent, msgID string, res *http.Response, userID uint) {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		msgRes := &msg.Response{Status: 500, Message: string(err.Error())}
		(*a).WriteMsg(makeResponse(msgID, msgRes))
		return
	}
	if res.StatusCode != 204 { // 登陆服务器返回失败
		msgRes := parseResBody(res.StatusCode, body, nil)
		(*a).WriteMsg(makeResponse(msgID, msgRes))
		return
	}
	err = delUser(userID)
	if err != nil {
		msgRes := &msg.Response{Status: 500, Message: string(err.Error())}
		(*a).WriteMsg(makeResponse(msgID, msgRes))
		return
	}
	err = tk.CleanSessionID(userID)
	if err != nil {
		msgRes := &msg.Response{Status: 500, Message: string(err.Error())}
		(*a).WriteMsg(makeResponse(msgID, msgRes))
		return
	}
	msgRes := parseResBody(204, body, nil)
	(*a).WriteMsg(makeResponse(msgID, msgRes))
}

// 从[] *http.Cookie中获取cookie
func getCookie(cookies [] *http.Cookie, cookieKey string) (res *http.Cookie) {
	for _, cookie := range cookies {
		if cookie.Name == cookieKey {
			res = cookie
		}
	}
	return
}

// 同步user表
// 检查userID，没有创建
func synUser(userID uint) error {
	db := postgre.DB.Where(&model.User{UserID: userID}).First(&model.User{})
	if db.Error != nil {
		if string(db.Error.Error()) == "record not found" {
			postgre.DB.Create(&model.User{UserID: userID}) // user表没有userID就创建
			log.Release("user[%d] is not existed, added in user table", userID)
		} else {
			return fmt.Errorf("user[%d] add failed", userID)
		}
	}
	return nil
}

// 删除user表
func delUser(userID uint) error {
	user := &model.User{}
	db := postgre.DB.Where(&model.User{UserID: userID}).First(user)
	if db.Error != nil {
		return nil // 不存在就不需要删除
	}
	db = postgre.DB.Delete(user)
	if db.Error != nil {
		return db.Error
	}
	err := tk.CleanSessionID(userID)
	if err != nil {
		return err
	}
	err = api.CleanUserAccessKey(userID)
	if err != nil {
		return err
	}
	return nil
}

// 登陆服务器请求中添加UserID字段
func addUserID(userID uint, rawJson json.RawMessage) (ret json.RawMessage) {
	var err error
	defer func() {
		if err != nil {
			log.Error("json add user id failed: %v", err)
			ret = rawJson
			return
		}
	}()
	rawJsonMap := make(map[string]interface{})
	if err = json.Unmarshal(rawJson, &rawJsonMap); err != nil {
		return
	}
	rawJsonMap["UserID"] = userID
	if data, err := json.Marshal(rawJsonMap); err != nil {
		return
	} else {
		ret = data
	}
	return
}

// 创建 api key
func HandleApiKeyCreate(args []interface{}) {
	a, userID, _, err := GetUserInfo(args[1])
	if err != nil {
		a.WriteMsg(&msg.Response{Status: 400, Message: string(err.Error())})
		return
	}
	data, ok := args[0].(*account.ApiKeyCreate)
	if !ok {
		a.WriteMsg(&msg.Response{Status: 400, Message: "req for create api key is not right"})
		return
	}
	msgID := msg.GetMsgID(data)

	retKey, err := api.SetAccessKey(userID)
	if err != nil {
		msgRes := &msg.Response{Status: 400, Message: string(err.Error())}
		a.WriteMsg(msg.MakeResponse(msgID, msgRes))
		return
	}

	msgRes := &msg.Response{Status: 201, Message: "", Data: retKey}
	a.WriteMsg(msg.MakeResponse(msgID, msgRes))
}

// 查询 api key
func HandleApiKeyQuery(args []interface{}) {
	a, userID, _, err := GetUserInfo(args[1])
	if err != nil {
		a.WriteMsg(&msg.Response{Status: 400, Message: string(err.Error())})
		return
	}
	data, ok := args[0].(*account.ApiKeyQuery)
	if !ok {
		a.WriteMsg(&msg.Response{Status: 400, Message: "req for create api key is not right"})
		return
	}
	msgID := msg.GetMsgID(data)

	apiKeys, err := api.GetUserAccessKeys(userID)
	if err != nil {
		msgRes := &msg.Response{Status: 400, Message: string(err.Error())}
		a.WriteMsg(msg.MakeResponse(msgID, msgRes))
		return
	}
	msgRes := &msg.Response{Status: 200, Message: "", Data: apiKeys}
	a.WriteMsg(msg.MakeResponse(msgID, msgRes))
}

// 删除 api key
func HandleApiKeyDelete(args []interface{}) {
	a, _, _, err := GetUserInfo(args[1])
	if err != nil {
		a.WriteMsg(&msg.Response{Status: 400, Message: string(err.Error())})
		return
	}
	data, ok := args[0].(*account.ApiKeyDelete)
	if !ok {
		a.WriteMsg(&msg.Response{Status: 400, Message: "req for create api key is not right"})
		return
	}
	msgID := msg.GetMsgID(data)

	err = api.DelAccessKey(data.AccessKey)
	if err != nil {
		msgRes := &msg.Response{Status: 400, Message: string(err.Error())}
		a.WriteMsg(msg.MakeResponse(msgID, msgRes))
		return
	}

	msgRes := &msg.Response{Status: 204, Message: "", Data: nil}
	a.WriteMsg(msg.MakeResponse(msgID, msgRes))
}
