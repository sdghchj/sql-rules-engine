# rules_engine
rules engine to filter and reshape json

## Example

#### Use rule engine by sql,only support 'select ... from ... where ...'
```$go
    eng := NewJsonEngine(true).
        RegisterRuleFunction("clientid", 
            func(i rule.Rule) func([]interface{}) interface{} {
                    return func([]interface{}) interface{} {
                        return i.Name()
                }
            })
    
    //sql keywords'case insensitive
    _, err := eng.ParseSql(
                `select 
                    "3" as a.a,
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
                    clientid() as c.d,
                    b.c[2] + b.c[3] as b.b 
                from "aaa/bbb"
                where a < 2 and b.c[4] = 5 and e.f = 2`)
    if err != nil {
        fmt.Println(err)
    }

    text := `{"a":1,"b":{"c":[1,2,3,4,5]},"c":"123456789","e":{"f":2}}`
    jsonText, err := eng.ConvertJson("aaa/bbb",text)
     if err != nil {
        fmt.Println(err)
    }
    fmt.Println(string(jsonText))
    /*output
    {
        "a": {
            "a": "3",
            "b": "hello",
            "c": 15,
            "d": "3456",
            "e": "2019",
            "f": {
                "a": 1,
                "b": {
                    "c": [
                        1,
                        2,
                        3,
                        4,
                        5
                    ]
                },
                "c": "123456789",
                "e": {
                    "f": 2
                }
            }
        },
        "b": {
            "b": 7,
            "c": [
                1,
                2,
                3,
                4,
                5
            ]
        },
        "c": {
            "a": 2,
            "b": 5,
            "c": 1,
            "d": [
                1,
                1,
                3
            ],
            "e": "aaa/bbb"
        }
    }*/
    
    text = `{"a":5,"b":{"c":[6,5,4,3,2,1]},"c":"987654321"}`
    jsonText, err = eng.ConvertJson("aaa/bbb",text)
     if err != nil {
        fmt.Println(err)
    }
    fmt.Println(string(jsonText)) // null
```

#### Use rule
```$go
    rule := rule.NewJsonRule(true)
    
    //query
    fieldFilter := filter.NewFieldFilter(nil)
    fieldFilter.Parse("a < 2 && b.c[4] == 5 && e.f == 2")

    rule.AddHandler(fieldFilter)

    m := mapper.NewMapper(nil)

    //select fields
    m.AddField("3", "a.a")
    m.AddField("'hello'", "a.b")
    m.AddField("Sum(b.c)", "a.c")
    m.AddField("Substr(c,2)", "a.d")
    m.AddCurrentTimestampField("a.e", nil)
    m.AddField("*", "a.f")
    m.AddField("b.c", "")
    m.AddField("b.c[2] + b.c[3]", "b.a")
    m.AddField("CurrentTimestamp()", "b.b")

    rule.AddHandler(m)
    
    text := `
            {
                "a":1,
                "b":{
                        "c":[1,2,3,4,5]
                    },
                "c":"123456789",
                "e":{
                        "f":2
                    }
            }
            `
    jsonText, _ := rule.ConvertJson(text)
    
    fmt.Println(string(jsonText))
    /*	
        //output:
        {
            "a": {
                "a": 3,
                "b": "hello",
                "c": 15,
                "d": "3456789",
                "e": 1556611910,
                "f": {
                    "a": 1,
                    "b": {
                        "c": [
                            1,
                            2,
                            3,
                            4,
                            5
                        ]
                    },
                    "c": "123456789",
                    "e": {
                        "f": 2
                    }
                }
            },
            "b": {
                "a": 7,
                "b": 1556611910,
                "c": [
                    1,
                    2,
                    3,
                    4,
                    5
                ]
            }
        }
    */
    
    tm := time.Now()
    for i := 0; i < 10000; i ++  {
        _,_ = rule.ConvertJson(text)
    }
    fmt.Println( "time elapse:",time.Now().Sub(tm).Nanoseconds(),"ns")
    /*
        //output:
        time elapse: 4265625000 ns
    */
```

## Supported golang operators
* logic : && || !
* number: + - * / %  > < >= <=  == != & | ^ >> <<
* string: + > < >= <= == !=
* array : [index]
* other : == !=

## Supported golang constant
* nil

## Supported other operators in sql
*   =    as  ==
*   and  as  &&
*   or   as  ||
*   not  as  !

## Supported sql constant
*   null as  nil

## Supported functions (case insensitive)
* sum(numberArray)
* sum(num1,num2,num3...)
* average(numberArray)
* average(num1,num2,num3...)
* len(array)
* min(numberArray)
* min(num1,num2,num3...)
* max(numberArray)
* max(num1,num2,num3...)
* array(val1,val2,val3...) 
* substr(text,pos,length) : length is optional
* string(number)
* int(stringOrFloat)
* float(stringOrInt)
* timestamp(year,month,day,hour,minute,second)
* currenttimestamp()
* year(timestamp)
* month(timestamp)
* day(timestamp)
* hour(timestamp)
* minute(timestamp)
* second(timestamp)
* regex(text,pattern)
* in(value,array) : whether val is in array
* in(value,val1,val2,val3...) : whether val is one of val1,val2,val3...
* inrange(target,min,max) : whether target is between min and max.but if min > max, whether target >= min or target < right
* abs(number)
* exp(number)
* sqrt(number)
* power(numberX,numberY)
* ceil(number)
* floor(number)
* nullif(val,target)  : return null if val==target,or val1
* ifnull(val) : return true if val is null,or false
* iif(condition,whenTrue,whenFalse) : return whenTrue when condition is true,or whenFalse

## author

email: sdghchj@qq.com
