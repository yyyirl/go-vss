package categories

type (
	keyType interface {
		int | string
	}

	Item[T keyType, D any] struct {
		ID       T             `json:"id"`
		Pid      T             `json:"pid"`
		Name     string        `json:"name"`
		Raw      D             `json:"raw,omitempty,optional"`
		Children []*Item[T, D] `json:"children,omitempty,optional"`
	}

	Category[T keyType, D any] struct {
		List, Trees []*Item[T, D]
	}
)

func New[T keyType, D any]() *Category[T, D] {
	return &Category[T, D]{}
}

// Conv 转换列表为分类列表
func (c *Category[T, D]) Conv(list []D, call func(D) *Item[T, D]) *Category[T, D] {
	var length = len(list)
	c.List = make([]*Item[T, D], length)
	for key, item := range list {
		var v = call(item)
		if any(v.ID) == nil {
			continue
		}

		if tmp, ok := any(v.ID).(string); ok {
			if tmp == "" {
				continue
			}
		}

		if tmp, ok := any(v.ID).(int); ok {
			if tmp == 0 {
				continue
			}
		}

		c.List[key] = v
	}

	if len(c.List) <= 0 {
		return c
	}

	// 将分类结构化
	c.Trees = c.makeTrees(T(0))
	return c
}

// SubFlatList 结构化分类
func (c *Category[T, D]) makeTrees(pid T) []*Item[T, D] {
	var children []*Item[T, D]
	for _, item := range c.List {
		var value = new(Item[T, D])
		*value = *item
		if value.Pid == pid {
			children = append(children, value)
			value.Children = c.makeTrees(value.ID)
		}
	}

	if len(children) <= 0 {
		children = []*Item[T, D]{}
	}

	return children
}

// SubFlatList 子集树状结构转平铺列表
func (c *Category[T, D]) SubFlatList(trees []*Item[T, D]) []*Item[T, D] {
	var list []*Item[T, D]
	for _, item := range trees {
		var val = new(Item[T, D])
		*val = *item
		val.Children = nil

		list = append(list, val)
		if len(item.Children) > 0 {
			list = append(list, c.SubFlatList(item.Children)...)
		}
	}

	return list
}

// FindTrees 查找指定id下所有子集包含自身
func (c *Category[T, D]) FindTrees(parentIds []T, list []D, call func(D) *Item[T, D]) []T {
	var trees = c.Conv(list, call).Trees
	if len(trees) <= 0 {
		return nil
	}

	var (
		ids     = append([]T{}, parentIds...)
		records []*Item[T, D]
	)
	// 查询
	for _, id := range parentIds {
		var item = c.FindId(id, trees)
		if item == nil {
			continue
		}

		records = append(records, item)
	}

	for _, item := range records {
		for _, item := range c.SubFlatList(item.Children) {
			ids = append(ids, item.ID)
		}
	}

	return ids
}

func (c *Category[T, D]) FindId(id T, data []*Item[T, D]) *Item[T, D] {
	for _, item := range data {
		if item.ID == id {
			return item
		}

		if len(item.Children) > 0 {
			var data = c.FindId(id, item.Children)
			if data != nil {
				return data
			}
		}
	}

	return nil
}

// Find 查找id
func (c *Category[T, D]) Find(ID T) *Item[T, D] {
	return c.find(ID, c.Trees)
}

func (c *Category[T, D]) find(ID T, subs []*Item[T, D]) *Item[T, D] {
	for _, item := range subs {
		if item.ID == ID {
			return item
		}

		if len(item.Children) > 0 {
			if val := c.find(ID, item.Children); val != nil {
				return val
			}
		}
	}

	return nil
}

// FindParents 查找祖级
func (c *Category[T, D]) FindParents(ID T) []*Item[T, D] {
	var current *Item[T, D]
	for _, item := range c.List {
		if item.ID == ID {
			current = new(Item[T, D])
			*current = *item
		}
	}

	if current == nil {
		return nil
	}

	var data = []*Item[T, D]{current}
	if pid, ok := any(current.Pid).(string); ok {
		if pid == "" {
			return data
		}
	}

	if pid, ok := any(current.Pid).(int); ok {
		if pid == 0 {
			return data
		}
	}

	var (
		list  = append([]*Item[T, D]{current}, c.findParents(current.Pid)...)
		res   = make([]*Item[T, D], len(list))
		index = 0
	)
	for i := len(list) - 1; i >= 0; i-- {
		res[index] = list[i]
		index += 1
	}

	return res
}

func (c *Category[T, D]) findParents(pid T) []*Item[T, D] {
	var parents []*Item[T, D]
	for _, item := range c.List {
		if item.ID == pid {
			var val = new(Item[T, D])
			*val = *item
			parents = append(parents, val)

			if pid, ok := any(val.Pid).(string); ok {
				if pid == "" {
					break
				}
			}

			if pid, ok := any(val.Pid).(int); ok {
				if pid == 0 {
					break
				}
			}

			parents = append(parents, c.findParents(val.Pid)...)
		}
	}

	return parents
}

// var data = categories.New[int, CItem]().Conv(
// 	[]CItem{
// 		{ID: 1, Pid: 0, Name: "数码"},
// 		{ID: 2, Pid: 0, Name: "家电"},
// 		{ID: 3, Pid: 1, Name: "数码-手机"},
// 		{ID: 4, Pid: 1, Name: "数码-电脑"},
// 		{ID: 5, Pid: 1, Name: "数码-耳机"},
// 		{ID: 6, Pid: 3, Name: "数码-苹果-手机"},
// 		{ID: 7, Pid: 3, Name: "数码-华为-手机"},
// 		{ID: 8, Pid: 3, Name: "数码-小米-手机"},
// 		{ID: 9, Pid: 6, Name: "数码-苹果-手机-iPhone16"},
// 		{ID: 10, Pid: 6, Name: "数码-苹果-手机-iPhone16 pro max"},
// 		{ID: 11, Pid: 8, Name: "数码-小米-手机-1"},
// 		{ID: 12, Pid: 8, Name: "数码-小米-手机-2"},
// 		{ID: 13, Pid: 4, Name: "数码-电脑-apple"},
// 		{ID: 14, Pid: 4, Name: "数码-电脑-华为"},
// 		{ID: 15, Pid: 4, Name: "数码-电脑-小米"},
// 		{ID: 16, Pid: 13, Name: "数码-电脑-apple-1"},
// 		{ID: 17, Pid: 13, Name: "数码-电脑-apple-2"},
// 		{ID: 18, Pid: 2, Name: "家电-1"},
// 		{ID: 19, Pid: 2, Name: "家电-2"},
// 	},
// 	func(item CItem) *categories.Item[int] {
// 		return &categories.Item[int]{
// 			ID: item.ID, Pid: item.Pid, Name: item.Name, Raw: item,
// 		}
// 	},
// )
//
// jj, _ := json.Marshal(data.FindParents(9))
// var str bytes.Buffer
// _ = json.Indent(&str, jj, "", "    ")
// fmt.Printf("\n format: %+v \n", str.String())
