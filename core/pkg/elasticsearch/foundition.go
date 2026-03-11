/**
 * @Author:         yi
 * @Description:    foundition
 * @Version:        1.0.0
 * @Date:           2024/12/24 16:02
 */
package elasticsearch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	elastic "github.com/olivere/elastic/v7"

	"skeyevss/core/pkg/functions"
)

const (
	SORT_ASC  = "ASC"
	SORT_DESC = "DESC"
)

const (
	LogicalOperatorOrSymbol    = "||"
	LogicalOperatorOr          = "or"
	LogicalOperatorOrUppercase = "OR"
)

var LogicalOperatorsOr = []string{
	LogicalOperatorOrSymbol,
	LogicalOperatorOr,
	LogicalOperatorOrUppercase,
}

var (
	actionInsert = "insert"
	actionUpdate = "update"
	actionDelete = "delete"
	actionUpsert = "upsert"

	ActionInsert ActionType = &actionInsert
	ActionUpdate ActionType = &actionUpdate
	ActionDelete ActionType = &actionDelete
	ActionUpsert ActionType = &actionUpsert
)

type (
	ActionType *string

	OrderItem struct {
		Column string `json:"column"`
		Value  string `json:"value"`
	}

	ConditionItem struct {
		Column   string        `json:"column"`
		Value    interface{}   `json:"value"`
		Values   []interface{} `json:"values"`
		Operator string        `json:"operator"`
		UseNil   bool          `json:"-"`

		// 原始条件
		// Original string `json:"-"`

		LogicalOperator string           `json:"logicalOperator"`
		Inner           []*ConditionItem `json:"inner"`
	}

	UpdateItem struct {
		Column string      `json:"column"`
		Value  interface{} `json:"value"`
	}

	VectorQueryItem struct {
		Column string    `json:"column"`
		Value  []float64 `json:"value"`
	}

	ReqParams struct {
		Columns    []string         `json:"columns"`
		Orders     []*OrderItem     `json:"orders"`
		Conditions []*ConditionItem `json:"conditions"`
		Limit      int              `json:"limit"`
		Page       int              `json:"page"`
		Keyword    string           `json:"keyword"`
		UniqueId   string           `json:"uniqueId"`

		Data       []*UpdateItem          `json:"data"`
		DataRecord map[string]interface{} `json:"-"` // data 转 record

		// es
		VectorQuery *VectorQueryItem `json:"vectorQuery"`
		MinScore    *float64         `json:"minScore"`
		// 忽略基础条件
		IgnoreBaseConditions bool `json:"ignoreBaseConditions"`

		Backend bool `json:"-"`
	}
)

type (
	QueryConditionsCall func(params *ReqParams) (elastic.Query, error)

	Model interface {
		// 获取主键值
		GetPKValue() string
		// 获取主键字段
		GetPKColumn() string
		// 转map
		ToMap() (map[string]interface{}, error)
		// 所有字段
		Columns() []string
		// 自定义查询条件
		QueryConditions() QueryConditionsCall
		// 设置分数
		SetScore(data interface{}, score float64)
	}

	Repository[T Model] struct {
		PrimaryKey,
		Index string

		Model   T
		Client  *elastic.Client
		Columns []string
	}

	EsScript struct {
		UniqueId string                 `json:"uniqueId"`
		Params   map[string]interface{} `json:"params"`
		Script   string                 `json:"script"`
	}
)

// -------------------- find

// 获取单条记录
func (r *Repository[T]) FindRow(ctx context.Context, fetchSourceContext []string, query elastic.Query) (*T, error) {
	res, err := r.Client.Search(r.Index).Pretty(true).Query(query).FetchSourceContext(
		elastic.NewFetchSourceContext(true).Include(fetchSourceContext...),
	).Size(1).Do(ctx)
	if err != nil {
		return nil, err
	}

	if len(res.Hits.Hits) <= 0 {
		return nil, ErrNotFound
	}

	var row T
	if err := functions.JSONUnmarshal(res.Hits.Hits[0].Source, &row); err != nil {
		return nil, err
	}

	return &row, nil
}

// 获取列表
func (r *Repository[T]) FindList(
	ctx context.Context,
	score *float64,
	limit,
	page int,
	fetchSourceContext []string,
	query elastic.Query,
	sortCall func(req *elastic.SearchService) (*elastic.SearchService, error),
) ([]*T, error) {
	if limit <= 0 {
		limit = 20
	}

	if len(fetchSourceContext) <= 0 {
		fetchSourceContext = r.Model.Columns()
	}

	var request = r.Client.Search(r.Index).Pretty(true)
	if sortCall != nil {
		var err error
		request, err = sortCall(request)
		if err != nil {
			return nil, err
		}
	}

	request = request.Query(query).FetchSourceContext(
		elastic.NewFetchSourceContext(true).Include(fetchSourceContext...),
	).From(
		functions.PageOffset(page, limit),
		// ).Size(limit).PostFilter(elastic.NewRangeQuery("_score").Gt(1.4)).Do(ctx)
	).Size(limit)
	if score != nil {
		request = request.MinScore(*score)
	}

	res, err := request.Do(ctx)
	if err != nil {
		return nil, err
	}

	if len(res.Hits.Hits) <= 0 {
		// return nil, ErrNotFound
		return nil, nil
	}

	var records []*T
	for _, item := range res.Hits.Hits {
		var row T
		if err := functions.JSONUnmarshal(item.Source, &row); err != nil {
			return nil, err
		}

		var data = &row
		if item.Score != nil {
			row.SetScore(data, *item.Score)
		}

		records = append(records, data)
	}

	return records, nil
}

