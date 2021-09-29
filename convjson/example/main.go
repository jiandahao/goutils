package main

import (
	"encoding/json"
	"fmt"

	"github.com/jiandahao/goutils/convjson"
)

func main() {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(`
	{
		"name": "tom",
		"age": 27,
		"gender": "male",
		"extras": [
			{
				"city": "beijing",
				"countryCode": "CN",
				"cotunry": "china"
			},
			[
				{
					"hobbies": ["playing game", "swiming", "watching movies"]
				}
			]
		],
		"scores": {
			"math": 98,
			"history": 100,
			"chemistry": 99
		}
	}
	`), &data); err != nil {
		fmt.Println(err)
		return
	}

	val := convjson.NewValue(data)

	name, _ := val.Get("name")
	age, _ := val.Get("age")
	city, _ := val.Get("extras[0].city")
	hobbies, _ := val.Get("extras[1][0].hobbies")

	mathScore, _ := val.Get("scores")
	fmt.Println("name", name.MustString())
	fmt.Println("age", age.MustInt())
	fmt.Println("city", city.MustString())
	fmt.Println("hobbies", hobbies.MustString())
	fmt.Println("math scores", mathScore.MustInt())

	// get with self-defined delimiter
	city, _ = val.Get("extras[0]/city", convjson.WithDelimiter("/"))
	fmt.Println("city", city.MustString())

	val.Set("scores.math", 100)
	mathScore, _ = val.Get("scores.math")
	fmt.Println("math scores", mathScore.MustInt())

}
