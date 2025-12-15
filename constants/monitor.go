package constants

type MonitorType string

const (
	MonitorTypeCpu     MonitorType = "cpu"
	MonitorTypeDisk    MonitorType = "disk"
	MonitorTypeNetSend MonitorType = "net_send"
	MonitorTypeNetRecv MonitorType = "net_recv"
	MonitorTypeMem     MonitorType = "mem"
)

type MonitorKV struct {
	Key   int64   `json:"key"`
	Value float64 `json:"value"`
}
