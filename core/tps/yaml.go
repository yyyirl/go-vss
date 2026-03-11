package tps

import "github.com/zeromicro/go-zero/core/logx"

type (
	YamlFoundation struct {
		Company string `json:",optional"`
		ProductName,
		Version string

		SevBase YamlSevBaseConfig `json:"SevBase"`
	}

	YamlAuth struct {
		AesKey              string `json:",optional"`
		JwtSecret           string `json:",optional"`
		BackendApiJwtSecret string `json:",optional"`
		ConnType            string `json:",optional"`
		LoginRememberExpire int64  `json:",default=172800000,optional"`
		LoginExpire         int64  `json:",default=172800000,optional"`
		TokenType           int64  `json:",default=0,optional"`
	}

	YamlStreamPlayProxyPath struct {
		WS,
		HTTP string
	}

	YamlEmail struct {
		Emails,
		Host,
		Port,
		Username,
		Password string
	}

	YamlRedis struct {
		Host,
		Pass string
		// 集群
		Hosts []string `json:",optional"`

		MaxIdle,
		MaxActive,
		IdleTimeout int
	}

	YamlFFMpeg struct {
		Home               string
		Port               uint
		ContainerizedState bool
	}

	YamlDatabases struct {
		BaseDBName string
		SaveSqlDir string `json:",optional"`
		MysqlBase  string `json:",optional"`
		SqliteBase string `json:",optional"`

		MysqlUsername string `json:",optional"`
		MysqlPassword string `json:",optional"`
		MysqlHost     string `json:",optional"`
		MysqlPort     int    `json:",optional"`
	}

	YamlAccounts struct {
		BackendUsername,
		BackendPassword,
		BackendSuperUsername,
		BackendSuperPassword,
		BackendShowcaseUsername,
		BackendShowcasePassword string
	}

	YamlRedisCluster struct {
		Hosts    []string
		Password string
		DialTimeout,
		ReadTimeout,
		WriteTimeout int
	}

	YamlWechatMina struct {
		Appid,
		Secret string
	}

	YamlWechat struct {
		AppId,
		AppSecret,
		Token,
		OpenAppId,
		OpenAppSecret,
		WebOpenAppId,
		WebOpenAppSecret string
	}

	YamlWechatPay struct {
		MchId,
		SerialNo,
		ApiV3Key,
		PrivateKeyPath,
		// XcxAppid,
		// XcxNotifyUrl,
		MpAppid,
		NotifyUrl string
	}

	YamlAli struct {
		OssUtilCmd,
		AccessId,
		AccessKey,
		EndpointIntranet,
		Endpoint,
		StsEndPoint,
		RegionId,
		Bucket,
		BaseHost,
		Host,
		RegionIdVideo,
		StsAccessId,
		StsAccessKey,
		StsSetRoleArn,
		StsRoleName string

		StsTokenExpire,
		VideoAuthTimeout int64
	}

	YamlAliSms struct {
		AccessKeyId,
		AccessKeySecret,
		RegionId,
		SignName,
		TmpRegister,
		TmpLogin,
		TmpFindPwd,
		TmpBindMobile,
		TmpUnbindMobile,
		TmpBindBankAccount string
	}

	YamlApple struct {
		KeyId,
		TeamId,
		ClientId,
		SigningKeyPath,
		WebClientId,
		ApplePrivateKey,
		SharedSecret string
	}

	YamlCloudKit struct {
		KeyID,
		Container,
		PrivateKeyPath string
	}

	YamlSavePath struct {
		Log   string `json:",optional"`
		Image string
		File  string
		Pdf   string
	}

	YamlTaskTimer []struct {
		Type  string
		Delay int64
	}

	YamlJobs []struct {
		Type   string
		Repeat bool
		Delay  struct {
			Mode  string `json:",options=[month,day,hour,minute,second,10-second,5-second,10-minute,11-second]"`
			Fixed string
		}
	}

	YamlElasticsearch struct {
		Host,
		Port,
		Username,
		Password string
	}

	YamlDingTalk struct {
		AgentId,
		AppKey,
		AppSecret string
	}

	YamlJPush struct {
		Authorization,
		Production string
	}

	YamlAlipay struct {
		Mode,
		ReturnUrl,
		NotifyUrl,
		Appid,
		PrivateKey,
		PublicKey string
	}

	YamlPay struct {
		Alipay YamlAlipay `json:",optional"`
	}

	YamlVector struct {
		ToolHost,
		CharPort string
	}

	YamlBaseConfig struct {
		YamlFoundation

		Log  logx.LogConf
		Mode string `json:",default=pro,options=dev|test|rt|pre|pro"`
		Name string

		MinioApiTarget string `json:",optional"`
		InternalIP,
		ExternalIP string

		ConfigPath YamlConfigPath `json:"ConfigPath,optional"`
		SevRes     YamlSevRes     `json:"SevRes,optional"`
	}

	YamlSevRes struct {
		AssetDir,
		RedisAsset,
		MysqlAsset,
		EtcdAsset,
		FfmpegAsset,
		BackendWebAsset,
		AppSevAsset,
		EtcAsset,

		BackendWebCodePath,
		MediaServerCodePath,
		GoPath,

		AssetScriptsDir,
		// 数据目录
		AssetCertDir,
		AssetDataDir,
		AssetDatalog,
		MysqlData,
		RedisData,
		EtcdData,
		ApiDocDir,

		// 存放依赖服务配置文件
		ConfigDir,
		MysqlConfigPath,
		RedisConfigPath,
		EtcdConfigPath string

		DBBaseName string

		UseEtcd,
		UseMysql,
		UseRedis,
		UseFfmpeg bool

		DatabaseType string `json:",default=mysql,options=sqlite|mysql"`
	}

	YamlSSLKey struct {
		PublicKey,
		PrivateKey string
	}

	YamlSevBaseConfig struct {
		Root,

		MysqlUsername,
		MysqlPassword,
		RedisPassword,

		SevNameMysql,
		SevNameRedis,
		SevNameEtcd,
		SevNameGuard,
		SevNameMediaServer,
		SevNameVss,
		SevNameCron,
		SevNameDB,
		SevNameBackendApi,
		SevNameWebSev string

		EtcdHost,
		MysqlHost,
		RedisHost string

		MysqlPort,
		RedisPort,
		EtcdPort,
		WebSevPort,
		MediaServerPort,
		VssPort,
		CronPort,
		GuardPort,
		DBPort,
		BackendApiPort,
		MediaServerHttpsPort,
		MediaServerRtspPort,
		MediaServerRtmpPort int
		MediaRtcIceHostNatToIps string `json:",optional"`

		ProxyApiExternalTarget,
		ProxyApiExternal,
		ProxyFile,
		ProxyFileUrl,
		ProxyApiBase string

		MSNotifyOnPubStart,
		MSNotifyOnPubStop,
		MSNotifyOnPushStart,
		MSNotifyOnPushStop,
		MSNotifyOnRelayPullStart,
		MSNotifyOnRelayPullStop,
		MSNotifyOnRtmpConnect,
		MSNotifyOnSubStart string

		SSL  YamlSSLKey
		Keys YamlUKey

		UseEtcd,
		UseMysql,
		UseRedis bool

		DatabaseType string `json:",default=mysql,options=sqlite|mysql"`
	}

	YamlUKey struct {
		MediaServer,
		Vss,
		Cron,
		DB,
		BackendApi,
		WebSev string
	}

	YamlConfigPath struct {
		GuardConf,
		WebSevConf,
		MediaServerGrpcConf,
		VssConf,
		CronConf,
		DbGrpcConf,
		BackendApiConf string
	}

	YamlSip struct {
		UseExternalIP bool // 是否使用外网ip
		// 国标级联
		CascadeSipPort,
		// 服务监听的 tcp/udp 端口号
		Port int
		// gb/t28181 20 位国标 ID
		ID string
		// 域
		Domain string
		// 注册密码
		Password string
		// 使用密码校验
		UsePassword bool
		// catalog定时器
		CatalogInterval,
		HeartbeatTimeout,
		MediaReceiveStreamTimeout int64
		MediaNoWatchingTimeout int64
		UseExternalWan,
		MediaServerVssSameMachine bool

		Expire      uint32
		SendTimeout int

		MediaServerStreamPortMax,
		MediaServerStreamPortMin uint

		// 服务器可用端口范围
		UsableMinPort,
		UsableMaxPort uint
	}

	YamlOnvif struct {
		MulticastIP      string
		WsDiscoveryPort  uint
		DiscoveryTimeout uint
	}

	// rpc拦截器
	YamlRpcInterceptorConf struct {
		UseRpcCallerRetry  bool // 是否开启重试
		RpcCallerRetryMax, // 重试次数
		RpcCallerRetryWaitInterval uint // 重试等待时间 单位/毫秒

		UseRpcKeepalive                 bool // 是否启用keepalive
		RpcKeepaliveTime                uint // 发送 keepalive 探测的时间间隔 单位/s
		RpcKeepaliveTimeout             uint // 等待响应超时时间 单位/s
		RpcKeepalivePermitWithoutStream bool // 即使没有活跃的流也发送 keepalive
	}

	YamlPProf struct {
		BackendApiPort,
		DbRpcPort,
		VssPort,
		WebPort,
		CronPort,
		MediaServerPort uint

		BackendApiName,
		DbRpcName,
		VssName,
		WebName,
		CronName,
		MediaServerName string
	}

	YamlGenUniqueId struct {
		Platform,
		Dir,
		Nvr,
		Camera string
	}
)
