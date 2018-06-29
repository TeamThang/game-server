### 用户账户接口

#### 接口说明

服务交互采用websocket协议，json序列化传输

msgID为每个消息唯一ID

返回码基于Restful风格：

```
201: POST成功
200: GET成功
204: DELETE成功
400: 请求不正确
500: 服务内部错误
```

#### 用户创建

`msgID`: "UserCreate"

`说明`: LoginName，Mobile，Email不能和数据库中已有数据重复

`param`:

arg1 | arg2 | type | desc
-- | -- | -- | --
BasicInfo | LoginName | str | 登录名，必填
BasicInfo | Password | str | 密码(6-20位)，必填
BasicInfo | Mobile | str | 手机，选填
BasicInfo | Email | str | 邮箱，选填
BasicInfo | Source | str | 用户来源
UserInfo | Name | str | 真实姓名,选填
UserInfo | Age | int | 年龄，选填
UserInfo | Birthday | time | 生日(格式:"2018-04-25 18:55:10")，选填
UserInfo | Address | str | 地址, 选填

`request`
```json
{
	"UserCreate": {
		"BasicInfo": {
			"LoginName": "test6",
			"Password": "123",
			"Mobile": "6",
			"Email": "6@163.com",
			"Source": "quantity"
		},
		"UserInfo": {
			"Name": "蜘蛛侠",
			"Age": 123,
			"Birthday": "2018-04-25 18:55:10",
			"Address": "太阳星星月亮"
		}
	}
}
```
`response`:
```json
{
	"UserCreate": {
		"status": 201,  // 消息
		"message": "",  // 请求状态
		"data": {
			"userID": 80
		}
	}
}
```

#### 登录

`msgID`: "Login"

`说明1`: 登录顺序: LoginName->Mobile->Email;这三个至少上报一个

`说明2`: 鉴权顺序: Server->Name，"all"为所有权限，必须配置并赋权给用户

`param`:

arg | type | desc
-- | -- | --
LoginName | int | 登录名，选填
Mobile | str | 登录手机，选填
Email | str | 登录邮箱，选填
Password | str | 登录密码，必填
Right | dict | 权限，Server:登录服务;Name:配置的权限名

`request`
```json
{
	"Login": {
		"Password": "123",
		"LoginName": "test1",
		"Right": {
			"Server": "login",
			"Name": "all"
		}
	}
}
```
`response`:
```json
{
	"Login": {
		"status": 201,
		"message": "login success",
		"data": {
			"token": "1faa8be4-4234-481d-977d-964130409b8a",
			"userID": 62
		}
	}
}
```

#### 发送邮件验证码

`msgID`: "VerifyEmailSend"

`params`:

arg | type | desc
-- | -- | --

`request`
```json
{
	"VerifyEmailSend": {
	}
}
```
`response`:
```json
{"VerifyEmailSend":{"status":200, "message":"","data":{}}}
```

#### 验证邮件验证码

`msgID`: "VerifyEmailCheck"

`params`:

arg | type | desc
-- | -- | --
Code | str | 邮箱验证码，必填

`request`
```json
{
	"VerifyEmailCheck": {
		"Code": "bohGN",
	}
}
```
`response`:
```
{
	"VerifyEmailCheck": {
		"status": 201,
		"message": "",
		"data": {
			"result": true  // 验证结果
		}
	}
}
```

#### 用户销户

`msgID`: "UserDelete"

`备注`: 需要admin权限或鉴权token和UserID是否匹配

`params`:

arg | type | desc
-- | -- | --

`request`
```json
{
	"UserDelete": {
	}
}
```
`response`:
```json
{
	"UserDelete": {
		"data": {},
		"message": null,
		"status": 204
	}
}
```

#### 用户信息修改

`msgID`: "UserUpdate"

`备注`: 需要admin权限或鉴权token和UserID是否匹配

`说明`: 按上报的参数修改当前值，不需要修改的字段就不要上报。LoginName，Mobile，Email不能和现有数据库中重复

`params`:

arg1 | arg2 | type | desc
-- | -- | -- | --
BasicInfo | LoginName | str | 登录名，选填
BasicInfo | Mobile | str | 手机，选填
BasicInfo | Email | str | 邮箱，选填
UserInfo | Name | str | 真实姓名,选填
UserInfo | Age | int | 年龄，选填
UserInfo | Birthday | time | 生日(格式:"2018-04-25 18:55:10")，选填
UserInfo | Address | str | 地址, 选填

