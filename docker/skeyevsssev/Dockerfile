#FROM alpine
FROM alpine:latest

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk --no-cache add ffmpeg bash ca-certificates tzdata openssh-client
# x264、x265
# RUN apk add --no-cache ffmpeg bash ca-certificates ffmpeg4-libs ffmpeg4-dev

RUN bash --version && \
    ssh -V && \
    ffmpeg -version | head -1

ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

WORKDIR /app
RUN ffmpeg -version

# 功能缺失使用以下镜像
#FROM debian:stable-slim
#RUN apt-get update && apt-get install -y --no-install-recommends ffmpeg && rm -rf /var/lib/apt/lists/*
#WORKDIR /app
#RUN ffmpeg -version

# libc
#FROM ubuntu:22.04
#RUN apt update && apt install -y ca-certificates
