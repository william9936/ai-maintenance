package service

import (
	"context"
	"testing"
)

func TestMonitorService(t *testing.T) {
	ctx := context.Background()
	monitorService := NewMonitorService()
	t.Logf("cpu: %f", monitorService.cpu(ctx))
	t.Logf("memory: %f", monitorService.memory(ctx))
	t.Logf("disk: %f", monitorService.disk(ctx))
	sent, recv := monitorService.net(ctx)
	t.Logf("net: %f, %f", sent, recv)
}
