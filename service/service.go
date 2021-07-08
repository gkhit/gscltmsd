package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gkhit/gscltmsd/db"
	"github.com/gkhit/gscltmsd/mq"
	"github.com/gkhit/gscltmsd/sm2x"
)

type (
	// Options
	Options struct {
		Mqtt     mq.Options `json:"mqtt"`
		Database db.Options `json:"database"`
	}

	// Service
	Service struct {
		opt *Options
		db  *sql.DB
		clt mqtt.Client
		ctx context.Context
	}
)

// NewOptions
func NewOptions() *Options {
	return &Options{
		Mqtt: mq.Options{
			Host:                 "127.0.0.1",
			Port:                 1883,
			Ssl:                  false,
			AuthType:             mq.NoneAuth,
			Insecure:             false,
			KeepAlive:            30,
			ConnectTimeout:       30,
			MaxReconnectInterval: 60,
			Qos:                  0,
			Topic:                "#",
		},
		Database: db.Options{
			Host:        "127.0.0.1",
			Port:        1433,
			DBName:      "master",
			User:        "sa",
			Timeout:     30,
			ToXML:       false,
			XMLRoot:     "doc",
			XMLExtArray: false,
		},
	}
}

// Load loading options from file
func (o *Options) Load(path string) error {
	if len(strings.TrimSpace(path)) <= 0 {
		return nil
	}

	jsonFile, err := os.Open(path)
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(byteValue, o)
	if err != nil {
		return err
	}

	return nil
}

// New return new service instance
func New(o *Options) (s *Service) {
	return &Service{
		opt: o,
		db:  db.New(&o.Database),
		clt: mq.NewClient(&o.Mqtt),
	}
}

// Start starting service instance
func (s *Service) Start() {

	var cancel context.CancelFunc

	s.ctx, cancel = context.WithCancel(context.Background())

	defer cancel()

	if token := s.clt.Subscribe(s.opt.Mqtt.Topic, s.opt.Mqtt.Qos, s.getHandler()); token.Wait() && token.Error() != nil {
		log.Fatalf("[ERROR] Can't subscribe to topic \"%s\". %v\n", s.opt.Mqtt.Topic, token.Error())
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	s.clt.Disconnect(250)
}

func (s *Service) getHandler() mqtt.MessageHandler {
	var f = func(client mqtt.Client, message mqtt.Message) {
		go s.mqttHandler(client, message)
	}
	return f
}

func (s *Service) mqttHandler(c mqtt.Client, m mqtt.Message) {
	var (
		err     error
		payload []byte
		src     map[string]interface{}
	)

	if err = json.Unmarshal(m.Payload(), &src); err != nil {
		log.Printf("[ERROR] Can't converting data of topic \"%s\". %v\n", m.Topic(), err)
		return
	}

	if s.opt.Database.XMLExtArray {
		cp := sm2x.DefaultConversionParameters()
		cp.ExtendArray = true
		payload, err = sm2x.Map2XMLParameters(src, cp, s.opt.Database.XMLRoot)
	} else {
		payload, err = sm2x.Map2XML(src, s.opt.Database.XMLRoot)
	}

	ctx, cancel := context.WithTimeout(s.ctx, time.Duration(s.opt.Database.Timeout)*time.Second)
	defer cancel()
	conn, err := s.db.Conn(ctx)
	if err != nil {
		log.Printf("[ERROR] Can't connect to SQL server. %v\n", err)
		return
	}
	defer conn.Close()

	// for debug
	// log.Printf("[INFO] Topic: %s\n %s\n", m.Topic(), string(payload))

	_, err = conn.ExecContext(ctx, s.opt.Database.EntryPointFunc, m.Topic(), string(payload))
	if err != nil {
		log.Printf("[ERROR] Call SQL server entry point error. %v\n", err)
	}
}
