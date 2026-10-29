package service

import (
	"context"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/process"

	"mixm/backend/internal/model"
)

type System struct{ *Core }

// Stats 读取宿主机 CPU、内存、磁盘和进程统计。
func (s *System) Stats(ctx context.Context, actor model.Actor) (any, error) {
	if err := s.permission(ctx, actor, "Dashboard"); err != nil {
		return nil, err
	}

	c, err := cpu.PercentWithContext(ctx, 0, false)
	if err != nil {
		return nil, err
	}

	m, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, err
	}

	d, err := disk.UsageWithContext(ctx, "/")
	if err != nil {
		return nil, err
	}

	p, err := process.PidsWithContext(ctx)
	if err != nil {
		return nil, err
	}

	n := float64(0)
	if len(c) > 0 {
		n = c[0]
	}

	return model.Record{
		"cpu": n,
		"memory": model.Record{
			"total":   m.Total,
			"used":    m.Used,
			"percent": m.UsedPercent,
		},
		"disk": model.Record{
			"total":   d.Total,
			"used":    d.Used,
			"percent": d.UsedPercent,
		},
		"processes": len(p),
		"timestamp": time.Now(),
	}, nil
}

// Load 读取宿主机近期负载。
func (s *System) Load(ctx context.Context, actor model.Actor) (any, error) {
	if err := s.permission(ctx, actor, "Dashboard"); err != nil {
		return nil, err
	}

	v, err := load.AvgWithContext(ctx)
	if err != nil {
		return nil, err
	}

	return model.Record{
		"load_1":  v.Load1,
		"load_5":  v.Load5,
		"load_15": v.Load15,
	}, nil
}