`request`
```json
{
	"UserUpdate": {
		"BasicInfo": {
			"LoginName": "test64",
			"Password": "1234"
		},
		"UserInfo": {
			"Name": "钢铁侠",
			"Age": 55,
			"Birthday": "2001-04-25 18:55:10",
			"Address": "太阳星星月亮火箭"
		},
	}
}
```
`response`:
```json
{
	"UserUpdate": {
		"data": {},
		"message": "",
		"status": 201
	}
}
```

#### 用户密码修改

`msgID`: "PwdChange"

`备注`: 需要admin权限或鉴权token和UserID是否匹配

`params`：

arg | type | desc
-- | -- | --
OldPW | str | 当前密码，必填
NewPW |  str | 新密码，必填

`request`
```json
{
	"PwdChange": {
		"OldPW": "123",
		"NewPW": "456",
	}
}
```
`response`:
```json
{
	"PwdChange": {
		"data": {},
		"message": "",
		"status": 201
	}
}
```

#### 用户信息查询


`msgID`: "UserQuery"

`备注`: 需要admin权限或鉴权token和UserID是否匹配

`params`：

arg | type | desc
-- | -- | --


`request`
```
{
	"UserQuery": {
	}
}
```
`response`:
```json
{
	"UserQuery": {
		"data": {
			"data": {
				"BasicInfo": {
					"ID": 64,
					"LoginName": "test64",
					"Mobile": "8",
					"Email": "8@163.com",
					"Source": "quantity",
					"CreatedAt": "2018-05-22 11:03:44",
					"UpdatedAt": "2018-05-22 16:08:10",
					"DeletedAt": "",
					"LoginTime": "0001-01-01 08:05:43",
					"LoginCount": 0
				},
				"UserInfo": {
					"Name": "钢铁侠",
					"Age": 55,
					"Birthday": "2001-04-26 02:55:10",
					"Address": "太阳星星月亮火箭",
					"Other": "aaaaaaa"
				}
			}
		},
		"message": "",
		"status": 200
	}
}
```

#### 账户登录

`msgID`: "Login"

`备注`: 需要admin权限或鉴权token和UserID是否匹配

`params`：

`说明1`: 登录顺序: LoginName->Mobile->Email;这三个至少上报一个

`说明2`: 鉴权顺序: Server->Name，"all"为所有权限，必须配置并赋权给用户

arg | type | desc
-- | -- | --
LoginName | int | 登录名，选填
Mobile | str | 登录手机，选填
Email | str | 登录邮箱，选填
Password | str | 登录密码，必填
Right | dict | 权限，Server:登录服务;Name:配置的权限名


`request`
```json
{
	"Login": {
		"Password": "123",
		"LoginName": "test6",
		"Right": {
			"Server": "login",
			"Name": "all"
		},
	}
}
```
`response`:
```json
statusCode: 201  // CREATED
{
    "message": "login success",
    "status": 201
}
Cookie:
session_id: ae979a3b-3773-488a-b3c6-b4fb5e31e5c0  // 登录服务的token(uuid)
login_server: all  // 成功登录的服务
```

#### 账户注销

`msgID`: "Logout"

`备注`: 需要admin权限或鉴权token和UserID是否匹配

`params`：

arg | type | desc
-- | -- | --

`request`
```json
{
	"Logout": {
	}
}
```
`response`:
```json
{
	"Logout": {
		"data": {},
		"message": "logout success",
		"status": 200
	}
}
```

#### 获当前登录的用户信息

`msgID`: "GetUserInfo"

`备注`: 需要admin权限或鉴权token和UserID是否匹配

`params`：

arg | type | desc
-- | -- | --

`request`
```json
{
	"GetUserInfo": {
	}
}
```
`response`:
```json
{
{
	"GetUserInfo": {
		"data": {
			"data": {
				"BasicInfo": {
					"ID": 62,
					"LoginName": "admin",
					"Mobile": "18520880258",
					"Email": "yuan.zhaoyi@163.com",
					"Source": "quantity",
					"CreatedAt": "2018-05-22 11:01:06",
					"UpdatedAt": "2018-05-22 16:19:59",
					"DeletedAt": "",
					"LoginTime": "2018-05-22 16:19:56",
					"LoginCount": 22
				},
				"UserInfo": {
					"Name": "灭霸",
					"Age": 100,
					"Birthday": "1990-10-05 08:00:00",
					"Address": "太阳星星月亮",
					"Other": null
				}
			}
		},
		"message": "",
		"status": 200
	}
}
```

