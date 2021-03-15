package egorm

import (
	"net/http"
	"time"

	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/core/emetric"
	"github.com/gotomicro/ego/server/egovernor"
	jsoniter "github.com/json-iterator/go"
)

func init() {
	type gormStatus struct {
		Gorms map[string]interface{} `json:"gorms"`
	}
	var rets = gormStatus{
		Gorms: make(map[string]interface{}, 0),
	}
	egovernor.HandleFunc("/debug/gorm/stats", func(w http.ResponseWriter, r *http.Request) {
		rets.Gorms = stats()
		_ = jsoniter.NewEncoder(w).Encode(rets)
	})
	go monitor()
}

func monitor() {
	for {
		time.Sleep(time.Second * 10)
		iterate(func(name string, db *Component) bool {
			sqlDB, err := db.DB()
			if err != nil {
				elog.EgoLogger.With(elog.FieldComponent(PackageName)).Panic("monitor db error", elog.FieldErr(err))
				return false
			}

			stats := sqlDB.Stats()
			emetric.LibHandleSummary.Observe(float64(stats.Idle), name, "idle")
			emetric.LibHandleSummary.Observe(float64(stats.InUse), name, "inuse")
			emetric.LibHandleSummary.Observe(float64(stats.WaitCount), name, "wait")
			emetric.LibHandleSummary.Observe(float64(stats.OpenConnections), name, "conns")
			emetric.LibHandleSummary.Observe(float64(stats.MaxOpenConnections), name, "max_open_conns")
			emetric.LibHandleSummary.Observe(float64(stats.MaxIdleClosed), name, "max_idle_closed")
			emetric.LibHandleSummary.Observe(float64(stats.MaxLifetimeClosed), name, "max_lifetime_closed")
			return true
		})
	}
}
