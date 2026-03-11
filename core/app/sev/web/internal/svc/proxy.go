// @Title        web代理服务器
// @Description  proxy
// @Create       yirl 2025/3/18 16:36

package svc

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"

	"skeyevss/core/app/sev/web/internal/config"
	"skeyevss/core/constants"
	"skeyevss/core/pkg/functions"
)

type Proxy struct {
	conf *config.Config

	certPem,
	certKey,
	webStaticDir string
}

const (
	proxyPath         = "proxyPath"
	proxyExternalPath = "proxyExternalPath"
)

func NewProxy(conf *config.Config, webStaticDir, certPem, certKey string) *Proxy {
	return &Proxy{
		conf: conf,

		certPem:      certPem,
		certKey:      certKey,
		webStaticDir: webStaticDir,
	}
}

func (p *Proxy) Start() {
	var ginMode = gin.DebugMode
	if p.conf.Mode == constants.ENV_PRODUCTION {
		ginMode = gin.ReleaseMode
	}

	gin.SetMode(ginMode)
	var router = gin.New()
	if p.conf.Mode != constants.ENV_PRODUCTION {
		router.Use(gin.Logger())
	}

	router.Use(
		func(c *gin.Context) {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", "*")
			// c.Header("Access-Control-Allow-Headers", "Platform, Sck, MId, DeviceVersion, Version, Language, X-Dev-M, Authorization, Content-Length, X-CSRF-Token, Token,session,X_Requested_With,Accept, Origin, IP, Connection, Accept-Encoding, Accept-Language,DNT, X-CustomHeader, Keep-Alive, User-Agent, X-Requested-With, If-Modified-Since, Cache-Control, Content-Type, Pragma")
			c.Header("Access-Control-Allow-Headers", "*")
			// 跨域关键设置 让浏览器可以解析
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type, Expires, Last-Modified, Pragma")

			c.Header("Cache-Control", "private, max-age=10")
			// 缓存请求信息 单位为秒
			c.Header("Access-Control-Max-Age", "172800")
			//  跨域请求是否需要带cookie信息 默认设置为true
			c.Header("Access-Control-Allow-Credentials", "true")
			// 设置返回格式是json
			c.Set("content-type", "application/json")

			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(http.StatusNoContent)
			}

			c.Next()
		},
	)

	router.Any(p.conf.SevBase.ProxyApiBase+"/*"+proxyPath, func(context *gin.Context) {
		p.apiBackendProxyHandler(context)
	})

	router.Any(p.conf.SevBase.ProxyApiExternal+"/*"+proxyExternalPath, func(context *gin.Context) {
		p.apiExternalProxyHandler(context)
	})

	// 静态文件请求
	router.Use(static.Serve("/", static.LocalFile(p.webStaticDir, true)))
	// 文件上传 文件
	router.StaticFS("/x-assets/source", http.Dir("source"))

	router.NoRoute(func(c *gin.Context) {
		if c.Request.Method == http.MethodGet {
			c.File(filepath.Join(p.webStaticDir, "index.html"))
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		}
	})

	// https
	if functions.FileExists(p.certKey) && functions.FileExists(p.certPem) {
		if err := router.RunTLS(":"+strconv.Itoa(p.conf.SevBase.WebSevPort), p.certPem, p.certKey); err != nil {
			panic(err)
		}
		return
	}

	var addr = ":" + strconv.Itoa(p.conf.SevBase.WebSevPort)
	if gin.Mode() == gin.ReleaseMode {
		functions.PrintStyle("green", "web server on ["+addr+"]")
	}
	// http
	if err := router.Run(addr); err != nil {
		panic(err)
	}
}

func (p *Proxy) apiBackendProxyHandler(c *gin.Context) {
	(&httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = "http"
			req.URL.Host = p.conf.InternalIP + ":" + strconv.Itoa(p.conf.SevBase.BackendApiPort)
			req.URL.Path = c.Param(proxyPath)
			req.Method = c.Request.Method
			req.Header = c.Request.Header
			req.Host = p.conf.InternalIP + ":" + strconv.Itoa(p.conf.SevBase.WebSevPort)
		},
	}).ServeHTTP(c.Writer, c.Request)
}

func (p *Proxy) apiExternalProxyHandler(c *gin.Context) {
	var original = strings.Trim(c.Param(proxyExternalPath), "/")
	if original == "" {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": "No proxy path",
		})
		return
	}

	res, err := url.Parse(original)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": "url parse error",
			"error":   err,
		})
		return
	}

	(&httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = res.Scheme
			req.URL.Host = res.Host
			req.URL.Path = res.Path
			req.Method = c.Request.Method
			req.Header = c.Request.Header
			req.Host = p.conf.InternalIP + ":" + strconv.Itoa(p.conf.SevBase.WebSevPort)
		},
	}).ServeHTTP(c.Writer, c.Request)
}
