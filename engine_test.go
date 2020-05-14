package engine

import (
	"fmt"
	"github.com/sdghchj/sql-rules-engine/filter"
	"github.com/sdghchj/sql-rules-engine/mapper"
	"github.com/sdghchj/sql-rules-engine/rule"
	"testing"
	"time"
)

func TestJsonEngineSql(t *testing.T) {
	eng := NewJsonEngine(true).RegisterRuleFunction("clientid", func(i rule.Rule) func([]interface{}) interface{} {
		return func([]interface{}) interface{} {
			return i.Name()
		}
	})

	_, err := eng.ParseSql(`select "3" as a.a,
								'hello' as a.b,
								Sum(b.c) as a.c,
								Substr(c,2,4) as a.d,
								string(year(currenttimestamp())) as a.e,
								* as a.f,
								ceil(1.5) as c.a,
								max(b.c) as c.b,
								min(b.c) as c.c,
								array(1,a,3) as c.d,
								b.f,
								b.c,
								clientid() as c.e,
								b.c[2] + b.c[3] as b.b 
							from "aaa/bbb"
							where a < 2 and (b.c[4] = 5 or e.f = 2)`)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
}

func TestJsonFilter(t *testing.T) {
	flt := filter.NewFieldFilter(nil)

	err := flt.Parse(`a < 2 && (b.c[4] == 5 || e.f == 2)`)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}

	text := `{"a":1,"b":{"c":[1,2,3,4,5]},"c":"123456789","e":{"f":2}}`

	ret := flt.MatchJson(text)
	fmt.Println(ret)

	text = `{"a":5,"b":{"c":[6,5,4,3,2,1]},"c":"987654321"}`

	ret = flt.MatchJson(text)
	fmt.Println(ret)
}

func TestJsonSqlFilter(t *testing.T) {
	flt := filter.NewFieldFilter(nil)

	err := flt.Parse(`a < 2 and (b.c[4] == 5 or e.f == 2)`)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}

	text := `{"a":1,"b":{"c":[1,2,3,4,5]},"c":"123456789","e":{"f":2}}`

	ret := flt.MatchJson(text)
	fmt.Println(ret)

	text = `{"a":5,"b":{"c":[6,5,4,3,2,1]},"c":"987654321"}`

	ret = flt.MatchJson(text)
	fmt.Println(ret)
}

func TestJsonEngineConvert(t *testing.T) {
	eng := NewJsonEngine(true).RegisterRuleFunction("clientid", func(i rule.Rule) func([]interface{}) interface{} {
		return func([]interface{}) interface{} {
			return i.Name()
		}
	})

	_, err := eng.ParseSql(`select "3" as a.a,
								'hello' as a.b,
								Sum(b.c) as a.c,
								Substr(c,2,4) as a.d,
								string(year(currenttimestamp())) as a.e,
								* as a.f,
								ceil(1.5) as c.a,
								max(b.c) as c.b,
								min(b.c) as c.c,
								array(1,a,3) as c.d,
								b.f,
								b.c,
								clientid() as c.e,
								b.c[2] + b.c[3] as b.b 
							from "aaa/bbb"
							where a < 2 and b.c[4] = 5 and e.f = 2`)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}

	text := `{"a":1,"b":{"c":[1,2,3,4,5]},"c":"123456789","e":{"f":2}}`

	jsonText, err := eng.ConvertJson("aaa/bbb", text)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	fmt.Println(string(jsonText))

	text = `{"a":5,"b":{"c":[6,5,4,3,2,1]},"c":"987654321"}`

	jsonText, err = eng.ConvertJson("aaa/bbb", text)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	fmt.Println(string(jsonText))
}

func TestJsonEngineConvertFromArray(t *testing.T) {
	eng := NewJsonEngine(true).RegisterRuleFunction("clientid", func(i rule.Rule) func([]interface{}) interface{} {
		return func([]interface{}) interface{} {
			return i.Name()
		}
	})

	_, err := eng.ParseSql(`select "3" as a.a,
								'hello' as a.b,
								Sum(root[0].b.c) as a.c,
								Substr(root[0].c,2,4) as a.d,
								string(year(currenttimestamp())) as a.e,
								* as a.f,
								ceil(1.5) as c.a,
								max(root[0].b.c) as c.b,
								min(root[0].b.c) as c.c,
								array(1,root[0].a,3) as c.d,
								root[0].b.f,
								root[0].b.c as b.c,
								clientid() as c.e,
								root[0].b.c[2] + root[0].b.c[3] as b.b 
							from "aaa/bbb"
							where root[0].a < 2 and root[0].b.c[4] = 5 and root[0].e.f = 2`)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}

	text := `[{"a":1,"b":{"c":[1,2,3,4,5]},"c":"123456789","e":{"f":2}}]`

	jsonText, err := eng.ConvertJson("aaa/bbb", text)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	fmt.Println(string(jsonText))

	text = `[{"a":5,"b":{"c":[6,5,4,3,2,1]},"c":"987654321"}]`

	jsonText, err = eng.ConvertJson("aaa/bbb", text)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	fmt.Println(string(jsonText))
}

func TestJsonEngineEvent(t *testing.T) {
	eng := NewJsonEngine(true)

	_, err := eng.ParseRuleAsyncEvent("aaa", `a < 2 && b.c[4] == 5 && e.f == 2`, func(src interface{}) {

		fmt.Println("do some work")
	})
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}

	text := `{"a":1,"b":{"c":[1,2,3,4,5]},"c":"123456789","e":{"f":2}}`
	err = eng.HandleJsonAsync(text)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	text = `{"a":5,"b":{"c":[6,5,4,3,2,1]},"c":"987654321"}`
	err = eng.HandleJsonAsync(text)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
}

func TestJsonRule(t *testing.T) {
	rule := rule.NewJsonRule(true)

	fieldFilter := filter.NewFieldFilter(nil)
	err := fieldFilter.Parse("a < 2 && b.c[4] == 5 && e.f == 2")
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	rule.AddHandler(fieldFilter)

	m := mapper.NewMapper(nil)

	err = m.AddField("3", "a.a")
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	err = m.AddField("'hello'", "a.b")
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	err = m.AddField("Sum(b.c)", "a.c")
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	err = m.AddField("Substr(c,2)", "a.d")
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	err = m.AddCurrentTimestampField("a.e", nil)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	err = m.AddField("*", "a.f")
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	err = m.AddField("b.c", "")
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	err = m.AddField("b.c[2] + b.c[3]", "b.a")
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	err = m.AddField("currenttimestamp()", "b.b")
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	//fail
	err = m.AddField("1a.2d", "")
	if err != nil {
		t.Log(err)
	} else {
		t.Errorf("want no nil")
		t.Fail()
	}

	//fail
	err = m.AddCurrentTimestampField("1a.#d", nil)
	if err != nil {
		t.Log(err)
	} else {
		t.Errorf("want no nil")
		t.Fail()
	}

	rule.AddHandler(m)

	text := `{"a":1,"b":{"c":[1,2,3,4,5]},"c":"123456789","e":{"f":2}}`
	jsonText, err := rule.ConvertJson(text)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	fmt.Println(string(jsonText))

	tm := time.Now()
	for i := 0; i < 10000; i++ {
		_, _ = rule.ConvertJson(text)
	}
	t.Log("time elapse:", time.Now().Sub(tm).Nanoseconds(), "ns")

	text = `{"a":5,"b":{"c":[6,5,4,3,2,1]},"c":"987654321"}`
	jsonText, err = rule.ConvertJson(text)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	fmt.Println(string(jsonText))
}