#### 添加自定义权限

`msgID`: "RightCreate"

`备注`: 需要鉴权token权限，需要login服务admin权限

`params`:

arg | type | desc
-- | -- | --
Server | str | 支持的服务，必填
Name | str | 服务对应权限，必填
Desc| str | 权限描述, 选填

```json
{
	"RightCreate": {
		"Server": "quantity",
		"Name": "all",
		"Desc": "量化服务权限",
	}
}
```
`response`:
```json
{
	"RightCreate": {
		"status": 201,
		"message": "",
		"data": {}
	}
}
```

#### 删除自定义权限

`msgID`: "RightDelete"

`备注`: 需要鉴权token权限，需要login服务admin权限

`params`:

arg | type | desc
-- | -- | --
RightID | str | 权限表ID，必填


`request`
```json
{
	"RightDelete": {
		"RightID": 4,
	}
}
```
`response`:
```json
{
	"RightDelete": {
		"status": 204,
		"message": null,
		"data": {}
	}
}
```

#### 修改自定义权限

`msgID`: "RightUpdate"

`备注`: 需要鉴权token权限，需要login服务admin权限

`params`:

arg | type | desc
-- | -- | --
RightID | str | 权限表ID，必填
Server | str | 支持的服务，选填
Name | str | 服务对应权限，选填
Desc| str | 权限描述, 选填

`request`
```json
{
	"RightUpdate": {
		"RightID": 6,
		"Server": "all",
		"Name": "all",
		"Desc": "所有的权限噢",
	}
}
```
`response`:
```json
{
	"RightUpdate": {
		"status": 201,
		"message": "",
		"data": {}
	}
}
```

#### 查询所有自定义权限

`msgID`: "RightQuery"

`备注`: 需要鉴权token权限，需要login服务admin权限

`params`:

arg | type | desc
-- | -- | --

`request`
```
{
	"RightQuery": {
	}
}
```

`response`:
```json
{
	"RightQuery": {
		"status": 200,
		"message": "",
		"data": {
			"data": [{
				"ID": 3,
				"Server": "login",
				"Name": "all",
				"Desc": "登录服务权限",
				"UserRightRelations": null
			}, {
				"ID": 2,
				"Server": "all",
				"Name": "all",
				"Desc": "所有的权限",
				"UserRightRelations": null
			}, {
				"ID": 5,
				"Server": "login",
				"Name": "all",
				"Desc": "登录服务权限",
				"UserRightRelations": null
			}, {
				"ID": 8,
				"Server": "quantity",
				"Name": "all",
				"Desc": "量化服务权限",
				"UserRightRelations": null
			}, {
				"ID": 6,
				"Server": "all",
				"Name": "all",
				"Desc": "所有的权限噢",
				"UserRightRelations": null
			}]
		}
	}
}
```

#### 用户绑定权限

`msgID`: "RightBind"



`params`:

arg | type | desc
-- | -- | --
RightID | str | 权限ID，必填

`request`
```
{
	"RightBind": {
		"RightID": 3,
	}
}
```
`response`:
```json
{
	"RightBind": {
		"data": {},
		"message": "",
		"status": 201
	}
}
```

#### 用户解除绑定权限

`msgID`: "RightUnBind"



`params`:

arg | type | desc
-- | -- | --
RightID | str | 权限ID，必填

`request`
```
{
	"RightUnBind": {
		"RightID": 3,
	}
}
```
`response`:
```json
{
	"RightUnBind": {
		"data": {},
		"message": "",
		"status": 201
	}
}
```

### 用户查询绑定权限

`msgID`: "BindRightQuery"



`params`:

arg | type | desc
-- | -- | --

`request`
```
{
	"BindRightQuery": {
	}
}
```
`response`:
```json
{
	"BindRightQuery": {
		"data": {
			"data": {
				"UserID": 51,
				"Rights": [{
					"ID": 3,
					"Server": "login",
					"Name": "all",
					"Desc": "登录服务权限",
					"UserRightRelations": null
				}]
			}
		},
		"message": "",
		"status": 200
	}
}
```

### 密码邮箱找回

`msgID`: "EmailRestPwdSend"



`说明1`: LoginName，Mobile，Email至少上报一个，用于确定找回密码的账户

