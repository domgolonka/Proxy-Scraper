package models

import "time"

type Proxy struct {
	IP        string    `json:"ip"`
	Port      string    `json:"port"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

func (p *Proxy) ToString() string {
	if p.Port == "" {
		return p.Type + "://" + p.IP
	}
	return p.Type + "://" + p.IP + ":" + p.Port
}

func NewProxy(ip, port, typeOf string) *Proxy {
	return &Proxy{
		IP:        ip,
		Port:      port,
		Type:      typeOf,
		CreatedAt: time.Now().UTC(),
	}
}
