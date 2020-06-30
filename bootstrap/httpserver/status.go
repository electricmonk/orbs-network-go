package httpserver

import (
	"encoding/json"
	"github.com/orbs-network/orbs-network-go/config"
	"github.com/orbs-network/scribe/log"
	"net/http"
	"time"
)

type Payload struct {
	Uptime int64

	BlockStorage_BlockHeight   int64
	StateStorage_BlockHeight   int64
	BlockStorage_LastCommitted int64

	Gossip_IncomingConnections int64
	Gossip_OutgoingConnections int64

	Management_LastUpdated  int64
	Management_Subscription string

	Version config.Version
}

type StatusResponse struct {
	Timestamp time.Time
	Status    string
	Error     string
	Payload   Payload
}

func (s *HttpServer) getStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status := StatusResponse{
		Timestamp: time.Now(),
		Status:    s.getStatusWarningMessage(),
		Payload: Payload{
			Uptime: s.getGaugeValueFromMetrics("Runtime.Uptime.Seconds"),

			BlockStorage_BlockHeight:   s.getGaugeValueFromMetrics("BlockStorage.BlockHeight"),
			StateStorage_BlockHeight:   s.getGaugeValueFromMetrics("StateStorage.BlockHeight"),
			BlockStorage_LastCommitted: s.getGaugeValueFromMetrics("BlockStorage.LastCommitted.TimeNano"),

			Gossip_IncomingConnections: s.getGaugeValueFromMetrics("Gossip.IncomingConnection.Active.Count"),
			Gossip_OutgoingConnections: s.getGaugeValueFromMetrics("Gossip.OutgoingConnection.Active.Count"),

			Management_LastUpdated:  s.getGaugeValueFromMetrics("Management.LastUpdateTime"),
			Management_Subscription: s.getStringValueFromMetrics("Management.Subscription.Current"),

			Version: config.GetVersion(),
		},
	}

	data, _ := json.MarshalIndent(status, "", "  ")

	_, err := w.Write(data)
	if err != nil {
		s.logger.Info("error writing index.json response", log.Error(err))

	}

}

func (s *HttpServer) getStatusWarningMessage() string {
	maxTimeSinceLastBlock := s.config.TransactionPoolTimeBetweenEmptyBlocks().Nanoseconds() * 10
	if maxTimeSinceLastBlock < 600000000 { // ten minutes
		maxTimeSinceLastBlock = 600000000
	}
	if s.getGaugeValueFromMetrics("ConsensusAlgo.LeanHelix.LastCommitted.TimeNano")+maxTimeSinceLastBlock <
		time.Now().UnixNano() {
		return "Last Successful Committed Block was too long ago"
	}

	if len(s.config.ManagementFilePath()) != 0 && s.config.ManagementPollingInterval() > 0 {
		maxIntervalSinceLastSuccessfulManagementUpdate := int64(s.config.ManagementPollingInterval().Seconds()) * 20
		if s.metricRegistry.Get("Management.Data.LastSuccessfulUpdateTime").Export().LogRow()[0].Value().(int64)+maxIntervalSinceLastSuccessfulManagementUpdate <
			time.Now().Unix() {
			return "Last Successful Management Update was too long ago"
		}
	}

	return "OK"
}

func (s *HttpServer) getGaugeValueFromMetrics(name string) (value int64) {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("could not retrieve metric", log.String("metric", name))
		}
	}()

	rows := s.metricRegistry.Get(name).Export().LogRow()
	value = rows[len(rows)-1].Int
	return value
}

func (s *HttpServer) getStringValueFromMetrics(name string) (value string) {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("could not retrieve metric", log.String("metric", name))
		}
	}()

	rows := s.metricRegistry.Get(name).Export().LogRow()
	value = rows[len(rows)-1].StringVal
	return
}
