package mqtt

import (
	"fmt"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func TestSendMessage(t *testing.T) {
	testCases := []struct {
		name       string
		client     mqtt.Client
		expectsErr bool
	}{
		{
			name:       "success",
			client:     &mqttClientMock{},
			expectsErr: false,
		},
		{
			name:       "timeout",
			client:     &mqttClientTimeoutMock{},
			expectsErr: true,
		},
		{
			name:       "publish error",
			client:     &mqttClientErrorMock{},
			expectsErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := Mqtt{client: tc.client}
			err := m.SendMessage("device/test", `{"state":"on"}`)
			if tc.expectsErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tc.expectsErr && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

// --- Timeout mock ---
type mqttClientTimeoutMock struct{}

func (m *mqttClientTimeoutMock) IsConnected() bool       { return true }
func (m *mqttClientTimeoutMock) IsConnectionOpen() bool  { return true }
func (m *mqttClientTimeoutMock) Connect() mqtt.Token     { return &mqtt.DummyToken{} }
func (m *mqttClientTimeoutMock) Disconnect(quiesce uint) {}
func (m *mqttClientTimeoutMock) Publish(topic string, qos byte, retained bool, payload interface{}) mqtt.Token {
	return &dummyTimeoutToken{}
}
func (m *mqttClientTimeoutMock) Subscribe(topic string, qos byte, callback mqtt.MessageHandler) mqtt.Token {
	return &mqtt.DummyToken{}
}
func (m *mqttClientTimeoutMock) SubscribeMultiple(filters map[string]byte, callback mqtt.MessageHandler) mqtt.Token {
	return &mqtt.DummyToken{}
}
func (m *mqttClientTimeoutMock) Unsubscribe(topics ...string) mqtt.Token             { return &mqtt.DummyToken{} }
func (m *mqttClientTimeoutMock) AddRoute(topic string, callback mqtt.MessageHandler) {}
func (m *mqttClientTimeoutMock) OptionsReader() mqtt.ClientOptionsReader {
	return mqtt.ClientOptionsReader{}
}

type dummyTimeoutToken struct{ mqtt.DummyToken }

func (t *dummyTimeoutToken) WaitTimeout(_ time.Duration) bool { return false }

// --- Error mock ---
type mqttClientErrorMock struct{}

func (m *mqttClientErrorMock) IsConnected() bool       { return true }
func (m *mqttClientErrorMock) IsConnectionOpen() bool  { return true }
func (m *mqttClientErrorMock) Connect() mqtt.Token     { return &mqtt.DummyToken{} }
func (m *mqttClientErrorMock) Disconnect(quiesce uint) {}
func (m *mqttClientErrorMock) Publish(topic string, qos byte, retained bool, payload interface{}) mqtt.Token {
	return &dummyErrorToken{}
}
func (m *mqttClientErrorMock) Subscribe(topic string, qos byte, callback mqtt.MessageHandler) mqtt.Token {
	return &mqtt.DummyToken{}
}
func (m *mqttClientErrorMock) SubscribeMultiple(filters map[string]byte, callback mqtt.MessageHandler) mqtt.Token {
	return &mqtt.DummyToken{}
}
func (m *mqttClientErrorMock) Unsubscribe(topics ...string) mqtt.Token             { return &mqtt.DummyToken{} }
func (m *mqttClientErrorMock) AddRoute(topic string, callback mqtt.MessageHandler) {}
func (m *mqttClientErrorMock) OptionsReader() mqtt.ClientOptionsReader {
	return mqtt.ClientOptionsReader{}
}

type dummyErrorToken struct{ mqtt.DummyToken }

func (t *dummyErrorToken) WaitTimeout(_ time.Duration) bool { return true }
func (t *dummyErrorToken) Error() error                     { return fmt.Errorf("publish error") }

type mqttClientMock struct {
}

func (m *mqttClientMock) IsConnected() bool       { return true }
func (m *mqttClientMock) IsConnectionOpen() bool  { return true }
func (m *mqttClientMock) Connect() mqtt.Token     { return &mqtt.DummyToken{} }
func (m *mqttClientMock) Disconnect(quiesce uint) {}
func (m *mqttClientMock) Publish(topic string, qos byte, retained bool, payload interface{}) mqtt.Token {
	return &mqtt.DummyToken{}
}
func (m *mqttClientMock) Subscribe(topic string, qos byte, callback mqtt.MessageHandler) mqtt.Token {
	return &mqtt.DummyToken{}
}
func (m *mqttClientMock) SubscribeMultiple(filters map[string]byte, callback mqtt.MessageHandler) mqtt.Token {
	return &mqtt.DummyToken{}
}
func (m *mqttClientMock) Unsubscribe(topics ...string) mqtt.Token             { return &mqtt.DummyToken{} }
func (m *mqttClientMock) AddRoute(topic string, callback mqtt.MessageHandler) {}
func (m *mqttClientMock) OptionsReader() mqtt.ClientOptionsReader             { return mqtt.ClientOptionsReader{} }
