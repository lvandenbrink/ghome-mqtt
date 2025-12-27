package fulfillment

import log "log/slog"

type DisconnectResponse struct {
}

func (f *Fulfillment) disconnect(requestId string, payload PayloadRequest) DisconnectResponse {
	log.Info("handle disconnect request", "request", requestId, "payload", payload)
	return DisconnectResponse{}
}
