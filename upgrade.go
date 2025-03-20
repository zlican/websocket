package websocket

import (
	"fmt"
	"github.com/gin-generator/ginctl/package/get"
	"github.com/gorilla/websocket"
	"net/http"
	"sync/atomic"
)

const (
	Token           = "token"
	Source          = "Source"
	Version         = "Version"
	DeviceId        = "Device-Id"
	DownloadChannel = "Download-Channel"
)

func Upgrade(w http.ResponseWriter, req *http.Request) {

	if atomic.LoadUint64(&Manager.Total) >= Manager.Max {
		http.Error(w, "websocket service connections exceeded the upper limit", http.StatusInternalServerError)
		return
	}

	conn, err := (&websocket.Upgrader{
		ReadBufferSize:  get.Int("app.read_buffer_size", 4096),
		WriteBufferSize: get.Int("app.write_buffer_size", 4096),
		CheckOrigin: func(r *http.Request) bool {
			// 从header头里取参数，如果没有就从url参数里取
			if r.Header.Get(Token) == "" {
				r.Header.Set(Token, r.URL.Query().Get(Token))
			}
			if r.Header.Get(Source) == "" {
				r.Header.Set(Source, r.URL.Query().Get(Source))
			}
			if r.Header.Get(Version) == "" {
				r.Header.Set(Version, r.URL.Query().Get(Version))
			}
			if r.Header.Get(DeviceId) == "" {
				r.Header.Set(DeviceId, r.URL.Query().Get(DeviceId))
			}
			if r.Header.Get(DownloadChannel) == "" {
				r.Header.Set(DownloadChannel, r.URL.Query().Get(DownloadChannel))
			}
			return true
		},
	}).Upgrade(w, req, nil)

	if err != nil {
		http.NotFound(w, req)
		return
	}

	client := NewClient(conn, req)
	client.Send <- Send{
		Protocol: websocket.TextMessage,
		Message:  []byte(fmt.Sprintf("{\"code\": 200,\"message\": \"success\",\"content\": \"\"}")),
	}

	// 监听读
	go client.Read()
	// 监听写
	go client.Write()
	// 监听订阅
	go client.Receive()
	// 心跳检测
	go client.Heartbeat()

	Manager.Register <- client
}
