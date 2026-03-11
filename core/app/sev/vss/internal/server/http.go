package server

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	httpHandler "skeyevss/core/app/sev/vss/internal/handler/http"
	interceptor "skeyevss/core/app/sev/vss/internal/interceptor"
	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/pkg/functions"
)

type HttpSev struct {
	svcCtx *types.ServiceContext
}

func NewHttpSev(svcCtx *types.ServiceContext) *HttpSev {
	return &HttpSev{
		svcCtx: svcCtx,
	}
}

func (h *HttpSev) Start() {
	gin.SetMode(gin.ReleaseMode)
	var router = gin.New()
	// if h.svcCtx.Config.Mode != constants.ENV_PRODUCTION {
	// 	router.Use(gin.Logger())
	// }

	router.Use(
		func(c *gin.Context) {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", "*")
			// c.Header("Access-Control-Allow-Headers", "Platform, Sck, MId, DeviceVersion, Version, Language, X-Dev-M, Authorization, Content-Length, X-CSRF-Token, Token,session,X_Requested_With,Accept, Origin, IP, Connection, Accept-Encoding, Accept-Language,DNT, X-CustomHeader, Keep-Alive, User-Agent, X-Requested-With, If-Modified-Since, Cache-Control, Content-Type, Pragma")
			c.Header("Access-Control-Allow-Headers", "*")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type, Expires, Last-Modified, Pragma")

			c.Header("Cache-Control", "private, max-age=10")
			// 缓存请求信息 单位为秒
			c.Header("Access-Control-Max-Age", "172800")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Set("content-type", "application/json")

			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(http.StatusNoContent)
			}

			c.Next()
		},
	)

	{
		var aipGroup = router.Group("/api")
		aipGroup.Use(interceptor.HttpHeader(), interceptor.Timeout(time.Duration(h.svcCtx.Config.Timeout)*time.Millisecond))
		httpHandler.RegisterApiHandlers(h.svcCtx, aipGroup)
	}

	var addr = ":" + strconv.Itoa(h.svcCtx.Config.Http.Port)
	if gin.Mode() == gin.ReleaseMode {
		functions.PrintStyle("green", "web server on ["+addr+"]")
	}
	if err := router.Run(addr); err != nil {
		panic(err)
	}
}
