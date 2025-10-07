package api

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/cnaize/meds/src/core/metrics"
)

func Register(r *gin.Engine) {
	// register prometheus metrics
	reg := prometheus.NewRegistry()
	metrics.Get().Register(reg)

	root := r.Group("/v1")
	root.GET("/metrics", gin.WrapH(promhttp.HandlerFor(reg, promhttp.HandlerOpts{})))

	// register api endpoints
}
