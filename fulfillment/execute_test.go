package fulfillment

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFillMessage(t *testing.T) {
	// messageHandlerMock := &MessageHandlerMock{map[string][]string{}}
	fulfillment := &Fulfillment{
		executionTemplates: map[string]string{
			"command.Test":    `{"argument":"%s"}`,
			"command.TestTwo": `{"first_argument":"%s", "second_argument":"%s"}`,
			"command.TestInt": `{"argument":"%d"}`,
			"command.TestVar": `{"argument":"%v"}`,
		},
	}

	tests := []struct {
		name           string
		device         string
		command        string
		args           []interface{}
		expectedResult string
		expectedError  error
	}{
		{
			name:           "Format a string argument",
			device:         "device",
			command:        "command.Test",
			args:           []interface{}{"this"},
			expectedResult: `{"argument":"this"}`,
			expectedError:  nil,
		},
		{
			name:           "Format two strings arguments",
			device:         "device",
			command:        "command.TestTwo",
			args:           []interface{}{"this", "that"},
			expectedResult: `{"first_argument":"this", "second_argument":"that"}`,
			expectedError:  nil,
		},
		{
			name:           "Incorrect number of arguments",
			device:         "device",
			command:        "command.Test",
			args:           []interface{}{"this", "that"},
			expectedResult: "",
			expectedError:  errors.New(`failed number of arguments 2 doesn't, match arguments in template 1 in: {"argument":"%s"}`),
		},
		{
			name:           "Format a Int argument / questionable typesafety",
			device:         "device",
			command:        "command.TestInt",
			args:           []interface{}{42},
			expectedResult: `{"argument":"42"}`,
			expectedError:  nil,
		},
		{
			name:           "Incorrect type of argument / questionable typesafety",
			device:         "device",
			command:        "command.Test",
			args:           []interface{}{42},
			expectedResult: `{"argument":"%!s(int=42)"}`,
			expectedError:  nil,
		},
		{
			name:           "Test var argument",
			device:         "device",
			command:        "command.TestVar",
			args:           []interface{}{42},
			expectedResult: `{"argument":"42"}`,
			expectedError:  nil,
		},
		{
			name:           "Unknown template error",
			device:         "no-template-device",
			command:        "trait",
			args:           []interface{}{},
			expectedResult: "",
			expectedError:  errors.New("failed to find command `trait` for device `no-template-device` in execution template"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := fulfillment.fillMessage(test.device, test.command, test.args...)

			assert.Equal(t, test.expectedResult, result)
			if test.expectedError == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, test.expectedError.Error())
			}
		})
	}
}

