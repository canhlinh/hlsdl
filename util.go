package hlsdl

import (
	"encoding/json"
	"fmt"
)

func printStruct(v interface{}) {
	d, _ := json.Marshal(v)
	fmt.Println(string(d))
}
