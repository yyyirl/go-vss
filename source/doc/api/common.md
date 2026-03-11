# 通用

## Request Header参数

| 字段名                             | 字段注释         | 字段类型   | 字段值               |
|------------------------------------|------------------|------------|---------------------|
| Authorization | token鉴权        | string     | xxxxx              |
| refreshToken  | 是否刷新token    | number     | 1 \| 0             |
| Content-Type  | 内容类型         | string     | application/json   |

---
## Response

```shell
{
    "timestamp": 1762393485500,
    "node": "0-1",
    "version": "V1.0.2",
    "data": {},
    "cc": 0.011,
    "license": "授权码与本机序列号不匹配",
    "reset-pwd": false,
    "logout": false,
    "message": "等待设备注册",
    "token": "I9cAW9dk5vKUbXkfmnlwM9TfMMHaOdnK4XM8vyVzN1EVCkKG/D10B1IQ7gTGg6H6iCNCm+EyM2UN1VmuoU7GhzdLP8rGvPn7/FmAKOLSLyV8oP0LEHS7lbh+Kr9zrzmCxVsW1dOpFzaE7AfcWOBCya9gO0LlAsPA6ZQWV9dtPooXfbk/70nFcYwk93wYBZYu",
    "code": 10000
}
```

| 字段名                                  | 字段注释                                   | 字段类型   |
|----------------------------------------|--------------------------------------------|------------|
| timestamp   | 当前请求服务器时间 单位/毫秒                | number     |
| node        | 当前请求服务器所在节点                     | string     |
| version     | 服务版本                                   | string     |
| data        | 请求耗时 单位/秒                           | object     |
| cc          | 授权信息(当授权错误时返回, 默认为空)        | number     |
| license     | 是否需要更新当前密码                       | string     |
| reset-pwd   | 是否需要跳转登录                           | boolean    |
| logout      | 请求消息, 大部分时间此字段是给予http code非200时候使用 | boolean    |
| message     | token续期, 如果此字段不为空需要替换本地token缓存 | string     |
| token       | 自定义code                                | string     |
| code        | 业务数据                                   | number     |


> **注意事项：**
> - 接口请求失败后错误码统一处理http code (400, 401, 403)
> - 大部分场景在请求失败后response body不会返回错误码
> - 以上字段如果说没有被填充那么将不会在response body显示
---

## 通用请求参数数据结构

### ReqParams 结构体
| 字段名 | 字段注释 | 字段类型 | 可选性 |
|--------|----------|----------|--------|
| orders | 排序规则 | []*orderItem | 可选 |
| conditions | 查询条件 | []*conditionItem | 可选 |
| uniqueIds | 唯一ID列表 | []string | 可选 |
| limit | 分页大小 | int | 可选 |
| page | 页码 | int | 可选 |
| keyword | 关键词搜索 | string | 可选 |
| uniqueId | 唯一ID | string | 可选 |
| all | 是否查询全部 | bool | 可选 |
| type | 类型 | int64 | 可选 |
| ignoreNotFound | 是否忽略未找到 | bool | 可选 |
| ignoreUpdateColumns | 忽略更新列 | []string | 可选 |
| data | 更新数据 | []*updateItem | 可选 |
| bulkUpdates | 批量更新数据 | []*bulkUpdateItem | 可选 |
| backend | 是否后端请求 | bool | 可选 |

### orderItem 排序项
| 字段名 | 字段注释 | 字段类型 |
|--------|----------|----------|
| column | 排序列名 | string |
| value | 排序方式 | orderType |

### conditionItem 条件项
| 字段名 | 字段注释 | 字段类型 | 可选性 |
|--------|----------|----------|--------|
| column | 条件列名 | string | 必选 |
| value | 条件值 | interface{} | 可选 |
| values | 条件值列表 | []interface{} | 可选 |
| operator | 操作符 | string | 可选 |
| useNil | 是否使用空值 | bool | 内部 |
| original | 原始条件 | *conditionOriginalItem | 可选 |
| logicalOperator | 逻辑操作符 | string | 可选 |
| inner | 内部条件 | []*conditionItem | 可选 |

### conditionOriginalItem 原始条件
| 字段名 | 字段注释 | 字段类型 |
|--------|----------|----------|
| query | 查询语句 | string |
| values | 参数值 | []interface{} |

### updateItem 更新项
| 字段名 | 字段注释 | 字段类型 |
|--------|----------|----------|
| column | 更新列名 | string |
| value | 更新值 | interface{} |

### bulkUpdateItem 批量更新项
| 字段名 | 字段注释 | 字段类型 | 可选性 |
|--------|----------|----------|--------|
| column | 更新字段 | string | 必选 |
| records | 更新记录 | []*bulkUpdateInner | 必选 |

### bulkUpdateInner 批量更新内部项
| 字段名 | 字段注释 | 字段类型 | 可选性 |
|--------|----------|----------|--------|
| pk | 主键值 | interface{} | 必选 |
| val | 更新内容 | interface{} | 必选 |
| type | 更新方式 | int64 | 可选 |

---

## 注解说明

**通用参数说明：**
- **可选字段**: 标记为"可选"的字段在JSON中是可选的
- **内部字段**: 标记为"内部"的字段不参与JSON序列化

**查询功能：**
- **columns**: 指定返回的字段，为空时返回所有字段
- **orders**: 支持多字段排序
- **conditions**: 支持复杂的查询条件组合
- **keyword**: 全局关键词搜索
- **limit/page**: 分页查询参数

**更新功能：**
- **data**: 单条记录更新
- **bulkUpdates**: 批量更新多条记录
- **ignoreUpdateColumns**: 排除不需要更新的字段

**批量更新类型说明：**
- **type**:
  - 0: 使用占位符方式更新
  - 1: 使用原始数据方式更新

**逻辑操作符说明：**
- **logicalOperator**: 支持AND、OR等逻辑运算符
- **operator**: 支持=、!=、>、<、LIKE、IN等比较运算符

| 操作符 | 说明 | 用途 |
|--------|------|------|
| = | 等于 | 普通等于比较 |
| < | 小于 | 数值比较 |
| <= | 小于等于 | 数值比较 |
| > | 大于 | 数值比较 |
| >= | 大于等于 | 数值比较 |
| != | 不等于 | 不等于比较 |
| IN | 包含在列表中 | 多值匹配 |
| notin | 不包含在列表中 | 多值排除 |
| like | 模糊匹配 | 字符串模糊查询 |
| match | 匹配查询 | 全文搜索匹配 |
| match_phrase | 短语匹配 | 全文搜索短语匹配 |

---

## 通用请求示例

### list
```json
{
    "limit": 20,
    "page": 1,
    "orders": [{"column": "updatedAt", "value": "asc"}],
    "conditions": [
        {"column": "name", "values": ["a"]},
        {"column": "parentId", "value": 0}
    ]
}
```

### add
```json
{
    "record": {
        "parentId": 0,
        "name": "ceshi1",
        "roleIds": [1],
        "remark": "sa",
        "state": 1
    }
}
```

### update
```json
{
  "conditions": [{"column": "id", "value": 1}],
  "data": [
    {"column": "state", "value": 0},
    {"column": "name", "value": "ceshi1"}
  ]
}
```

### delete
```json
{"conditions": [{"column": "id", "value": 2}]}
```