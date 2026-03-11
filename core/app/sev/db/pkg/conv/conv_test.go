package conv

import (
	"math/rand"
	"testing"
	"time"

	faker "github.com/bxcodec/faker/v3"

	"skeyevss/core/pkg/orm"
)

// go test -bench=. -benchmem
// go test -bench='Benchmark_orm2pb_test' -benchmem -benchtime=100x
// go test -bench='Benchmark_orm2pb2orm_test' -benchmem -benchtime=100x

func Benchmark_orm2pb_test(b *testing.B) {
	var r = &randReq{}
	for i := 0; i < b.N; i++ {
		// _, err := New("dev").ToPBParams(r.RandomReqParams())
		_, err := New("dev").ToPBParams(r.ReqParams1())
		if err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_orm2pb2orm_test(b *testing.B) {
	var r = &randReq{}
	for i := 0; i < b.N; i++ {
		// data, err := New("dev").ToPBParams(r.RandomReqParams())
		data, err := New("dev").ToPBParams(r.ReqParams1())
		if err != nil {
			b.Fatal(err)
		}

		_, err = New("dev").ToOrmParams(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

type randReq struct{}

func (r *randReq) ReqParams1() *orm.ReqParams {
	return &orm.ReqParams{
		Columns: []string{"aaa", "bbb"},
		Orders: []*orm.OrderItem{
			{
				Column: "ccc",
				Value:  orm.OrderDesc,
			},
			{
				Column: "ddd",
				Value:  orm.OrderAsc,
			},
		},
		Conditions: []*orm.ConditionItem{
			{
				Column: "eee",
				Value:  11.1,
				Values: []interface{}{
					-1,
					2,
					3.3,
					true,
					"stringstringstringstring",
				},
				Operator: ">=",
				UseNil:   true,
				Original: &orm.ConditionOriginalItem{
					Query: "queryqueryquery",
					Values: []interface{}{
						-1,
						2,
						3.3,
						true,
						"stringstringstringstring",
					},
				},
				LogicalOperator: "&&",
				Inner: []*orm.ConditionItem{
					{
						Column: "eee",
						Value:  11.1,
						Values: []interface{}{
							-1,
							2,
							3.3,
							true,
							"stringstringstringstring",
						},
						Operator: ">=",
						UseNil:   true,
						Original: &orm.ConditionOriginalItem{
							Query: "queryqueryquery",
							Values: []interface{}{
								-1,
								2,
								3.3,
								true,
								"stringstringstringstring",
							},
						},
						LogicalOperator: "&&",
						Inner:           []*orm.ConditionItem{},
						Columns:         []string{"aaa", "bbb"},
					},
					{
						Column: "eee",
						Value:  11.1,
						Values: []interface{}{
							-1,
							2,
							3.3,
							true,
							"stringstringstringstring",
						},
						Operator: ">=",
						UseNil:   true,
						Original: &orm.ConditionOriginalItem{
							Query: "queryqueryquery",
							Values: []interface{}{
								-1,
								2,
								3.3,
								true,
								"stringstringstringstring",
							},
						},
						LogicalOperator: "&&",
						Inner:           []*orm.ConditionItem{},
						Columns:         []string{"aaa", "bbb"},
					},
				},
				Columns: []string{"aaa", "bbb"},
			},
		},
		UniqueIds:      []string{"fff", "gggg"},
		Limit:          1,
		Page:           2,
		Keyword:        "hhh",
		UniqueId:       "jjjj",
		All:            true,
		Type:           3,
		IgnoreNotFound: true,
		Data: []*orm.UpdateItem{
			{
				Column: "kkkkk",
				Value:  1,
			},
			{
				Column: "lll",
				Value:  false,
			},
		},
		DataRecord: map[string]interface{}{
			"llll": 1,
			"mmmm": 1.1,
			"nnn":  false,
			"ooo":  "`````falseasdasd``*&^$%```",
		},
		Backend: true,
	}
}

func (r *randReq) RandomReqParams() *orm.ReqParams {
	rand.Seed(time.Now().UnixNano())
	req := &orm.ReqParams{
		UniqueIds:      r.randomStringSlice(2, 5),
		Limit:          rand.Intn(100) + 1,
		Page:           rand.Intn(10) + 1,
		Keyword:        faker.Word(),
		UniqueId:       faker.UUIDDigit(),
		All:            rand.Float32() > 0.5,
		Type:           int64(rand.Intn(5)),
		IgnoreNotFound: rand.Float32() > 0.5,
		Backend:        rand.Float32() > 0.5,
		Columns:        r.randomStringSlice(2, 10),
		Orders:         r.randomOrderItems(1, 3),
		Conditions:     r.randomConditionItems(1, 3),
		Data:           r.randomUpdateItems(1, 5),
		DataRecord:     r.randomDataRecord(3, 5),
	}

	return req
}

func (r *randReq) randomStringSlice(minLen, maxLen int) []string {
	length := rand.Intn(maxLen-minLen+1) + minLen
	result := make([]string, length)
	for i := range result {
		result[i] = faker.Word()
	}
	return result
}

func (r *randReq) randomOrderItems(minLen, maxLen int) []*orm.OrderItem {
	length := rand.Intn(maxLen-minLen+1) + minLen
	items := make([]*orm.OrderItem, length)

	for i := range items {
		items[i] = &orm.OrderItem{
			Column: faker.Word(),
			Value:  r.randomOrderValue(),
		}
	}
	return items
}

func (r *randReq) randomOrderValue() orm.OrderType {
	if rand.Float32() > 0.5 {
		return orm.OrderAsc
	}
	return orm.OrderDesc
}

func (r *randReq) randomConditionItems(minLen, maxLen int) []*orm.ConditionItem {
	length := rand.Intn(maxLen-minLen+1) + minLen
	items := make([]*orm.ConditionItem, length)

	for i := range items {
		items[i] = &orm.ConditionItem{
			Column:          faker.Word(),
			Value:           r.randomValue(),
			Values:          r.randomValueSlice(1, 5),
			Operator:        r.randomOperator(),
			UseNil:          rand.Float32() > 0.5,
			LogicalOperator: r.randomLogicalOperator(),
			Columns:         r.randomStringSlice(1, 3),
			Inner:           r.randomConditionItems(0, 2), // 递归生成嵌套条件
			Original: &orm.ConditionOriginalItem{
				Query:  faker.Sentence(),
				Values: r.randomValueSlice(1, 3),
			},
		}
	}
	return items
}

func (r *randReq) randomUpdateItems(minLen, maxLen int) []*orm.UpdateItem {
	length := rand.Intn(maxLen-minLen+1) + minLen
	items := make([]*orm.UpdateItem, length)

	for i := range items {
		items[i] = &orm.UpdateItem{
			Column: faker.Word(),
			Value:  r.randomValue(),
		}
	}
	return items
}

func (r *randReq) randomDataRecord(minLen, maxLen int) map[string]interface{} {
	length := rand.Intn(maxLen-minLen+1) + minLen
	result := make(map[string]interface{}, length)

	for i := 0; i < length; i++ {
		key := faker.Word()
		result[key] = r.randomValue()
	}
	return result
}

func (r *randReq) randomValue() interface{} {
	switch rand.Intn(5) {
	case 0:
		return rand.Intn(1000)
	case 1:
		return rand.Float64() * 100
	case 2:
		return rand.Float32() > 0.5
	case 3:
		return faker.Sentence()
	default:
		return faker.Paragraph()
	}
}

func (r *randReq) randomValueSlice(minLen, maxLen int) []interface{} {
	length := rand.Intn(maxLen-minLen+1) + minLen
	result := make([]interface{}, length)
	for i := range result {
		result[i] = r.randomValue()
	}
	return result
}

func (r *randReq) randomOperator() string {
	ops := []string{"=", "!=", ">", "<", ">=", "<=", "LIKE", "IN"}
	return ops[rand.Intn(len(ops))]
}

func (r *randReq) randomLogicalOperator() string {
	if rand.Float32() > 0.5 {
		return "&&"
	}
	return "||"
}