`说明2`: 邮件默认发送到注册邮箱

`params`:

arg | type | desc
-- | -- | --
LoginName | int | 登录名，选填
Mobile | str | 登录手机，选填
Email | str | 登录邮箱，选填
ResetUrl | str | 密码重置url，前端传过来

`request`
```
{
	"EmailRestPwdSend": {
		"LoginName": "admin",
		"ResetUrl": "www.rongshutong.com",
	}
}
```
`response`:
```json
{
	"EmailRestPwdSend": {
		"data": {
			"userID": 62
		},
		"message": "",
		"status": 200
	}
}
```

### 密码邮箱重置

`msgID`: "EmailRestPwdCheck"



`说明1`: LoginName，Mobile，Email至少上报一个，用于确定找回密码的账户

`说明2`: 邮件默认发送到注册邮箱

`params`:

arg | type | desc
-- | -- | --
Code | str | 密码重置验证码
NewPW | str | 新密码

`request`
```
{
	"EmailRestPwdCheck": {
		"Code": "QJRFJ",
		"NewPW": "123",
	}
}
```
`response`:
```json
{
	"EmailRestPwdCheck": {
		"data": {},
		"message": "",
		"status": 200
	}
}
```

### 添加提醒邮箱

`msgID`: "NotifyEmailCreate"



`params`:

arg | type | desc
-- | -- | --
Email | str | 提醒邮箱
Subscribed | str | 是否订阅，订阅后会发送提醒邮件

`request`
```
{
	"NotifyEmailCreate": {
		"Email": "zhaoyi.yuan@bitmain.com",
		"Subscribed": true,
	}
}
```
`response`:
```json
{
	"NotifyEmailCreate": {
		"data": {},
		"message": "",
		"status": 201
	}
}
```

### 查询提醒邮箱

`msgID`: "NotifyEmailQuery"



`params`:

arg | type | desc
-- | -- | --

`request`
```
{
	"NotifyEmailQuery": {
	}
}
```
`response`:
```json
{
	"NotifyEmailQuery": {
		"data": {
			"data": [{
				"ID": 8,
				"CreatedAt": "2018-05-24 14:20:52",
				"UpdatedAt": "2018-05-24 14:20:52",
				"UserID": 62,
				"Email": "zhaoyi.yuan@bitmain.com",
				"Subscribed": true
			}]
		},
		"message": "",
		"status": 201
	}
}
```

### 发送提醒邮件

`msgID`: "NotifyEmailSend"



`params`:

arg | type | desc
-- | -- | --
Subject | str | 邮件主题,必填
ContentType | str | 邮件格式,必填
Content  | str | 邮件内容,必填

`request`
```
{
	"NotifyEmailSend": {
		"Subject": "test",
		"ContentType": "text/html",
		"Content": "我知识来测试一下",
	}
}
```
`response`:
```json
{
	"NotifyEmailSend": {
		"data": {},
		"message": "",
		"status": 201
	}
}
```

### 提醒邮箱订阅

`msgID`: "NotifyEmailQuery"



`说明`: 订阅后的邮箱才会发送提醒邮件

`params`:

arg | type | desc
-- | -- | --
EmailID | int | 邮箱ID, 必填, 查询接口返回ID

`request`
```
{
	"NotifyEmailSub": {
		"EmailID": 8,
	}
}
```
`response`:
```json
{
	"NotifyEmailSub": {
		"data": {},
		"message": "",
		"status": 201
	}
}
```

### 提醒邮箱取消订阅

`msgID`: "NotifyEmailUnSub"



`说明`: 未订后的邮箱不会发送提醒邮件

`params`:

arg | type | desc
-- | -- | --
EmailID | int | 邮箱ID, 必填, 查询接口返回ID

`request`
```
{
	"NotifyEmailUnSub": {
		"EmailID": 8,
	}
}
```
`response`:
```json
{
	"NotifyEmailUnSub": {
		"data": {},
		"message": "",
		"status": 201
	}
}
```

### 提醒邮箱删除

`msgID`: "NotifyEmailDelete"



`params`:

arg | type | desc
-- | -- | --
EmailID | int | 邮箱ID, 必填, 查询接口返回ID

`request`
```
{
	"NotifyEmailDelete": {
		"EmailID": 8,
	}
}
```
`response`:
```json
{
	"NotifyEmailDelete": {
		"data": {},
		"message": null,
		"status": 204
	}
}
```