package websocket

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"huadai/package/validator"
	"net/http"

	"github.com/gin-generator/ginctl/package/logger"
)

type JsonHandler struct{}

type Response struct {
	RequestId string `json:"request_id"`
	Event     string `json:"event"`
	Code      uint32 `json:"code"`
	Message   string `json:"message"`
	Content   string `json:"content"`
}

type Message struct {
	RequestId string `json:"request_id" validate:"required"`
	Event     string `json:"event" validate:"required"`
	Request   string `json:"request" validate:"omitempty"`
	Data      []byte `json:"data" validate:"omitempty"`
}

func NewJsonHandler() *JsonHandler {
	return &JsonHandler{}
}

func NewResponse() *Response {
	return &Response{}
}

func (j *JsonHandler) Distribute(client *Client, message []byte) (err error) {
	var msg Message
	err = json.Unmarshal(message, &msg)
	if err != nil {
		return
	}

	err = validator.ValidateStructWithOutCtx(msg)
	if err != nil {
		return Do(client, &Response{
			RequestId: msg.RequestId,
			Event:     msg.Event,
			Code:      http.StatusBadRequest,
			Message:   err.Error(),
		})
	}

	route, err := RouteManager.GetRoute(msg.Event)
	if err != nil {
		return Do(client, &Response{
			RequestId: msg.RequestId,
			Event:     msg.Event,
			Code:      http.StatusInternalServerError,
			Message:   err.Error(),
		})
	}

	response := route.Execute(&Request{
		Client:    client,
		Send:      []byte(msg.Request),
		Data:      msg.Data,
		RequestId: msg.RequestId,
		Event:     msg.Event,
	})

	if response != nil {
		return Do(client, response)
	}
	return
}

func (m Response) Do(client *Client) (err error) {
	bytes, err := json.Marshal(m)
	if err != nil {
		logger.ErrorString("Response", "Do",
			fmt.Sprintf("%s, fd: %s", err.Error(), client.Fd))
		return
	}

	client.SendMessage(Send{
		Protocol: websocket.TextMessage,
		Message:  bytes,
	})
	return
}