func TestExecute(t *testing.T) {
	messageHandlerMock := &MessageHandlerMock{map[string]string{}, map[string]error{}}
	fulfillment := &Fulfillment{
		devices: map[string]Device{
			"test-device": {
				Topic: "topic/device-id/set",
				State: LocalState{},
			},
			"other-device": {
				Topic: "topic/device1-id/set",
				State: LocalState{},
			},
		},
		handler: messageHandlerMock,
		executionTemplates: map[string]string{
			"action.devices.commands.volumeRelative": `{"volume":"%s"}`,
		},
	}

	tests := []struct {
		name                string
		requestId           string
		payload             PayloadRequest
		sendMessageResponse error
		expectedResult      ExecuteResponse
		expectedPublication bool
		expectedTopic       string
		expectedMessage     string
	}{
		// TODO test more cases
		{
			name:      "Test a request",
			requestId: "test-request",
			payload: PayloadRequest{
				AgentUserID: "",
				Commands: []CommandRequest{
					{
						Devices: []DeviceRequest{
							{
								ID: "test-device",
							},
						},
						Execution: []ExecutionRequest{
							{
								Command: "action.devices.commands.volumeRelative",
								Params: ParamsRequest{
									RelativeSteps: -1,
								},
							},
						},
					},
				},
			},
			expectedResult: ExecuteResponse{
				RequestID: "test-request",
				Payload: ExecutePayload{
					Commands: []ExecuteCommands{
						{
							Ids:    []string{"test-device"},
							Status: Success,
							States: ExecuteStates{
								Online:        true,
								CurrentVolume: 9,
							},
						},
					},
				},
			},
			expectedPublication: true,
			expectedTopic:       "topic/device-id/set",
			expectedMessage:     `{"volume":"decrease"}`,
		},
		{
			name:      "Test multiple devices request",
			requestId: "test-request",
			payload: PayloadRequest{
				AgentUserID: "",
				Commands: []CommandRequest{
					{
						Devices: []DeviceRequest{
							{
								ID: "test-device",
							},
							{
								ID: "other-device",
							},
						},
						Execution: []ExecutionRequest{
							{
								Command: "action.devices.commands.volumeRelative",
								Params: ParamsRequest{
									RelativeSteps: -1,
								},
							},
						},
					},
				},
			},
			expectedResult: ExecuteResponse{
				RequestID: "test-request",
				Payload: ExecutePayload{
					Commands: []ExecuteCommands{
						{
							Ids:    []string{"test-device"},
							Status: Success,
							States: ExecuteStates{
								Online:        true,
								CurrentVolume: 9,
							},
						},
						{
							Ids:    []string{"other-device"},
							Status: Success,
							States: ExecuteStates{
								Online:        true,
								CurrentVolume: 9,
							},
						},
					},
				},
			},
			expectedPublication: true,
			expectedTopic:       "topic/device-id/set",
			expectedMessage:     `{"volume":"decrease"}`,
		},
		{
			name:      "Test an unknown template",
			requestId: "test-request",
			payload: PayloadRequest{
				AgentUserID: "",
				Commands: []CommandRequest{
					{
						Devices: []DeviceRequest{
							{
								ID: "test-device",
							},
						},
						Execution: []ExecutionRequest{
							{
								Command: "action.devices.commands.mute",
								Params: ParamsRequest{
									Mute: true,
								},
							},
						},
					},
				},
			},
			expectedResult: ExecuteResponse{
				RequestID: "test-request",
				Payload: ExecutePayload{
					Commands: []ExecuteCommands{
						{
							Ids:       []string{"test-device"},
							Status:    Error,
							ErrorCode: "hardError",
						},
					},
				},
			},
			expectedPublication: false,
		},
		{
			name:      "Test empty request",
			requestId: "test-empty",
			payload: PayloadRequest{
				AgentUserID: "",
				Commands:    []CommandRequest{},
			},
			expectedResult: ExecuteResponse{
				RequestID: "test-empty",
				Payload: ExecutePayload{
					Commands:  []ExecuteCommands{},
					ErrorCode: "",
				},
			},
			expectedPublication: false,
		},
		{
			name:      "Device offline error",
			requestId: "offline-request",
			payload: PayloadRequest{
				AgentUserID: "",
				Commands: []CommandRequest{
					{
						Devices: []DeviceRequest{{ID: "test-device"}},
						Execution: []ExecutionRequest{{
							Command: "action.devices.commands.volumeRelative",
							Params:  ParamsRequest{RelativeSteps: 1},
						}},
					},
				},
			},
			sendMessageResponse: fmt.Errorf("device offline"),
			expectedResult: ExecuteResponse{
				RequestID: "offline-request",
				Payload: ExecutePayload{
					Commands: []ExecuteCommands{
						deviceOfflineCommand("test-device", ""),
					},
				},
			},
			expectedTopic:       "topic/device-id/set",
			expectedMessage:     `{"volume":"increase"}`,
			expectedPublication: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			messageHandlerMock.Reset()
			messageHandlerMock.sendResponses[test.expectedTopic] = test.sendMessageResponse

			result := fulfillment.execute(test.requestId, test.payload)
			assert.Equal(t, test.expectedResult, result)
			if test.expectedPublication {
				assert.Contains(t, messageHandlerMock.messages, test.expectedTopic)
				assert.Equal(t, test.expectedMessage, messageHandlerMock.messages[test.expectedTopic])
			} else {
				assert.Empty(t, messageHandlerMock.messages)
			}
		})
	}
}

func TestStateChange(t *testing.T) {
	messageHandlerMock := &MessageHandlerMock{map[string]string{}, map[string]error{}}
	fulfillment := &Fulfillment{
		devices: map[string]Device{
			"test-device": {
				Topic: "topic/device-id/set",
				State: LocalState{
					State: "this",
					On:    false,
				},
			},
		},
		handler: messageHandlerMock,
		executionTemplates: map[string]string{
			"action.devices.commands.OnOff": `{"state":"%s"}`,
		},
	}

	payload := PayloadRequest{
		Commands: []CommandRequest{
			{
				Devices: []DeviceRequest{
					{
						ID: "test-device",
					},
				},
				Execution: []ExecutionRequest{
					{
						Command: "action.devices.commands.OnOff",
						Params: ParamsRequest{
							On: true,
						},
					},
				},
			},
		},
	}

	_ = fulfillment.execute("test-request", payload)

	assert.Equal(t, true, fulfillment.devices["test-device"].State.On)
	assert.Equal(t, "this", fulfillment.devices["test-device"].State.State)
}

