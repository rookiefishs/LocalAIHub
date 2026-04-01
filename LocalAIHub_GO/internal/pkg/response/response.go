package response

import (
	"encoding/json"
	"net/http"
	"strings"

	"localaihub/localaihub_go/internal/pkg/appctx"
)

func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func AdminSuccess(w http.ResponseWriter, r *http.Request, data any) {
	JSON(w, http.StatusOK, map[string]any{
		"code":       0,
		"message":    "ok",
		"data":       data,
		"request_id": appctx.RequestID(r.Context()),
	})
}

func AdminError(w http.ResponseWriter, r *http.Request, status int, code int, message string) {
	translated := TranslateError(message)
	JSON(w, status, map[string]any{
		"code":       code,
		"message":    translated,
		"data":       nil,
		"request_id": appctx.RequestID(r.Context()),
	})
}

func TranslateError(errMsg string) string {
	translations := map[string]string{
		"invalid parameters":                "参数错误",
		"invalid provider id":               "无效的上游ID",
		"invalid provider key id":           "无效的Key ID",
		"invalid model id":                  "无效的模型ID",
		"invalid binding id":                "无效的绑定ID",
		"invalid client key id":             "无效的客户端Key ID",
		"invalid request log id":            "无效的日志ID",
		"invalid lock_until":                "无效的锁定时间",
		"invalid username or password":      "用户名或密码错误",
		"unauthorized":                      "未授权",
		"list models failed":                "获取模型列表失败",
		"list bindings failed":              "获取绑定列表失败",
		"create model failed":               "创建模型失败",
		"update model failed":               "更新模型失败",
		"delete model failed":               "删除模型失败",
		"create binding failed":             "创建绑定失败",
		"update binding failed":             "更新绑定失败",
		"delete binding failed":             "删除绑定失败",
		"list providers failed":             "获取上游列表失败",
		"create provider failed":            "创建上游失败",
		"update provider failed":            "更新上游失败",
		"delete provider failed":            "删除上游失败",
		"update provider status failed":     "更新上游状态失败",
		"test provider connection failed":   "测试上游连接失败",
		"get provider failed":               "获取上游信息失败",
		"provider not found":                "上游不存在",
		"list provider keys failed":         "获取上游Key列表失败",
		"create provider key failed":        "创建上游Key失败",
		"update provider key status failed": "更新上游Key状态失败",
		"delete provider key failed":        "删除上游Key失败",
		"list client keys failed":           "获取客户端Key列表失败",
		"get client key failed":             "获取客户端Key失败",
		"create client key failed":          "创建客户端Key失败",
		"update client key failed":          "更新客户端Key失败",
		"delete client key failed":          "删除客户端Key失败",
		"update client key status failed":   "更新客户端Key状态失败",
		"list routes failed":                "获取路由列表失败",
		"get route failed":                  "获取路由失败",
		"route not found":                   "路由不存在",
		"switch route failed":               "切换路由失败",
		"unlock route failed":               "解锁路由失败",
		"list request logs failed":          "获取请求日志失败",
		"get request log failed":            "获取请求日志详情失败",
		"request log not found":             "请求日志不存在",
		"list audit logs failed":            "获取审计日志失败",
		"model not found":                   "模型不存在",
		"client key not found":              "客户端Key不存在",
		"binding not found":                 "绑定不存在",
		"invalid request body":              "无效的请求体",
	}

	lower := strings.ToLower(errMsg)
	if trans, ok := translations[lower]; ok {
		return trans
	}
	return errMsg
}
