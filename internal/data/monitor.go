package data

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/Madou-Shinni/gin-quickstart/constants"
	"github.com/Madou-Shinni/go-logger"
	"go.uber.org/zap"
)

type Monitor interface {
	InterOne(ctx context.Context, data map[constants.MonitorType]*constants.MonitorKV) error
	State(ctx context.Context, tp constants.MonitorType, startTime, endTime int64) ([]*constants.MonitorKV, error)
}

type defaultMonitor struct {
	Monitor
}

func NewMonitor(m Monitor) Monitor {
	return &defaultMonitor{Monitor: m}
}

type fileMonitor struct {
	rm        sync.RWMutex
	data      map[constants.MonitorType][]*constants.MonitorKV
	path      string
	maxRecord int64
	stubTime  int64
	done      chan struct{}
}

func NewFileMonitor(path string, maxRecord int64, stubTime int64) Monitor {
	fm := &fileMonitor{
		path:      path,
		maxRecord: maxRecord,
		data:      make(map[constants.MonitorType][]*constants.MonitorKV),
		done:      make(chan struct{}),
		stubTime:  stubTime,
	}

	fm.load()

	go fm.stub()

	return fm
}

func (m *fileMonitor) InterOne(ctx context.Context, data map[constants.MonitorType]*constants.MonitorKV) error {
	m.rm.Lock()
	defer m.rm.Unlock()

	for k, v := range data {
		m.data[k] = append(m.data[k], v)
		if int64(len(m.data[k])) > m.maxRecord {
			m.data[k] = m.data[k][1:]
		}
	}

	return nil
}

// State
// 获取监控数据
// 1. 根据类型和时间范围从data中获取监控数据
// 2. 基于二分查找，从data中获取 startTime 到 endTime 之间的监控数据
// 3. 如果数据查找不到则查询两者相交的部分
func (m *fileMonitor) State(ctx context.Context, tp constants.MonitorType, startTime, endTime int64) ([]*constants.MonitorKV, error) {
	m.rm.RLock()
	defer m.rm.RUnlock()

	// 获取指定类型的监控数据
	records, exists := m.data[tp]
	if !exists {
		return nil, nil // 返回空切片表示没有数据
	}

	// 检查记录是否为空
	if len(records) == 0 {
		return nil, nil
	}

	// 使用二分查找找到 startTime 的位置
	startIndex := sort.Search(len(records), func(i int) bool {
		return records[i].Key >= startTime
	})

	// 使用二分查找找到 endTime 的位置
	endIndex := sort.Search(len(records), func(i int) bool {
		return records[i].Key > endTime
	})

	// 如果没有数据在时间范围内，返回空切片
	if startIndex >= len(records) || endIndex == 0 {
		return nil, nil
	}

	// 返回时间范围内的监控数据
	return records[startIndex:endIndex], nil
}

func (m *fileMonitor) load() {
	// 检查文件路径是否存在
	if _, err := os.Stat(m.path); os.IsNotExist(err) {
		return
	}

	// 读取文件内容
	jsonData, err := os.ReadFile(m.path)
	if err != nil {
		logger.Error("读取文件失败", zap.Error(err), zap.String("path", m.path))
		return
	}

	// 反序列化JSON数据
	if err := json.Unmarshal(jsonData, &m.data); err != nil {
		logger.Error("JSON反序列化失败", zap.Error(err), zap.String("path", m.path))
		return
	}
}

// 数据存根
// 验证path是否存在，不存在则创建
// 将data以json格式写入文件里, 覆盖原有数据
func (m *fileMonitor) stub() {
	handle := func() {
		m.rm.Lock()
		defer m.rm.Unlock()
		// 检查文件路径是否存在
		dir := filepath.Dir(m.path)
		// 创建目录
		if err := os.MkdirAll(dir, 0755); err != nil {
			logger.Error("创建目录失败", zap.Error(err), zap.String("dir", dir))
			return
		}

		// 将数据序列化为JSON
		jsonData, err := json.MarshalIndent(m.data, "", "  ")
		if err != nil {
			logger.Error("JSON序列化失败", zap.Error(err))
			return
		}

		// 将JSON数据写入文件，覆盖原有内容
		if err := os.WriteFile(m.path, jsonData, 0644); err != nil {
			logger.Error("写入文件失败", zap.Error(err), zap.String("path", m.path))
			return
		}
	}

	if m.stubTime == 0 {
		return
	}

	timer := time.NewTimer(time.Second * time.Duration(m.stubTime))
	for {
		select {
		case <-timer.C:
			handle()
			timer.Reset(time.Second * time.Duration(m.stubTime))
		case <-m.done:
			handle()
			timer.Stop()
			return
		}
	}
}
