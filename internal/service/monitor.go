package service

import (
	"context"
	"time"

	"github.com/Madou-Shinni/gin-quickstart/constants"
	"github.com/Madou-Shinni/gin-quickstart/internal/conf"
	"github.com/Madou-Shinni/gin-quickstart/internal/data"
	"github.com/hibiken/asynq"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

var MonitorServiceEx = NewMonitorService()

type MonitorStateReq struct {
	MonitorType constants.MonitorType `json:"monitor_type"`
	StartTime   int64                 `json:"start_time"`
	EndTime     int64                 `json:"end_time"`
}

type MonitorStateInfoResp struct {
	Cpu     []*constants.MonitorKV `json:"cpu"`
	Disk    []*constants.MonitorKV `json:"disk"`
	NetSend []*constants.MonitorKV `json:"net_send"`
	NetRecv []*constants.MonitorKV `json:"net_recv"`
	Mem     []*constants.MonitorKV `json:"mem"`
}

type MonitorStateResp struct {
	Data []*constants.MonitorKV `json:"data"`
}

type MonitorService struct {
	Monitor data.Monitor
}

func NewMonitorService() *MonitorService {
	return &MonitorService{
		Monitor: data.NewFileMonitor(conf.Conf.MonitorConfig.File.Path, conf.Conf.MonitorConfig.MaxRecord, conf.Conf.MonitorConfig.File.StubTime),
	}
}

func (m *MonitorService) Handle(ctx context.Context, task *asynq.Task) error {
	key := time.Now().Unix()
	send, recv := m.net(ctx)
	return m.Monitor.InterOne(ctx, map[constants.MonitorType]*constants.MonitorKV{
		constants.MonitorTypeCpu: {
			Key:   key,
			Value: m.cpu(ctx),
		},
		constants.MonitorTypeMem: {
			Key:   key,
			Value: m.memory(ctx),
		},
		constants.MonitorTypeDisk: {
			Key:   key,
			Value: m.disk(ctx),
		},
		constants.MonitorTypeNetSend: {
			Key:   key,
			Value: send,
		},
		constants.MonitorTypeNetRecv: {
			Key:   key,
			Value: recv,
		},
	})
}

func (m *MonitorService) State(ctx context.Context, req MonitorStateReq) (MonitorStateResp, error) {
	state, err := m.Monitor.State(ctx, req.MonitorType, req.StartTime, req.EndTime)
	if err != nil {
		return MonitorStateResp{}, err
	}
	return MonitorStateResp{
		Data: state,
	}, nil
}

func (m *MonitorService) All(ctx context.Context, startTime, endTime int64) (MonitorStateInfoResp, error) {
	state, err := m.Monitor.State(ctx, constants.MonitorTypeCpu, startTime, endTime)
	if err != nil {
		return MonitorStateInfoResp{}, err
	}
	state, err = m.Monitor.State(ctx, constants.MonitorTypeDisk, startTime, endTime)
	if err != nil {
		return MonitorStateInfoResp{}, err
	}
	state, err = m.Monitor.State(ctx, constants.MonitorTypeNetSend, startTime, endTime)
	if err != nil {
		return MonitorStateInfoResp{}, err
	}
	state, err = m.Monitor.State(ctx, constants.MonitorTypeNetRecv, startTime, endTime)
	if err != nil {
		return MonitorStateInfoResp{}, err
	}
	state, err = m.Monitor.State(ctx, constants.MonitorTypeMem, startTime, endTime)
	if err != nil {
		return MonitorStateInfoResp{}, err
	}
	return MonitorStateInfoResp{
		Cpu:     state,
		Disk:    state,
		NetSend: state,
		NetRecv: state,
		Mem:     state,
	}, nil
}

func (m *MonitorService) DeleteTask(ctx context.Context, task *asynq.Scheduler) error {
	err := task.Unregister("")
	if err != nil {
		return err
	}
	return nil
}

func (m *MonitorService) cpu(ctx context.Context) float64 {
	percent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return 0
	}
	return percent[0]
}

func (m *MonitorService) memory(ctx context.Context) float64 {
	memory, err := mem.VirtualMemory()
	if err != nil {
		return 0
	}
	return memory.UsedPercent
}

func (m *MonitorService) disk(ctx context.Context) float64 {
	ds, err := disk.Usage("/")
	if err != nil {
		return 0
	}
	return float64(ds.Used) / float64(ds.Total)
}

func (m *MonitorService) net(ctx context.Context) (float64, float64) {
	n, err := net.IOCounters(true)
	if err != nil {
		return 0, 0
	}
	return float64(n[0].BytesSent), float64(n[0].BytesRecv)
}
