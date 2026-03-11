# 常见问题解答

### 服务相关

**Q: linux mysql启动 报错**

A: mysql 报错 mysqld: error while loading shared libraries: libaio.so.1: cannot open shared object file: No such file or directory
```shell
# arm64
sudo ln -sf /lib/aarch64-linux-gnu/libaio.so.1t64 /lib/aarch64-linux-gnu/libaio.so.1
# amd64
sudo ln -s /usr/lib/x86_64-linux-gnu/libaio.so.1 /usr/lib/libaio.so.1
```

**Q: 启动日志查看**

A: linux
```shell
journalctl -u SkeyevssSevGuard -f
```

**Q: 通道快照图片无法查看**

A: 检查环境变量文件中 `SKEYEVSS_INTERNAL_IP`(内网ip) `SKEYEVSS_EXTERNAL_IP`(外网IP) `SKEYEVSS_WEB_SEV_PROXY_FILE_URL`(文件代理地址) 是否配置正确(参考网页 配置中心->服务器配置)



**Q: 公网WEBRTC无法播放**

A: 检查环境变量文件中 `SKEYEVSS_MEDIA_RTC_ICE_HOST_NAT_TO_IPS`是否有配置公网ip(rtc数据发送绑定ip 多个地址使用, 分隔)

---

### 仍未找到答案？

如果您有其他问题，欢迎随时联系我们的技术团队。
*   **邮箱：** [295222688@qq.com,1003275805@qq.com]
*   **qq群：** [102644504]
*   **服务时间：** 周一至周五 9:00 - 18:00