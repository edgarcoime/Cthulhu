package server

import "github.com/edgarcoime/Cthulhu-filemanager/internal/service"

type RMQServerConfig struct {
}

type rmqServer struct {
	service service.Service
}

func ListenRMQ(s service.Service, cfg RMQServerConfig) {

}
