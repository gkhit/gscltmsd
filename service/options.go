package service

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
)

type (
	// MqttAuthType type of authorizanion
	MqttAuthType int

	// MqttOptions options of mqtt server
	MqttOptions struct {
		Host                 string       `json:"host"`
		Port                 uint16       `json:"port"`
		Ssl                  bool         `json:"ssl"`
		AuthType             MqttAuthType `json:"auth_type"`
		Username             string       `json:"username"`
		Password             string       `json:"password"`
		CACert               string       `json:"ca_cert"`
		ClientCert           string       `json:"client_cert"`
		ClientKey            string       `json:"client_key"`
		Insecure             bool         `json:"insecure"`
		KeepAlive            int64        `json:"keep_alive"`
		ConnectTimeout       int64        `json:"connect_timeout"`
		MaxReconnectInterval int64        `json:"max_reconnect_interval"`
		Qos                  byte         `json:"qos"`
	}

	// DatabaseOptions options of database
	DatabaseOptions struct {
		Host     string `json:"host"`
		Port     uint16 `json:"port"`
		Database string `json:"base"`
		User     string `json:"user"`
		Password string `json:"password"`
		Timeout  int64  `json:"timeout"`
		ToXML    bool   `json:"to_xml"`
		XMLRoot  string `json:"xml_root"`
	}

	// TopicOptions topic and entry point
	TopicOptions struct {
		Topic      string `json:"topic"`
		EntryPoint string `json:"entry"`
	}

	// ServiceOptions options of service
	ServiceOptions struct {
		Mqtt     MqttOptions     `json:"mqtt"`
		Database DatabaseOptions `json:"database"`
		Routes   []TopicOptions  `json:"routes"`
	}
)

const (
	// NoneMqttAuth Без авторизации
	NoneMqttAuth MqttAuthType = iota
	// BasicMqttAuth Авторизация по логину и паролю
	BasicMqttAuth
	// CertMqttAuth Авторизация по сертификату
	CertMqttAuth
)

var (
	toStringMqttAuthType = map[MqttAuthType]string{
		NoneMqttAuth:  "none",
		BasicMqttAuth: "basic",
		CertMqttAuth:  "cert",
	}

	toIDMqttAuthType = map[string]MqttAuthType{
		"none":  NoneMqttAuth,
		"basic": BasicMqttAuth,
		"cert":  CertMqttAuth,
	}
)

func (s MqttAuthType) String() string {
	return toStringMqttAuthType[s]
}

// MarshalJSON marshals the enum as a quoted json string
func (s MqttAuthType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(toStringMqttAuthType[s])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmashals a quoted json string to the enum value
func (s *MqttAuthType) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	// Note that if the string cannot be found then it will be set to the zero value, 'Created' in this case.
	*s = toIDMqttAuthType[j]
	return nil
}

// NewOptions return new service options
func NewOptions() *ServiceOptions {
	return &ServiceOptions{
		Mqtt: MqttOptions{
			Host:                 "127.0.0.1",
			Port:                 1883,
			Ssl:                  false,
			AuthType:             NoneMqttAuth,
			Insecure:             false,
			KeepAlive:            30,
			ConnectTimeout:       30,
			MaxReconnectInterval: 60,
			Qos:                  0,
		},
		Database: DatabaseOptions{
			Host:     "127.0.0.1",
			Port:     1433,
			Database: "master",
			User:     "sa",
			Timeout:  30,
			ToXML:    false,
			XMLRoot:  "main",
		},
	}
}

// Load loading options from file
func (c *ServiceOptions) Load(path string) error {
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

	err = json.Unmarshal(byteValue, c)
	if err != nil {
		return err
	}

	return nil
}
