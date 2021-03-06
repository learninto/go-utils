package twirp_hook

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/learninto/goutil/conf"
	"github.com/learninto/goutil/ctxkit"
	"github.com/learninto/goutil/jwt"
	"github.com/learninto/goutil/log"
	"github.com/learninto/goutil/redis"
	"github.com/learninto/goutil/twirp"
	"github.com/learninto/goutil/xhttp"
)

func NewInternalHeaders() *twirp.ServerHooks {
	return &twirp.ServerHooks{
		RequestRouted: func(ctx context.Context) (context.Context, error) {
			req, ok := twirp.HttpRequest(ctx)
			if !ok {
				return ctx, nil
			}
			sign := req.Header.Get("Sign")

			ctx = ctxkit.WithSignKey(ctx, sign)                        // 注入签名
			ctx = ctxkit.WithDevice(ctx, req.Header.Get("Device"))     // 注入 用户设备  iso、android、web
			ctx = ctxkit.WithMobiApp(ctx, req.Header.Get("MobiApp"))   // 注入 APP 标识
			ctx = ctxkit.WithVersion(ctx, req.Header.Get("Version"))   // 注入 版本 标识
			ctx = ctxkit.WithPlatform(ctx, req.Header.Get("Platform")) // 注入 平台 标识
			ctx = ctxkit.WithUserIP(ctx, req.RemoteAddr)               // TODO 注入 客户端IP 标识  目前貌似不准确待测试

			/* ------ 用户信息 ------ */
			c, err := jwt.NewJWT().ParseToken(sign)
			if err != nil {
				return ctx, nil
			}
			var userId int64
			_ = json.Unmarshal(c.Data, &userId)
			ctx = ctxkit.WithUserID(ctx, userId) // 注入用户id

			return ctx, nil
		},
	}
}

// NewHeaders
func NewHeaders() *twirp.ServerHooks {
	type User struct {
		// Comment: 企业id
		CompanyID int64 `json:"company_id"`
		// Comment: 唯一标识
		ID int64 `json:"id"`
		// Comment：部门id
		DepartmentID int64 `json:"department_id"`
		// Comment: 角色id数组 英文逗号隔开
		PartIds string `json:"part_ids"`
		// Comment: 部门id数组 英文逗号隔开
		DepartmentIds string `json:"department_ids"`
		// Comment: 用户昵称
		NickName string `json:"nick_name"`
		// Comment: 用户登录账号
		UserName string `json:"user_name"`
		// Comment: 权限编码  多个，隔开
		RolesCodes string `json:"roles_codes"`
		// Comment: 是否生效。0：未生效；100：已生效
		// Default: 100
		Status int8 `json:"status"`
		// Comment: 公司
		Company struct {
			// Comment: 是否生效。0：未生效；100：已生效
			// Default: 100
			Status int `json:"status"`
			// Comment: 过期时间
			ExpiryTime int64 `json:"expiry_time"`
		} `json:"company"`
	}
	return &twirp.ServerHooks{
		RequestRouted: func(ctx context.Context) (context.Context, error) {
			req, ok := twirp.HttpRequest(ctx)
			if !ok {
				return ctx, nil
			}
			sign := req.Header.Get("Sign")

			ctx = ctxkit.WithSignKey(ctx, sign)                        // 注入签名
			ctx = ctxkit.WithDevice(ctx, req.Header.Get("Device"))     // 注入 用户设备  iso、android、web
			ctx = ctxkit.WithMobiApp(ctx, req.Header.Get("MobiApp"))   // 注入 APP 标识
			ctx = ctxkit.WithVersion(ctx, req.Header.Get("Version"))   // 注入 版本 标识
			ctx = ctxkit.WithPlatform(ctx, req.Header.Get("Platform")) // 注入 平台 标识
			ctx = ctxkit.WithUserIP(ctx, req.RemoteAddr)               // TODO 注入 客户端IP 标识  目前貌似不准确待测试

			///* ------ 用户信息 ------ */
			//c, err := jwt.NewJWT().ParseToken(sign)
			//if err != nil {
			//	return ctx, nil
			//}
			//u := User{}
			//_ = json.Unmarshal(c.Data, &u.ID)
			//ctx = ctxkit.WithUserID(ctx, u.ID) // 注入用户id

			resp, err := queryUserInfo(ctx, sign)
			if err != nil {
				return ctx, twirp.NewError(twirp.Unauthenticated, err.Error())
			}

			u := User{}
			_ = json.Unmarshal(resp, &u)
			if u.ID == 0 {
				return ctx, twirp.NewError(twirp.Unauthenticated, "请先登录")
			}
			if u.Status != 100 {
				return ctx, twirp.NewError(twirp.Unauthenticated, "抱歉您的账号已经被禁用")
			}
			if u.Company.Status != 100 {
				return ctx, twirp.NewError(twirp.Unauthenticated, "抱歉您所在的企业已经被禁用")
			}
			if u.Company.ExpiryTime <= time.Now().Unix() {
				return ctx, twirp.NewError(twirp.Unauthenticated, "抱歉您所在的企业已经过期了")
			}

			ctx = ctxkit.WithUserID(ctx, u.ID)                   // 注入用户id
			ctx = ctxkit.WithUserName(ctx, u.UserName)           // 注入用户登录账号
			ctx = ctxkit.WithNickName(ctx, u.NickName)           // 注入用户昵称
			ctx = ctxkit.WithCompanyID(ctx, u.CompanyID)         // 注入公司id
			ctx = ctxkit.WithDepartmentID(ctx, u.DepartmentID)   // 注入管辖部门id
			ctx = ctxkit.WithPartIds(ctx, u.PartIds)             // 注入角色id
			ctx = ctxkit.WithDepartmentIds(ctx, u.DepartmentIds) // 注入部门id
			ctx = ctxkit.WithRolesCodes(ctx, u.RolesCodes)       // 注入权限编码

			return ctx, nil
		},
	}
}

func queryUserInfo(ctx context.Context, sign string) (b []byte, err error) {
	userBody, _ := redis.Get(ctx, "default").Get(ctx, sign)
	if userBody != nil && userBody.Value != nil {
		return userBody.Value, nil
	}

	timeout := 2 * time.Second
	urlStr := conf.Get("FRAME_ADDR") + conf.Get("FRAME_REFRESH_USER_URI")
	req, _ := http.NewRequest(http.MethodPost, urlStr, bytes.NewReader([]byte("")))
	req.Header.Set("SIGN", sign)

	resp, err := xhttp.NewClient(timeout).Do(ctx, req)
	if err != nil {
		log.Get(ctx).Error("请求FRAME_REFRESH_USER_URI失败：", err)
		return b, nil
	}
	if resp.StatusCode != 200 {
		return
	}

	userBody, _ = redis.Get(ctx, "default").Get(ctx, sign)
	return userBody.Value, nil
}
