package utils

import (
	utils "GameWala-Arcade/utils"
	"crypto/tls"
	"crypto/x509"
	json "encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTService struct {
	client mqtt.Client
}

func NewMQTTService(broker, clientID, username, password, caCertPath string, tlsSkipVerify bool) (*MQTTService, error) {
	if strings.TrimSpace(broker) == "" {
		return nil, fmt.Errorf("mqtt broker URL is required")
	}

	if strings.TrimSpace(clientID) == "" {
		return nil, fmt.Errorf("mqtt client ID is required")
	}

	parsedBroker, err := url.Parse(broker)
	if err != nil {
		return nil, fmt.Errorf("invalid mqtt broker URL %q: %w", broker, err)
	}

	switch strings.ToLower(parsedBroker.Scheme) {
	case "tcp", "ssl", "tls", "ws", "wss":
	default:
		return nil, fmt.Errorf("unsupported mqtt broker scheme %q", parsedBroker.Scheme)
	}

	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID(clientID).
		SetUsername(username).
		SetPassword(password).
		SetAutoReconnect(true).
		SetConnectTimeout(10 * time.Second).
		SetKeepAlive(30 * time.Second).
		SetPingTimeout(10 * time.Second)

	if requiresTLS(parsedBroker.Scheme) {
		tlsConfig, err := newTLSConfig(caCertPath, tlsSkipVerify)
		if err != nil {
			return nil, err
		}
		opts.SetTLSConfig(tlsConfig)
	}

	opts.SetOnConnectHandler(func(client mqtt.Client) {
		utils.LogInfo("Connected to MQTT broker: %s", broker)
	})

	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		utils.LogError("MQTT connection lost: %v", err)
	})

	client := mqtt.NewClient(opts)

	token := client.Connect()
	token.Wait()

	if token.Error() != nil {
		return nil, token.Error()
	}

	return &MQTTService{client: client}, nil
}

func (m *MQTTService) Publish(topic string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		utils.LogError("could not marshal payload, error: %v", err)
		return err
	}

	token := m.client.Publish(topic, 1, false, data)
	token.Wait()

	if token.Error() != nil {
		utils.LogError("Publish error: %v", token.Error())
		return token.Error()
	} else {
		utils.LogInfo("Published to %s topic", topic)
	}
	return nil
}

func requiresTLS(scheme string) bool {
	switch strings.ToLower(scheme) {
	case "ssl", "tls", "wss":
		return true
	default:
		return false
	}
}

func newTLSConfig(caCertPath string, tlsSkipVerify bool) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: tlsSkipVerify,
		MinVersion:         tls.VersionTLS12,
	}

	if strings.TrimSpace(caCertPath) == "" {
		return tlsConfig, nil
	}

	caCertPEM, err := os.ReadFile(caCertPath)
	if err != nil {
		return nil, fmt.Errorf("read mqtt CA certificate %q: %w", caCertPath, err)
	}

	certPool, err := x509.SystemCertPool()
	if err != nil || certPool == nil {
		certPool = x509.NewCertPool()
	}

	if !certPool.AppendCertsFromPEM(caCertPEM) {
		return nil, fmt.Errorf("parse mqtt CA certificate %q", caCertPath)
	}

	tlsConfig.RootCAs = certPool
	return tlsConfig, nil
}
