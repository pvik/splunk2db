package main

import (
	"fmt"
)

func stringValFromInterface(val interface{}) string {
	return fmt.Sprintf("%s", val)
}
