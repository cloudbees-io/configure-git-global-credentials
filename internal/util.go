package internal

import (
	"fmt"
	"os"
)

func Debug(msg string, args ...any) {
	if os.Getenv("RUNNER_DEBUG") == "1" {
		fmt.Println("##[debug]" + fmt.Sprintf(msg, args...))
	}
}
