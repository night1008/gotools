package recover

import (
	"fmt"
	"log"
	"runtime/debug"
)

func Recover() {
	if err := recover(); err != nil {
		debugStack := string(debug.Stack())
		log.Println(fmt.Errorf("err=%v, stack=%s \nrecover from panic", err, debugStack))
	}
}
