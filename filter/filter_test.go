package filter

import (
	"testing"
)

func TestFilter(t *testing.T) {
	fieldFilter := NewFieldFilter(nil)
	err := fieldFilter.Parse("a < 2 && b == 5 || c == 2")
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	ret := fieldFilter.Match(map[string]interface{}{"a": 1, "b": 5, "c": 3})
	if !ret {
		t.Fail()
	}

	ret = fieldFilter.Match(map[string]interface{}{"a": 1, "b": 6, "c": 2})
	if !ret {
		t.Fail()
	}

	ret = fieldFilter.Match(map[string]interface{}{"a": 1, "b": 6, "c": 3})
	if ret {
		t.Fail()
	}
}

func TestFilter2(t *testing.T) {
	fieldFilter := NewFieldFilter(nil)
	err := fieldFilter.Parse("name == \"aaa\" && value < 2 || c >= \"bbb\"")
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	ret := fieldFilter.Match(map[string]interface{}{"value": 1, "name": "aaa", "c": 3})
	if !ret {
		t.Fail()
	}

	ret = fieldFilter.Match(map[string]interface{}{"value": 1, "name": "bbb", "c": "ccc"})
	if !ret {
		t.Fail()
	}

	ret = fieldFilter.Match(map[string]interface{}{"value": 1, "name": 6, "c": 3})
	if ret {
		t.Fail()
	}
}

func TestFilter3(t *testing.T) {
	fieldFilter := NewFieldFilter(nil)
	err := fieldFilter.Parse("a == true && b == false")
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	ret := fieldFilter.Match(map[string]interface{}{"a": true, "b": false, "c": 3})
	if !ret {
		t.Fail()
	}

	ret = fieldFilter.Match(map[string]interface{}{"a": true, "b": true, "c": "ccc"})
	if ret {
		t.Fail()
	}

	ret = fieldFilter.Match(map[string]interface{}{"a": true, "b": 1, "c": 3})
	if ret {
		t.Fail()
	}
}

func TestFilterForArray(t *testing.T) {
	fieldFilter := NewFieldFilter(nil)
	err := fieldFilter.Parse("isarray(root[ * ].a)")
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	ret := fieldFilter.Match([]map[string]interface{}{{"a": 1, "b": 5, "c": 3}, {"a": 2, "b": 6, "c": 4}})
	if !ret {
		t.Fail()
	}
}
