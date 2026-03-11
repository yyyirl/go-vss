#!/bin/bash

source ./constants.sh

cd $SERVER_PATH
$GO mod tidy && $GO mod vendor
exitState "go mod 执行失败"
