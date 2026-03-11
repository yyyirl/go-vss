#!/bin/bash

# 模板初始化生成
source ./constants.sh

mkdir -p $SERVER_TEMPLATE_PATH
goctl template init --home $SERVER_TEMPLATE_PATH