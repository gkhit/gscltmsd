package service

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	mq "github.com/eclipse/paho.mqtt.golang"
)

// NewMqttClientOptions Создает конфигурацию подключения к MQTT серверу
func (c *MqttOptions) NewMqttClientOptions() (*mq.ClientOptions, error) {
	var (
		brok      string
		err       error
		certPool  *x509.CertPool
		pemCerts  []byte
		cltCert   tls.Certificate
		tlsConfig *tls.Config = nil
	)

	opts := mq.NewClientOptions()

	brok = fmt.Sprintf("://%s:%d", c.Host, c.Port)

	if c.Ssl && len(c.CACert) > 0 {
		certPool = x509.NewCertPool()
		pemCerts, err = ioutil.ReadFile(c.CACert)

		if err != nil {
			return nil, err
		}

		if !certPool.AppendCertsFromPEM(pemCerts) {
			return nil, errors.New("Can't append CA certificate")
		}

		tlsConfig = &tls.Config{
			RootCAs: certPool,

			ClientAuth:         tls.NoClientCert,
			ClientCAs:          nil,
			InsecureSkipVerify: c.Insecure,
		}

		if c.AuthType == CertMqttAuth && len(c.ClientCert) > 0 && len(c.ClientKey) > 0 {
			// Import client certificate/key pair
			cltCert, err = tls.LoadX509KeyPair(c.ClientCert, c.ClientKey)
			if err != nil {
				return nil, err
			}
			tlsConfig.Certificates = []tls.Certificate{cltCert}
		}

		opts.AddBroker("ssl" + brok)
		opts.SetTLSConfig(tlsConfig)
	} else {
		opts.AddBroker("tcp" + brok)
	}

	if c.AuthType == BasicMqttAuth && len(c.Username) > 0 {
		opts.SetUsername(c.Username)
	}

	if c.AuthType == BasicMqttAuth && len(c.Password) > 0 {
		opts.SetPassword(c.Password)
	}

	opts.SetKeepAlive(time.Duration(c.KeepAlive) * time.Second)
	opts.SetConnectTimeout(time.Duration(c.ConnectTimeout) * time.Second)
	opts.SetMaxReconnectInterval(time.Duration(c.MaxReconnectInterval) * time.Second)
	opts.SetAutoReconnect(true)
	opts.SetCleanSession(true)

	return opts, nil
}
