package logger

import (
	"fmt"
)

func Info (msg interface{}) {
	fmt.Println(msg)
}

func Debug (msg interface{}) {
	fmt.Println(msg)
}

func Warn (msg interface{}) {
	fmt.Println("Warn : ")
	fmt.Println(msg)
}

func Error (msg interface{}) {
	fmt.Println("Error : ")
	fmt.Println(msg)
}