func TestOpenCloseTrait(t *testing.T) {
	messageHandlerMock := &MessageHandlerMock{map[string]string{}, map[string]error{}}
	fulfillment := &Fulfillment{
		devices: map[string]Device{
			"test-blinds": {
				Topic: "topic/blinds/set",
				State: LocalState{},
			},
		},
		handler: messageHandlerMock,
		executionTemplates: map[string]string{
			"action.devices.commands.OpenClose": `{"command":"%s"}`,
		},
	}

	tests := []struct {
		name                string
		requestId           string
		openPercent         int
		expectedStatus      ExecuteStatus
		expectedOpenPercent int
		expectedPublication bool
		expectedMessage     string
	}{
		{
			name:                "Open blinds with 75% (above 50%)",
			requestId:           "open-blinds",
			openPercent:         75,
			expectedStatus:      Success,
			expectedOpenPercent: 75,
			expectedPublication: true,
			expectedMessage:     `{"command":"OPEN"}`,
		},
		{
			name:                "Close blinds with 25% (below 50%)",
			requestId:           "close-blinds",
			openPercent:         25,
			expectedStatus:      Success,
			expectedOpenPercent: 25,
			expectedPublication: true,
			expectedMessage:     `{"command":"CLOSE"}`,
		},
		{
			name:                "Fully open at 100%",
			requestId:           "fully-open",
			openPercent:         100,
			expectedStatus:      Success,
			expectedOpenPercent: 100,
			expectedPublication: true,
			expectedMessage:     `{"command":"OPEN"}`,
		},
		{
			name:                "Fully closed at 0%",
			requestId:           "fully-closed",
			openPercent:         0,
			expectedStatus:      Success,
			expectedOpenPercent: 0,
			expectedPublication: true,
			expectedMessage:     `{"command":"CLOSE"}`,
		},
		{
			name:                "Boundary case at 50%",
			requestId:           "boundary-50",
			openPercent:         50,
			expectedStatus:      Success,
			expectedOpenPercent: 50,
			expectedPublication: true,
			expectedMessage:     `{"command":"OPEN"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			messageHandlerMock.Reset()

			payload := PayloadRequest{
				Commands: []CommandRequest{
					{
						Devices: []DeviceRequest{
							{
								ID: "test-blinds",
							},
						},
						Execution: []ExecutionRequest{
							{
								Command: "action.devices.commands.OpenClose",
								Params: ParamsRequest{
									OpenPercent:   test.openPercent,
									FollowUpToken: test.requestId,
								},
							},
						},
					},
				},
			}

			result := fulfillment.execute(test.requestId, payload)

			assert.Equal(t, test.requestId, result.RequestID)
			assert.Len(t, result.Payload.Commands, 1)
			command := result.Payload.Commands[0]
			assert.Equal(t, test.expectedStatus, command.Status)
			assert.Equal(t, test.expectedOpenPercent, command.OpenPercent)
			assert.Equal(t, test.requestId, command.FollowUpToken)

			if test.expectedPublication {
				assert.Contains(t, messageHandlerMock.messages, "topic/blinds/set")
				assert.Equal(t, test.expectedMessage, messageHandlerMock.messages["topic/blinds/set"])
			} else {
				assert.Empty(t, messageHandlerMock.messages)
			}
		})
	}
}

func TestTraitWithMissingTemplate(t *testing.T) {
	messageHandlerMock := &MessageHandlerMock{map[string]string{}, map[string]error{}}
	fulfillment := &Fulfillment{
		devices: map[string]Device{
			"test-blinds": {
				Topic: "topic/blinds/set",
				State: LocalState{},
			},
		},
		handler:            messageHandlerMock,
		executionTemplates: map[string]string{}, // No templates for commands under test
	}

	tests := []struct {
		name                string
		command             string
		params              ParamsRequest
		expectedFollowUpTok string
	}{
		{
			name:    "Missing template for OpenClose",
			command: "action.devices.commands.OpenClose",
			params: ParamsRequest{
				OpenPercent:   75,
				FollowUpToken: "test-token",
			},
			expectedFollowUpTok: "test-token",
		},
		{
			name:    "Missing template for OnOff",
			command: "action.devices.commands.OnOff",
			params: ParamsRequest{
				On: true,
			},
			expectedFollowUpTok: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			messageHandlerMock.Reset()

			payload := PayloadRequest{
				Commands: []CommandRequest{
					{
						Devices: []DeviceRequest{
							{
								ID: "test-blinds",
							},
						},
						Execution: []ExecutionRequest{
							{
								Command: test.command,
								Params:  test.params,
							},
						},
					},
				},
			}

			result := fulfillment.execute("test-request", payload)

			assert.Equal(t, "test-request", result.RequestID)
			assert.Len(t, result.Payload.Commands, 1)
			command := result.Payload.Commands[0]
			assert.Equal(t, ExecuteStatus(Error), command.Status)
			assert.Equal(t, "hardError", command.ErrorCode)
			assert.Equal(t, test.expectedFollowUpTok, command.FollowUpToken)
			assert.Empty(t, messageHandlerMock.messages)
		})
	}
}

type MessageHandlerMock struct {
	messages      map[string]string
	sendResponses map[string]error
}

func (m *MessageHandlerMock) Reset() {
	m.messages = map[string]string{}
	m.sendResponses = map[string]error{}
}

func (m *MessageHandlerMock) SendMessage(topic string, message string) error {
	m.messages[topic] = message
	if m.sendResponses[topic] != nil {
		return m.sendResponses[topic]
	}
	return nil
}

func (m *MessageHandlerMock) RegisterStateChangeListener(device string, topic string, callback func(string, map[string]interface{})) error {
	return nil
}
