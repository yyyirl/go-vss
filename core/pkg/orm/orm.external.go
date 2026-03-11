package orm

import (
	"time"
)

var _ Model = (*externalModel)(nil)

type externalModel struct {
}

func (e externalModel) ToMap() map[string]interface{} {
	panic("don't call me")
}

func (e externalModel) Columns() []string {
	panic("don't call me")
}

func (e externalModel) UniqueKeys() []string {
	panic("don't call me")
}

func (e externalModel) PrimaryKey() string {
	panic("don't call me")
}

func (e externalModel) TableName() string {
	panic("don't call me")
}

func (e externalModel) OnConflictColumns(_ []string) []string {
	panic("don't call me")
}

func (e externalModel) QueryConditions(_ []*ConditionItem) []*ConditionItem {
	panic("don't call me")
}

func (e externalModel) SetConditions(_ []*ConditionItem) []*ConditionItem {
	panic("don't call me")
}

func (e externalModel) UseCache() *UseCacheAdvanced {
	panic("don't call me")
}

func (e externalModel) Correction(_ ActionType) interface{} {
	panic("don't call me")
}

func (e externalModel) CorrectionMap(_ map[string]interface{}) map[string]interface{} {
	panic("don't call me")
}

type ExternalDB struct {
	internalDB *Foundation[externalModel]
}

func NewExternalDB(dbType string) *ExternalDB {
	var db = NewFoundation[externalModel](nil, externalModel{}, time.Second)
	return &ExternalDB{db.withDBType(dbType)}
}

func (db *ExternalDB) MakeCaseNumberCondition(column string) *ConditionOriginalItem {
	return db.internalDB.CaseNumberCondition(column)
}

func (db *ExternalDB) MakeCaseSubstrCondition(column string, start, end int, subs []interface{}) *ConditionOriginalItem {
	return db.internalDB.CaseSubstrCondition(column, start, end, subs)
}

func (db *ExternalDB) MakeCaseJSONContainsCondition(column string, data []interface{}) *ConditionOriginalItem {
	return db.internalDB.CaseJSONContainsCondition(column, data)
}

func (db *ExternalDB) MakeCaseJSONContainsAllArrCondition(column string, arr []interface{}) *ConditionOriginalItem {
	return db.internalDB.CaseJSONContainsAllArrCondition(column, arr)
}