// 获取总数
func (r *Repository[T]) FindTotal(ctx context.Context, score *float64, query elastic.Query) (int64, error) {
	var request = r.Client.Search(r.Index).Pretty(true).Query(query).Size(0)
	if score != nil {
		request = request.MinScore(*score)
	}

	res, err := request.Do(ctx)
	if err != nil {
		return 0, err
	}

	return res.Hits.TotalHits.Value, nil
}

// -------------------- update

// 根据文档id更新
func (r *Repository[T]) SetUpdate(ctx context.Context, data T) error {
	var uniqueId = data.GetPKValue()
	if uniqueId == "" {
		return errors.New("主键id不能为空")
	}

	if _, err := r.Client.Update().Pretty(true).Index(r.Index).Id(uniqueId).Doc(data).Do(ctx); err != nil {
		return err
	}

	return nil
}

// 脚本更新
func (r *Repository[T]) SetUpdateScriptQuery(ctx context.Context, query elastic.Query, script *elastic.Script) error {
	if _, err := r.Client.UpdateByQuery(r.Index).Query(query).Script(script).ProceedOnVersionConflict().Do(ctx); err != nil {
		return err
	}

	return nil
}

// 批量脚本更新
func (r *Repository[T]) SetBulkScriptUpdate(ctx context.Context, body []*EsScript) error {
	if len(body) <= 0 {
		return errors.New("脚本为空")
	}

	var bulkRequest = r.Client.Bulk()
	for _, value := range body {
		if value == nil {
			return errors.New("参数错误 value 不能为空")
		}

		bulkRequest = bulkRequest.Add(
			elastic.NewBulkUpdateRequest().ScriptedUpsert(true).Script(
				elastic.NewScriptInline(
					functions.TrimBlankChar(value.Script),
				).Lang(
					"painless",
				).Params(value.Params),
			).Index(r.Index).Id(value.UniqueId),
		)
	}

	res, err := bulkRequest.Do(ctx)
	if err != nil {
		return err
	}

	if res.Errors {
		errorsInfo, _ := functions.JSONMarshal(res.Items)
		return errors.New(string(errorsInfo))
	}

	return nil
}

// -------------------- delete

// 批量删除
func (r *Repository[T]) SetDelete(ctx context.Context, uniqueIds []string) error {
	_, err := r.Client.DeleteByQuery().Index(r.Index).Query(
		elastic.NewBoolQuery().Should(
			elastic.NewTermsQuery(
				r.Model.GetPKColumn(), functions.SliceToSliceAny(uniqueIds)...,
			),
		),
	).Do(ctx)

	if err != nil {
		return err
	}

	return nil
}

// -------------------- upsert

// 批量更新upsert
func (r *Repository[T]) SetUpsertBulk(ctx context.Context, records []T) error {
	bulkRequest := r.Client.Bulk()
	for _, item := range records {
		bulkRequest = bulkRequest.Add(
			elastic.NewBulkUpdateRequest().Index(r.Index).Id(item.GetPKValue()).Doc(item).DocAsUpsert(true),
		)
	}

	res, err := bulkRequest.Do(ctx)
	if err != nil {
		return err
	}

	if res.Errors {
		err, _ := functions.JSONMarshal(res.Items)
		return errors.New(string(err))
	}

	return nil
}

// -------------------- insert

func (r *Repository[T]) SetInsert(ctx context.Context, records []T, call func(item map[string]interface{}) map[string]interface{}) error {
	var bulkRequest = r.Client.Bulk()
	for _, item := range records {
		v, err := item.ToMap()
		if err != nil {
			return err
		}

		if call != nil {
			v = call(v)
		}

		bulkRequest = bulkRequest.Add(
			elastic.NewBulkCreateRequest().Index(r.Index).Id(item.GetPKValue()).Doc(v),
		)
	}

	res, err := bulkRequest.Do(ctx)
	if err != nil {
		return err
	}

	if res.Errors {
		err, _ := functions.JSONMarshal(res.Items)
		return errors.New(string(err))
	}

	return nil
}

// -------------------- sort

type MakeSortParams struct {
	Columns       []string
	Params        *ReqParams
	Req           *elastic.SearchService
	PK            string
	DefSortColumn string
}

