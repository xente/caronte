package metrics_publisher

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Init(port int) {

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(fmt.Sprint(":", port), nil)
}
