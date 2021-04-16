package lfu

import "testing"

func TestLFUCache(t *testing.T) {
	testCase := []struct {
		Op     string
		Params []string
	}{
		{
			"put",
			[]string{"1", "1"},
		},
		{
			"put",
			[]string{"2", "2"},
		},
		{
			"get",
			[]string{"1"},
		},
		{
			"put",
			[]string{"3", "3"},
		},
		{
			"get",
			[]string{"2"},
		},
		{
			"get",
			[]string{"3"},
		},
		{
			"put",
			[]string{"4", "4"},
		},
		{
			"get",
			[]string{"1"},
		},
		{
			"get",
			[]string{"3"},
		},
		{
			"get",
			[]string{"4"},
		},
	}

	// [null,null,null,1,null,-1,3,null,-1,3,4]
	expectGet := []interface{}{"1", nil, "3", nil, "3", "4"}

	var result []interface{}
	cache := NewCache(2)
	for _, test := range testCase {
		switch test.Op {
		case "put":
			cache.Put(test.Params[0], test.Params[1])
		case "get":
			result = append(result, cache.Get(test.Params[0]))
		}
	}

	for index := range result {
		if result[index] != expectGet[index] {
			t.Errorf("result mismatch, want: %v, but got: %v", expectGet, result)
		}
	}
}
