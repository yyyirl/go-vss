#!/bin/bash

# api生成
source ./constants.sh

# 服务名称 api项目
server_name="backend" # TODO

# 模块名称 {{.ModuleName}} TODO
service_module_name="cascade"
# 模块单数 {{.ServiceModuleNameSingular}} TODO
service_module_name_singular="Cascade"
# 模块复数 {{.ServiceModuleNamePlural}} TODO
service_module_name_plural="Cascade"
# rpc service {{.ServiceClient}} TODO
service_client="Device"
# rpc service name {{.ServiceName}} TODO
service_name="deviceservice"
# {{.LogType}} log type
log_type="Cascade"


if [[ ! -n "$server_name" ]]; then
    exitPrintln "项目名称不能为空"
    exit 1
fi

work_path=${SERVER_REST_PATH}/${server_name}/internal/logic/${service_module_name_plural}
mkdir -p $work_path

cd "${work_path}"
exitState "${work_path} 路径不存在"

\cp $SERVER_API_LOGIC_TEMPLATE_CUSTOM_PATH/createlogic.go.tpl $work_path/createlogic.go
\cp $SERVER_API_LOGIC_TEMPLATE_CUSTOM_PATH/deletelogic.go.tpl $work_path/deletelogic.go
\cp $SERVER_API_LOGIC_TEMPLATE_CUSTOM_PATH/updatelogic.go.tpl $work_path/updatelogic.go
\cp $SERVER_API_LOGIC_TEMPLATE_CUSTOM_PATH/listlogic.go.tpl $work_path/listlogic.go
\cp $SERVER_API_LOGIC_TEMPLATE_CUSTOM_PATH/rowlogic.go.tpl $work_path/rowlogic.go

cd $work_path
ls -1 | while read item; do
    if [ ! -d "$item" ]; then
        if [ "$(uname)" == "Darwin" ]; then
            sed -i '' "s|{{.ServiceModuleNameSingular}}|${service_module_name_singular}|g" $item
            sed -i '' "s|{{.ServiceModuleNamePlural}}|${service_module_name_plural}|g" $item
            sed -i '' "s|{{.ServiceClient}}|${service_client}|g" $item
            sed -i '' "s|{{.ServiceName}}|${service_name}|g" $item
            sed -i '' "s|{{.ModuleName}}|${service_module_name}|g" $item
            sed -i '' "s|{{.LogType}}|${log_type}|g" $item
        else
            sed -i "s|{{.ServiceModuleNameSingular}}|${service_module_name_singular}|g" $item
            sed -i "s|{{.ServiceModuleNamePlural}}|${service_module_name_plural}|g" $item
            sed -i "s|{{.ServiceClient}}|${service_client}|g" $item
            sed -i "s|{{.ServiceName}}|${service_name}|g" $item
            sed -i "s|{{.ModuleName}}|${service_module_name}|g" $item
            sed -i "s|{{.LogType}}|${log_type}|g" $item
        fi
    fi
done

