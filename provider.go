package proxy

import "github.com/domgolonka/proxy-scraper/models"

type Provider interface {
	List() ([]models.Proxy, error)
	Name() string
	SetProxy(models.Proxy)
}
