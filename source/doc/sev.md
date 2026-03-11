# Jaeger 端口说明

| 端口  | 协议 | 用途                      | 使用场景                           |
|-------|------|---------------------------|----------------------------------|
| 5775  | UDP  | 兼容性端口                | 旧客户端                         |
| 5778  | TCP  | 配置服务                  | 采样策略获取                     |
| 6831  | UDP  | Jaeger Thrift Compact     | 主要数据接收                     |
| 6832  | UDP  | Jaeger Thrift Binary      | 备用数据接收                     |
| 9411  | HTTP | Zipkin                    | Zipkin 格式数据                  |
| 14250 | HTTP | OpenTelemetry             | OTLP 协议数据                    |
| 14268 | HTTP | Jaeger Thrift HTTP        | 直接发送到 Collector             |
| 16686 | HTTP | Web UI                    | 用户界面                         |

## 可视化界面 http://localhost:16686/

## Docker 启动命令

```bash
docker run -d --name jaeger \
  -e COLLECTOR_ZIPKIN_HOST_PORT=:9411 \
  -p 5775:5775/udp \
  -p 6831:6831/udp \
  -p 6832:6832/udp \
  -p 5778:5778 \
  -p 16686:16686 \
  -p 14268:14268 \
  -p 14250:14250 \
  -p 9411:9411 \
  jaegertracing/all-in-one:1.29
```