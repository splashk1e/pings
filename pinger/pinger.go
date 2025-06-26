package pinger

import (
	"fmt"
	"sync"

	probing "github.com/prometheus-community/pro-bing"
)

type PingManager struct {
	pingers  map[string]*probing.Pinger
	results  map[string]bool
	comments map[string]string
	mu       sync.Mutex
}

const (
	pingCount        = 5
	concurrencyLimit = 300
)

func NewPingManager(ips []string) (*PingManager, error) {
	pingers := make(map[string]*probing.Pinger)
	for _, ip := range ips {
		if ip == "" {
			continue
		}
		pinger, err := probing.NewPinger(ip)
		//pinger.SetPrivileged(true)
		//pinger.Interval = 2 * time.Second
		pingers[ip] = pinger
		if err != nil {
			return nil, err
		}
	}
	return &PingManager{
		pingers:  pingers,
		results:  make(map[string]bool),
		comments: make(map[string]string),
		mu:       sync.Mutex{},
	}, nil
}

func (m *PingManager) Start() (map[string]bool, map[string]string) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrencyLimit)
	for ip, pinger := range m.pingers {
		wg.Add(1)
		sem <- struct{}{}
		go func(ip string, pinger *probing.Pinger) {
			defer wg.Done()
			defer func() { <-sem }()
			pinger.OnSend = (func(pkt *probing.Packet) {
				//fmt.Println("ping ", ip)
				if pkt.Seq == pingCount {
					pinger.Stop()
				}
			})
			pinger.OnFinish = (func(stat *probing.Statistics) {
				m.mu.Lock()
				defer m.mu.Unlock()
				if stat.PacketsSent-stat.PacketsRecv > 2 {
					m.results[ip] = false
					m.comments[ip] = fmt.Sprintf("Пакеты потеряны, дошло %d", stat.PacketsRecv)
					return
				} else {
					m.results[ip] = true
				}
				if stat.AvgRtt.Milliseconds() > 2000 {
					m.results[ip] = false
					m.comments[ip] = fmt.Sprintf("Среднее время ответа %v мс", stat.AvgRtt)
					return
				}
				if stat.AvgRtt.Milliseconds() > 1000 {
					m.results[ip] = true
					m.comments[ip] = fmt.Sprintf("Среднее время ответа %v мс", stat.AvgRtt)
				}
			})
			err := pinger.Run()
			if err != nil {
				m.results[ip] = false
				m.comments[ip] = fmt.Sprintf("Не получилось пропинговать с ошибкой: %s", err.Error())
				return
			}
		}(ip, pinger)
	}
	wg.Wait()
	return m.results, m.comments
}
