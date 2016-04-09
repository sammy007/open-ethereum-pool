package proxy

import (
	"encoding/json"
	"log"
	"net/http"
	"sync/atomic"
)

func (s *ProxyServer) StatusIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)

	reply := make(map[string]interface{})

	var upstreams []interface{}
	current := atomic.LoadInt32(&s.upstream)

	for i, u := range s.upstreams {
		upstream := map[string]interface{}{
			"name":    u.Name,
			"sick":    u.Sick(),
			"current": current == int32(i),
		}
		upstreams = append(upstreams, upstream)
	}
	reply["upstreams"] = upstreams

	err := json.NewEncoder(w).Encode(reply)
	if err != nil {
		log.Println("Error serializing API response: ", err)
	}
}
