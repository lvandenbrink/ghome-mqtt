package fulfillment

import (
	"fmt"
	log "log/slog"
	"maps"
	"net/http"
	"slices"
	"strings"
)

// type DeviceConfig struct {
// 	Devices []Device
// }

// type Device struct {
// 	ID         string     `json:"id"`
// 	Type       string     `json:"type"`
// 	Name       Name       `json:"name"`
// 	DeviceInfo DeviceInfo `json:"deviceInfo,omitempty"`
// 	Topic      string     `json:"topic"`
// 	Traits     []string   `json:"traits"`
// }

// type Name struct {
// 	DefaultNames []any  `json:"defaultNames,omitempty"`
// 	Name         string `json:"name,omitempty"`
// 	Nicknames    []any  `json:"nicknames,omitempty"`
// }

// type DeviceInfo struct {
// 	Manufacturer string `json:"manufacturer,omitempty"`
// 	Model        string `json:"model,omitempty"`
// 	HwVersion    string `json:"hwVersion,omitempty"`
// 	SwVersion    string `json:"swVersion,omitempty"`
// }

func (f *Fulfillment) StateHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html")

	template := "<html><body><table><tr><<th>Device</th><th>Type</th><th>Topic</th><th>State</th></tr>%s</table></body></html>"

	var output strings.Builder
	for _, deviceId := range slices.Sorted(maps.Keys(f.devices)) {
		device := f.devices[deviceId]
		output.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%v</td><td>%v</td><td>%v</td></tr>", deviceId, device.Type, device.Topic, toJson(device.State)))
	}

	_, err := fmt.Fprintf(w, template, output.String())
	if err != nil {
		log.Error("error state output", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (f *Fulfillment) setState(deviceId string, payload map[string]interface{}) {
	if _, ok := payload["state"]; !ok {
		log.Info("failed to get state for device", "device", deviceId, "payload", payload)
		return
	}
	state := strings.ToUpper(fmt.Sprintf("%v", payload["state"]))

	device := f.devices[deviceId]
	device.State = LocalState{
		State: state,
		On:    state == "ON",
	}
	log.Info("change state", "device", device, "old", f.devices[deviceId].State, "new", device.State)
	f.devices[deviceId] = device
}
