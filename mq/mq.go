package mq

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type (
	// AuthType type of authorizanion
	AuthType int

	// Options options of mqtt server
	Options struct {
		Host                 string   `json:"host"`
		Port                 uint16   `json:"port,omitempty"`
		Ssl                  bool     `json:"ssl,omitempty"`
		AuthType             AuthType `json:"auth_type,omitempty"`
		Username             string   `json:"username,omitempty"`
		Password             string   `json:"password,omitempty"`
		CACert               string   `json:"ca_cert,omitempty"`
		ClientCert           string   `json:"client_cert,omitempty"`
		ClientKey            string   `json:"client_key,omitempty"`
		Insecure             bool     `json:"insecure,omitempty"`
		KeepAlive            int64    `json:"keep_alive,omitempty"`
		ConnectTimeout       int64    `json:"connect_timeout,omitempty"`
		MaxReconnectInterval int64    `json:"max_reconnect_interval,omitempty"`
		Qos                  byte     `json:"qos,omitempty"`
		Topic                string   `json:"topic,omitempty"`
	}
)

const (
	// NoneMqttAuth Без авторизации
	NoneAuth AuthType = iota
	// BasicMqttAuth Авторизация по логину и паролю
	BasicAuth
	// CertMqttAuth Авторизация по сертификату
	CertAuth
)

var (
	toStringAuthType = map[AuthType]string{
		NoneAuth:  "none",
		BasicAuth: "basic",
		CertAuth:  "cert",
	}

	toIDAuthType = map[string]AuthType{
		"none":  NoneAuth,
		"basic": BasicAuth,
		"cert":  CertAuth,
	}
)

func (s AuthType) String() string {
	return toStringAuthType[s]
}

// MarshalJSON marshals the enum as a quoted json string
func (s AuthType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(toStringAuthType[s])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmashals a quoted json string to the enum value
func (s *AuthType) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	// Note that if the string cannot be found then it will be set to the zero value, 'Created' in this case.
	*s = toIDAuthType[j]
	return nil
}

// NewClient
func NewClient(o *Options) mqtt.Client {
	var (
		err       error
		certPool  *x509.CertPool
		pemCerts  []byte
		cltCert   tls.Certificate
		tlsConfig *tls.Config = nil
	)

	opts := mqtt.NewClientOptions()

	brok := fmt.Sprintf("://%s:%d", o.Host, o.Port)

	if o.Ssl && len(o.CACert) > 0 {
		certPool = x509.NewCertPool()
		pemCerts, err = ioutil.ReadFile(o.CACert)

		if err != nil {
			log.Fatalf("[ERROR] Can't connect to MQTT server. %v\n", err)
		}

		if !certPool.AppendCertsFromPEM(pemCerts) {
			log.Fatalf("[ERROR] Can't connect to MQTT server. %v\n", err)
		}

		tlsConfig = &tls.Config{
			RootCAs: certPool,

			ClientAuth:         tls.NoClientCert,
			ClientCAs:          nil,
			InsecureSkipVerify: o.Insecure,
		}

		if o.AuthType == CertAuth && len(o.ClientCert) > 0 && len(o.ClientKey) > 0 {
			// Import client certificate/key pair
			cltCert, err = tls.LoadX509KeyPair(o.ClientCert, o.ClientKey)
			if err != nil {
				log.Fatalf("[ERROR] Can't connect to MQTT server. %v\n", err)
			}
			tlsConfig.Certificates = []tls.Certificate{cltCert}
		}

		opts.AddBroker("ssl" + brok)
		opts.SetTLSConfig(tlsConfig)
	} else {
		opts.AddBroker("tcp" + brok)
	}

	if o.AuthType == BasicAuth && len(o.Username) > 0 {
		opts.SetUsername(o.Username)
	}

	if o.AuthType == BasicAuth && len(o.Password) > 0 {
		opts.SetPassword(o.Password)
	}

	opts.SetKeepAlive(time.Duration(o.KeepAlive) * time.Second)
	opts.SetConnectTimeout(time.Duration(o.ConnectTimeout) * time.Second)
	opts.SetMaxReconnectInterval(time.Duration(o.MaxReconnectInterval) * time.Second)
	opts.SetAutoReconnect(true)
	opts.SetCleanSession(true)

	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("[ERROR] Can't connect to MQTT server. %v\n", token.Error())
	}

	return client
}