func (r *Repository[T]) MakeSorts(params *MakeSortParams) (*elastic.SearchService, error) {
	if params.Params.UniqueId != "" {
		params.Req = params.Req.SortBy(
			elastic.NewScriptSort(
				elastic.NewScriptInline(
					"doc."+params.PK+".value == '"+params.Params.UniqueId+"' ? 1 : 0",
				),
				"number",
			).Order(false),
		)
	}

	if params.Params.Orders != nil {
		for _, item := range params.Params.Orders {
			if !functions.Contains(item.Column, params.Columns) {
				return nil, fmt.Errorf("sortBy item column is illegality, input `%s`", item.Column)
			}

			params.Req = params.Req.Sort(
				item.Column, strings.ToUpper(item.Value) == SORT_ASC,
			)
		}
		return params.Req, nil
	}

	if params.DefSortColumn == "" {
		return params.Req, nil
	}

	return params.Req.Sort(params.DefSortColumn, true), nil
}

// -------------------- conditions

func (r *Repository[T]) MakeConditions(params *ReqParams) (elastic.Query, error) {
	if queryConditionsCall := r.Model.QueryConditions(); queryConditionsCall != nil {
		condition, err := queryConditionsCall(params)
		if err != nil {
			return nil, err
		}

		conditions, err := r.makeConditions(params.Conditions)
		if err != nil {
			return nil, err
		}

		if condition != nil {
			conditions = append(conditions, condition)
		}
		return elastic.NewBoolQuery().Must(conditions...), nil
	}

	conditions, err := r.makeConditions(params.Conditions)
	if err != nil {
		return nil, err
	}

	return elastic.NewBoolQuery().Must(conditions...), nil
}

func (r *Repository[T]) MakeUpdateConditions(params *ReqParams) (elastic.Query, error) {
	conditions, err := r.makeConditions(params.Conditions)
	if err != nil {
		return nil, err
	}

	return elastic.NewBoolQuery().Must(conditions...), nil
}

func (r *Repository[T]) ConditionPreview(query elastic.Query) (string, error) {
	data, err := query.Source()
	if err != nil {
		return "", err
	}

	queryJSON, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}

	return string(queryJSON), nil
}

func (r *Repository[T]) makeConditions(conditions []*ConditionItem) ([]elastic.Query, error) {
	var conditionList []elastic.Query
	for _, item := range conditions {
		if len(item.Inner) > 0 {
			subConditions, err := r.makeConditions(item.Inner)
			if err != nil {
				return nil, err
			}

			if functions.Contains(item.LogicalOperator, LogicalOperatorsOr) {
				conditionList = append(conditionList, elastic.NewBoolQuery().Should(subConditions...))
			} else {
				conditionList = append(conditionList, elastic.NewBoolQuery().Must(subConditions...))
			}
			continue
		}

		var queryItem elastic.Query
		switch item.Operator {
		case "<":
			if item.Value == nil {
				return nil, errors.New("invalid condition value [1]")
			}

			queryItem = elastic.NewRangeQuery(item.Column).Lt(item.Value)

		case "<=":
			if item.Value == nil {
				return nil, errors.New("invalid condition value [2]")
			}

			queryItem = elastic.NewRangeQuery(item.Column).Lte(item.Value)

		case ">":
			if item.Value == nil {
				return nil, errors.New("invalid condition value [3]")
			}

			queryItem = elastic.NewRangeQuery(item.Column).Gt(item.Value)

		case ">=":
			if item.Value == nil {
				return nil, errors.New("invalid condition value [4]")
			}

			queryItem = elastic.NewRangeQuery(item.Column).Gte(item.Value)

		case "!=":
			if item.Value == nil {
				return nil, errors.New("invalid condition value [5]")
			}

			queryItem = elastic.NewBoolQuery().MustNot(elastic.NewTermQuery(item.Column, item.Value))

		case "in", "IN":
			if len(item.Values) <= 0 {
				return nil, errors.New("invalid condition values [1]")
			}

			queryItem = elastic.NewBoolQuery().Must(elastic.NewTermsQuery(item.Column, item.Values...))

		case "notin", "NOTIN":
			if len(item.Values) <= 0 {
				return nil, errors.New("invalid condition values [2]")
			}

			queryItem = elastic.NewBoolQuery().MustNot(elastic.NewTermsQuery(item.Column, item.Values...))

		case "like", "LIKE":
			v, ok := item.Value.(string)
			if !ok {
				return nil, errors.New("invalid condition value [6]")
			}

			queryItem = elastic.NewWildcardQuery(item.Column, v)

		default:
			if len(item.Values) > 0 {
				queryItem = elastic.NewTermsQuery(item.Column, item.Values...)
			} else if item.Value != nil {
				queryItem = elastic.NewTermQuery(item.Column, item.Value)
			} else {
				if !item.UseNil {
					return nil, errors.New("invalid condition value [7]")
				}
				queryItem = elastic.NewBoolQuery().MustNot(
					elastic.NewExistsQuery(item.Column),
				)
			}
		}
		conditionList = append(conditionList, queryItem)
	}

	return conditionList, nil
}
