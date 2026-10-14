package api

import (
	"net/http"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/process"
	"mixm-edu/backend/internal/store"
)

func (s *Server) systemRoutes() {
	s.route("GET /api/system/stats", func(w http.ResponseWriter, r *http.Request) (any, error) {
		if e := s.permission(r, "Dashboard"); e != nil {
			return nil, e
		}
		c, e := cpu.PercentWithContext(r.Context(), 0, false)
		if e != nil {
			return nil, e
		}
		m, e := mem.VirtualMemoryWithContext(r.Context())
		if e != nil {
			return nil, e
		}
		d, e := disk.UsageWithContext(r.Context(), "/")
		if e != nil {
			return nil, e
		}
		p, e := process.PidsWithContext(r.Context())
		if e != nil {
			return nil, e
		}
		n := float64(0)
		if len(c) > 0 {
			n = c[0]
		}
		return store.Row{"cpu": n, "memory": store.Row{"total": m.Total, "used": m.Used, "percent": m.UsedPercent}, "disk": store.Row{"total": d.Total, "used": d.Used, "percent": d.UsedPercent}, "processes": len(p), "timestamp": time.Now()}, nil
	})
	s.route("GET /api/system/load", func(w http.ResponseWriter, r *http.Request) (any, error) {
		if e := s.permission(r, "Dashboard"); e != nil {
			return nil, e
		}
		v, e := load.AvgWithContext(r.Context())
		if e != nil {
			return nil, e
		}
		return store.Row{"load_1": v.Load1, "load_5": v.Load5, "load_15": v.Load15}, nil
	})
}
