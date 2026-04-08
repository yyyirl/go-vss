#!/bin/bash

source ./constants.sh

model_singular_name=$(get_specific_parameter "-name" "$@")
model_plural_name=$(get_specific_parameter "-names" "$@")
backend_permissions=$(get_specific_parameter "-backend-permissions" "$@")
frontend_permissions=$(get_specific_parameter "-frontend-permissions" "$@")
model_path=$(get_specific_parameter "-model-path" "$@")

if [ -z "$model_singular_name" ]; then
    echo "-name 不能为空"
    exit 1
fi

if [ -z "$model_plural_name" ]; then
    echo "-names 不能为空"
    exit 1
fi

if [ -z "$backend_permissions" ]; then
    echo "-backend-permissions 不能为空"
    exit 1
fi

if [ -z "$frontend_permissions" ]; then
    echo "-frontend-permissions 不能为空"
    exit 1
fi

if [ -z "$model_path" ]; then
    echo "-model-path 不能为空"
    exit 1
fi

# 获取结构体名称（首字母大写）
struct_name=$(echo "$model_plural_name" | awk '{print toupper(substr($0,1,1)) substr($0,2)}')
ts_class_name="${struct_name}Item"

# 读取 type 结构体到变量
struct_content=$(awk "/^type $struct_name struct {/,/^}/" "$model_path/model.go")

# 临时文件存储字段
fields_file=$(mktemp)

# 逐行读取结构体内容
echo "$struct_content" | while IFS= read -r line; do
    # 跳过空行
    [ -z "$line" ] && continue

    # 跳过 struct 定义行和结束行
    echo "$line" | grep -q "type $struct_name struct {" && continue
    echo "$line" | grep -q "^}" && continue

    # 跳过包含 *orm.DefaultModel 的行
    echo "$line" | grep -q "DefaultModel" && continue

    # 提取字段行（以空格或tab开头，包含大写字母）
    if echo "$line" | grep -qE '^[[:space:]]+[A-Z]'; then
        # 提取字段名（第一个单词）
        field_name=$(echo "$line" | awk '{print $1}')

        # 提取类型（第二个单词，去掉指针符号）
        go_type=$(echo "$line" | awk '{print $2}' | sed 's/^\*//')

        # 提取 json tag
        json_name=""
        if echo "$line" | grep -q 'json:'; then
            json_name=$(echo "$line" | sed -n 's/.*json:"\([^"]*\)".*/\1/p')
        fi

        # 如果 json tag 是 "-"，跳过
        if [ "$json_name" = "-" ]; then
            continue
        fi

        # 如果没有 json tag，使用字段名并转换为驼峰
        if [ -z "$json_name" ]; then
            json_name=$(echo "$field_name" | awk '{print tolower(substr($0,1,1)) substr($0,2)}')
        fi

        # 转换为驼峰命名
        camel_name=$(echo "$json_name" | awk '{print tolower(substr($0,1,1)) substr($0,2)}')

        # 确定 TS 类型和默认值
        case "$go_type" in
            *uint64*|*int64*|*uint32*|*int32*|*uint*|*int*|*float64*|*float32*)
                ts_type="number"
                default_value="0"
                ;;
            *string*)
                ts_type="string"
                default_value='""'
                ;;
            *bool*)
                ts_type="boolean"
                default_value="false"
                ;;
            *\[\]*)
                ts_type="any[]"
                default_value="[]"
                ;;
            *map*)
                ts_type="Record<string, any>"
                default_value="{}"
                ;;
            *)
                ts_type="any"
                default_value="null"
                ;;
        esac

        # 写入临时文件
        echo "$camel_name|$ts_type|$default_value" >> "$fields_file"
    fi
done

# 初始化 TS 类字符串变量
ts_class_string=""
if [ -s "$fields_file" ]; then
    while IFS='|' read -r camel_name ts_type default_value; do
        ts_class_string+="    public $camel_name: $ts_type\n"
    done < "$fields_file"
fi

ts_class_string+="\n"
ts_class_string+="    constructor(data: Partial<Item>) {\n"

if [ -s "$fields_file" ]; then
    while IFS='|' read -r camel_name ts_type default_value; do
        ts_class_string+="        this.$camel_name = data?.$camel_name ?? $default_value\n"
    done < "$fields_file"
fi
ts_class_string+="    }"

# 清理临时文件
rm -f "$fields_file"

work_path=${MAIN_PATH}/tmp/tsx/${model_plural_name}
rm -rf $work_path
mkdir -p $work_path

\cp $TSX_TEMPLATE_CUSTOM_PATH/model.tsx.tpl $work_path/model.tsx
\cp $TSX_TEMPLATE_CUSTOM_PATH/index.tsx.tpl $work_path/index.tsx
\cp $TSX_TEMPLATE_CUSTOM_PATH/api.ts.tpl $work_path/api.ts

cd $work_path
ls -1 | while read item; do
    if [ ! -d "$item" ]; then
        if [ "$(uname)" == "Darwin" ]; then
            sed -i '' "s|{{.BackendPermissions}}|${backend_permissions}|g" $item
            sed -i '' "s|{{.FrontendPermissions}}|${frontend_permissions}|g" $item
            sed -i '' "s|{{.PluralName}}|${model_plural_name}|g" $item
            sed -i '' "s|{{.SingularName}}|${model_singular_name}|g" $item
            sed -i '' "s|{{.Model}}|${ts_class_string}|g" $item
        else
            sed -i "s|{{.BackendPermissions}}|${backend_permissions}|g" $item
            sed -i "s|{{.FrontendPermissions}}|${frontend_permissions}|g" $item
            sed -i "s|{{.PluralName}}|${model_plural_name}|g" $item
            sed -i "s|{{.SingularName}}|${model_singular_name}|g" $item
            sed -i "s|{{.Model}}|${ts_class_string}|g" $item
        fi
    fi
done

echo "✅ 前端代码目录: $work_path"
