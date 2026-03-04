package fulfillment

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuery(t *testing.T) {
	fulfillment := &Fulfillment{
		devices: map[string]Device{
			"outlet-1": {
				Type:  "action.devices.types.OUTLET",
				State: LocalState{On: true},
			},
			"light-1": {
				Type:  "action.devices.types.LIGHT",
				State: LocalState{On: false},
			},
			"blind-1": {
				Type:  "action.devices.types.BLINDS",
				State: LocalState{On: false},
			},
		},
	}

	tests := []struct {
		name     string
		deviceID string
		expected QueryDevice
	}{
		{
			name:     "Outlet returns on/off state",
			deviceID: "outlet-1",
			expected: QueryDevice{
				Status: QueryStatusSuccess,
				Online: true,
				On:     true,
			},
		},
		{
			name:     "Light returns brightness",
			deviceID: "light-1",
			expected: QueryDevice{
				Status:     QueryStatusSuccess,
				Online:     true,
				On:         false,
				Brightness: 100,
			},
		},
		{
			name:     "Blind returns open percent",
			deviceID: "blind-1",
			expected: QueryDevice{
				Status:      QueryStatusSuccess,
				Online:      true,
				OpenPercent: 0,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payload := PayloadRequest{
				Devices: []DeviceRequest{
					{ID: test.deviceID},
				},
			}

			response := fulfillment.query("test-request", payload)

			assert.Equal(t, "test-request", response.RequestID)
			assert.Len(t, response.Payload.Devices, 1)
			assert.Equal(t, test.expected, response.Payload.Devices[test.deviceID])
		})
	}
}

func TestHandleQueryIntentWithJaguarInput(t *testing.T) {
	fulfillment := &Fulfillment{
		devices: map[string]Device{
			"jaguar": {
				Type:  "action.devices.types.BLINDS",
				State: LocalState{On: true},
			},
		},
	}

	request := FullfillementRequest{
		RequestID: "query-jaguar-request",
		Inputs: []InputRequest{
			{
				Intent: "action.devices.QUERY",
				Payload: PayloadRequest{
					AgentUserID: "",
					Devices: []DeviceRequest{
						{ID: "jaguar", CustomData: nil},
					},
				},
			},
		},
	}

	responseValue := fulfillment.handle(request, "")
	response, ok := responseValue.(QueryResponse)
	if !assert.True(t, ok, "expected QueryResponse from QUERY intent") {
		return
	}

	assert.Equal(t, "query-jaguar-request", response.RequestID)
	device, exists := response.Payload.Devices["jaguar"]
	if !assert.True(t, exists, "expected jaguar in query response") {
		return
	}

	assert.Equal(t, QueryStatusSuccess, device.Status)
	assert.True(t, device.Online)
	assert.Equal(t, 100, device.OpenPercent)
}
