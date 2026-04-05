package utils

import (
	utils "GameWala-Arcade/utils"
	"crypto/tls"
	json "encoding/json"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTService struct {
	client mqtt.Client
}

func NewMQTTService(broker, clientID string, username string, password string) (*MQTTService, error) {

	tlsConfig := &tls.Config{
		InsecureSkipVerify: false, // true if using self-signed and you don't want CA check
	}

	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID(clientID).
		SetUsername(username).
		SetPassword(password).
		SetTLSConfig(tlsConfig).
		SetAutoReconnect(true)

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
