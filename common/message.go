package common

type Message struct {
	Id       int64  //用于排重
	From, To string // 带前缀u或g
	Content  string
	Time     int64 //精确到微秒
}
