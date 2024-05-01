package common

import (
	"fmt"
)

func FindValueFromMap(respData map[string]any, res string, key ...string) {
	for _, data := range respData {
		fmt.Printf("%T,%p\n", data, res)
		fmt.Printf("%T\n", respData)
		fmt.Println(respData)

		if len(key) == 1 {
			res = data.(string)
		}

		respData = data.(map[string]any)
	}

}
