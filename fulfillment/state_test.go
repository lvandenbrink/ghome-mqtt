package fulfillment

import (
	"testing"
)

func TestSetState(t *testing.T) {
	// Setup Fulfillment with one device
	f := &Fulfillment{
		devices: map[string]Device{
			"dev1": {
				State: LocalState{},
			},
		},
	}

	cases := []struct {
		name    string
		payload map[string]interface{}
		want    LocalState
	}{
		{
			name:    "lowercase on",
			payload: map[string]interface{}{"state": "on"},
			want:    LocalState{State: "ON", On: true},
		},
		{
			name:    "uppercase ON",
			payload: map[string]interface{}{"state": "ON"},
			want:    LocalState{State: "ON", On: true},
		},
		{
			name:    "OFF",
			payload: map[string]interface{}{"state": "OFF"},
			want:    LocalState{State: "OFF", On: false},
		},
		{
			name:    "missing state",
			payload: map[string]interface{}{},
			want:    LocalState{}, // Should not change
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// Reset device state before each test
			f.devices["dev1"] = Device{State: LocalState{}}
			f.setState("dev1", c.payload)
			got := f.devices["dev1"].State
			// Compare only State and On fields
			if got.State != c.want.State || got.On != c.want.On {
				t.Errorf("setState(%v) = %+v, want %+v", c.payload, got, c.want)
			}
		})
	}
}
