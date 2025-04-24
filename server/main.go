package main

import (
	"log"
	"os"
)

func InitLogger() {
	logFile, _ := os.OpenFile("log/im_server.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	log.SetOutput(logFile)
}

func main() {
	InitLogger()
}
