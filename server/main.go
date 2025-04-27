package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

func InitLogger() {
	logFile, _ := os.OpenFile("log/im_server.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	log.SetOutput(logFile)
}

func main() {
	InitLogger()
}

var upgrader = websocket.Upgrader{
	HandshakeTimeout: 1 * time.Second,
	ReadBufferSize:   100,
	WriteBufferSize:  100,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 用户注册，不需要传参数，返回用户id
func RegistUser(w http.ResponseWriter, r *http.Request) {
	uid := time.Now().UnixMicro() //生成全局唯一的群ID（正规来讲，应该借助于mysql的自增id）
	err := GetRabbitMQ().RegisterUser(uid, "u")
	if err != nil {
		log.Printf("创建用户%d失败:%s", uid, err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("创建用户失败"))
		return
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(strconv.FormatInt(uid, 10)))
		return
	}
}
