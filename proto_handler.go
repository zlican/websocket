package websocket

import (
	"google.golang.org/protobuf/proto"
	"huadai/package/websocket/pb"
	"net/http"
)

type ProtoHandler struct{}

func NewProtoHandler() *ProtoHandler {
	return &ProtoHandler{}
}

func NewPbRespond() *pb.Respond {
	return &pb.Respond{}
}

func (p *ProtoHandler) Distribute(client *Client, message []byte) (err error) {

	var msg pb.Request
	err = proto.Unmarshal(message, &msg)
	if err != nil {
		return
	}

	if msg.Event == "" {
		return Do(client, &pb.Respond{
			RequestId: msg.RequestId,
			Event:     msg.Event,
			Code:      http.StatusBadRequest,
			Message:   "event is required",
		})
	}

	route, err := RouteManager.GetRoute(msg.Event)
	if err != nil {
		return Do(client, &pb.Respond{
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
		return Do(client, &pb.Respond{
			RequestId: response.RequestId,
			Event:     response.Event,
			Code:      response.Code,
			Message:   response.Message,
			Content:   response.Content,
		})
	}
	return
}
