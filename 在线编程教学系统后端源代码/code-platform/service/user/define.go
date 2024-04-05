package user

import (
	"time"

	"code-platform/service/define"
)

type (
	PageResponse = define.PageResponse
	PageInfo     = define.PageInfo
)

const (
	VerificationCodeKeyPrefix = "vckp:%s"

	// VerificationCodeExpired 验证码5分钟后过期
	VerificationCodeExpired = time.Minute * 5
)

const (
	accessTokenDuration  = 120 * time.Minute
	refreshTokenDuration = 3 * 24 * time.Hour
)

const (
	// 登录凭证长度，36
	TokenLength = 36

	// 刷新 token 长度
	RefreshTokenLength = 28
)

type (
	CodingTimeInfo = define.CodingTimeInfo
	OuterUser      = define.OuterUser
)

var (
	ToOuterUser      = define.ToOuterUser
	BatchToOuterUser = define.BatchToOuterUser
)

type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	Role         uint16 `json:"role"`
}

const WelcomeHtmlTemplate = `<div style="background-color:#ECECEC; padding: 35px;">
    <table cellpadding="0" align="center"
        style="width: 600px; margin: 0px auto; text-align: left; position: relative; border-top-left-radius: 5px; border-top-right-radius: 5px; border-bottom-right-radius: 5px; border-bottom-left-radius: 5px; font-size: 14px; font-family:微软雅黑, 黑体; line-height: 1.5; box-shadow: rgb(153, 153, 153) 0px 0px 5px; border-collapse: collapse; background-position: initial initial; background-repeat: initial initial;background:#fff;">
        <tbody>
            <tr>
                <th valign="middle"
                    style="height: 25px; line-height: 25px; padding: 15px 35px; border-bottom-width: 1px; border-bottom-style: solid; border-bottom-color: #42a3d3; background-color: #49bcff; border-top-left-radius: 5px; border-top-right-radius: 5px; border-bottom-right-radius: 0px; border-bottom-left-radius: 0px;">
                    <font face="微软雅黑" size="5" style="color: rgb(255, 255, 255); ">更改你的密码</font>
                </th>
            </tr>
            <tr>
                <td>
                    <div style="padding:25px 35px 40px; background-color:#fff;">
                        <h2 style="margin: 5px 0px; ">
                            <font color="#333333" style="line-height: 20px; ">
                                <font style="line-height: 22px; " size="4">
                                    亲爱的同学</font>
                            </font>
                        </h2>
                        <p>首先感谢您使用本站！这是您的验证码：{{.Code}}<br>
                            <p align="right">{{.Layout}}</p>
                            <div style="width:700px;margin:0 auto;">
                                <div
                                    style="padding:10px 10px 0;border-top:1px solid #ccc;color:#747474;margin-bottom:20px;line-height:1.3em;font-size:12px;">
                                    <p>此为系统邮件，请勿回复<br>
                                        请保管好您的邮箱，避免账号被他人盗用
                                    </p>
                                    <p>©***</p>
                                </div>
                            </div>
                    </div>
                </td>
            </tr>
        </tbody>
    </table>
</div>`
