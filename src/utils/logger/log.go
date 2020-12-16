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

func Error (msg interface{}) {
	fmt.Println("Error : ")
	fmt.Println(msg)
}