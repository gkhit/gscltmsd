package service

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	mq "github.com/eclipse/paho.mqtt.golang"
	"github.com/gkhit/sm2x"
)

type (
	// Service struct
	Service struct {
		Options    *ServiceOptions
		Database   *PoolDB
		MQTTClient mq.Client
	}
)

// NewService return new service instance
func NewService(o *ServiceOptions) (s *Service) {
	return &Service{
		Options:  o,
		Database: nil,
	}
}

func (s *Service) init(ctx context.Context) (err error) {
	var (
		m *mq.ClientOptions
	)

	s.Database, err = NewDatabase(ctx, &s.Options.Database)
	if err != nil {
		return
	}

	if len(s.Options.Routes) > 0 {
		m, err = s.Options.Mqtt.NewMqttClientOptions()
		if err != nil {
			return
		}
		s.MQTTClient = mq.NewClient(m)

		if token := s.MQTTClient.Connect(); token.Wait() && token.Error() != nil {
			err = token.Error()
			return
		}

		for _, toc := range s.Options.Routes {
			if token := s.MQTTClient.Subscribe(toc.Topic, s.Options.Mqtt.Qos, s.getHandler(ctx, toc.EntryPoint)); token.Wait() && token.Error() != nil {
				err = token.Error()
				return
			}
		}
	}

	return nil
}

// Start starting service instance
func (s *Service) Start() {
	ctx := context.Background()
	ctxCancel, cancel := context.WithCancel(ctx)

	defer cancel()

	err := s.init(ctxCancel)
	if err != nil {
		log.Fatal(err)
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	s.MQTTClient.Disconnect(250)
}

func (s *Service) getHandler(ctx context.Context, entry string) mq.MessageHandler {
	var f = func(c mq.Client, m mq.Message) {
		var (
			err     error
			payload []byte
		)
		if s.Options.Database.ToXML && len(s.Options.Database.XMLRoot) > 0 {
			var src map[string]interface{}
			if err = json.Unmarshal(m.Payload(), &src); err != nil {
				log.Println(err.Error())
				return
			}
			payload, err = sm2x.Map2XML(src, "doc")
			if err != nil {
				log.Println(err.Error())
				return
			}
		} else {
			payload = m.Payload()
		}
		s.Database.CallEntryPoint(ctx, entry, m.Topic(), string(payload), time.Duration(s.Options.Database.Timeout)*time.Second)
	}
	return f
}
