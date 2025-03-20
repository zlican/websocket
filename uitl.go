package websocket

import (
	"encoding/json"
	"errors"
	w "github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
	"huadai/package/websocket/pb"
	"time"
)

func EventListener(interval time.Duration, eventFunc func(), stopChan <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 每当 Ticker 触发时，执行传入的函数
			eventFunc()
		case <-stopChan:
			// 收到停止信号，退出监听
			return
		}
	}
}

// Do 执行响应 TODO 这里可以使用范型
func Do(client *Client, response interface{}) (err error) {
	switch response.(type) {
	case *Response:
		resp, ok := response.(*Response)
		if !ok {
			return errors.New("文本结构体转换错误")
		}
		data, errs := json.Marshal(resp)
		if errs != nil {
			return errs
		}
		client.SendMessage(Send{
			Protocol: w.TextMessage,
			Message:  data,
		})
	case *pb.Respond:
		resp, ok := response.(*pb.Respond)
		if !ok {
			return errors.New("proto结构体转换错误")
		}
		data, errs := proto.Marshal(resp)
		if errs != nil {
			return errs
		}
		client.SendMessage(Send{
			Protocol: w.BinaryMessage,
			Message:  data,
		})
	}
	return
}
