package orm

import (
	"errors"
	"fmt"
	"strings"

	"skeyevss/core/pkg/functions"
)

type (
	ConditionOriginalItem struct {
		Query  string
		Values []interface{}
	}

	ConditionItem struct {
		Column   string        `json:"column"`
		Value    interface{}   `json:"value,optional"`
		Values   []interface{} `json:"values,optional"`
		Operator string        `json:"operator,optional"`
		UseNil   bool          `json:"-"`

		// 原始条件
		Original *ConditionOriginalItem `json:"original,optional"`

		// 与兄弟条件关系
		LogicalOperator string           `json:"logicalOperator,optional"`
		Inner           []*ConditionItem `json:"inner,optional"`

		Columns []string `json:"-"`
	}

	ConditionBuild[T Model] struct {
		conditions    []*ConditionItem
		originalModel T
		databaseType  string
	}
)

func (c *ConditionItem) operatorValidate() bool {
	if c.Operator == "" {
		c.Operator = "="
		return true
	}

	switch c.Operator {
	case "<", "<=", ">", ">=", "!=", "in", "IN", "notin", "NOTIN", "=", "like", "LIKE", "llike", "LLIKE", "jin", "JIN":
		return true
	}

	return false
}

func (c *ConditionItem) logicalOperatorValidate() bool {
	if c.LogicalOperator == "" {
		c.LogicalOperator = "AND"
		return true
	}

	for _, item := range LogicalOperators {
		if c.LogicalOperator == item {
			return true
		}
	}

	return false
}

func (c *ConditionItem) columnValidate() error {
	var columnSign = c.Column != ""
	if !columnSign && len(c.Inner) <= 0 {
		return errors.New("condition item column,inner 不能同时为空")
	}

	if columnSign {
		if c.Value == nil && len(c.Values) <= 0 {
			return fmt.Errorf("condition item [%s] value values 不能同时为空", c.Column)
		}

		if !functions.Contains(c.Column, c.Columns) {
			return fmt.Errorf("条件字段[%s]不存在", c.Column)
		}
	}

	return nil
}

func NewConditionBuild[T Model](conditions []*ConditionItem, model T, databaseType string) *ConditionBuild[T] {
	return &ConditionBuild[T]{
		conditions:    conditions,
		originalModel: model,
		databaseType:  databaseType,
	}
}

func (c *ConditionBuild[T]) Do(emptyCondition bool) (string, []interface{}, error) {
	if len(c.conditions) <= 0 {
		if emptyCondition {
			return "", nil, nil
		}

		return "", nil, errors.New("conditions is nil")
	}

	var (
		columns      = c.originalModel.Columns()
		wheres       []string
		placeholders []interface{}
	)
	for _, item := range c.conditions {
		item.Columns = columns
		if item.Original == nil {
			if err := item.columnValidate(); err != nil {
				return "", nil, err
			}
		}

		if !item.operatorValidate() {
			return "", nil, errors.New("condition item operator 值不匹配")
		}

	RETRY:
		if !item.logicalOperatorValidate() {
			return "", nil, errors.New("condition item logicalOperator 值不匹配")
		}

		if len(item.Inner) > 0 {
			whereStr, args, err := NewConditionBuild[T](item.Inner, c.originalModel, c.databaseType).Do(emptyCondition)
			if err != nil {
				return "", nil, err
			}

			whereStr = fmt.Sprintf(" %s %s", item.LogicalOperator, whereStr)
			for _, item := range LogicalOperators {
				whereStr = strings.Trim(whereStr, " "+item+" ")
			}

			wheres = append(wheres, fmt.Sprintf(" %s (%s)", item.LogicalOperator, whereStr))
			placeholders = append(placeholders, args...)
			continue
		}

		if item.Original != nil && item.Original.Query != "" {
			wheres = append(wheres, fmt.Sprintf(" %s (%s)", item.LogicalOperator, item.Original.Query))
			placeholders = append(placeholders, item.Original.Values...)
			continue
		}

		if len(item.Values) > 0 {
			if strings.ToLower(item.Operator) == "jin" {
				item.Original = NewExternalDB(c.databaseType).MakeCaseJSONContainsCondition(item.Column, item.Values)
				goto RETRY
			}

			var operator = "in"
			if strings.ToLower(item.Operator) == "notin" {
				operator = "not in"
			}

			wheres = append(
				wheres,
				fmt.Sprintf(
					" %s `%s` %s (%s)",
					item.LogicalOperator,
					item.Column,
					operator,
					strings.Trim(strings.Repeat("?,", len(item.Values)), ","),
				),
			)

			placeholders = append(placeholders, item.Values...)
			continue
		}
		wheres = append(wheres, fmt.Sprintf(" %s `%s` %s ?", item.LogicalOperator, item.Column, item.Operator))
		if item.Operator == "like" {
			placeholders = append(placeholders, fmt.Sprintf("%%%v%%", item.Value))
		} else {
			placeholders = append(placeholders, item.Value)
		}
	}

	var where string
	for _, item := range wheres {
		where += item
	}

	for _, item := range LogicalOperators {
		where = strings.Trim(where, " "+item+" ")
	}

	return where, placeholders, nil
}
