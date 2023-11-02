package proxy

import (
	"crypto/rand"
	"github.com/domgolonka/proxy-scraper/models"
	"github.com/domgolonka/proxy-scraper/providers"
	aq "github.com/emirpasic/gods/queues/arrayqueue"
	"github.com/patrickmn/go-cache"
	"log/slog"
	"math/big"
	"os"
	"reflect"
	"sync"
	"time"
)

var (
	instance  *ProxyGenerator
	usedProxy sync.Map
	once      sync.Once
)

type Verify func(log *slog.Logger, proxy models.Proxy) bool

type ProxyGenerator struct { //nolint
	lastValidProxy models.Proxy
	cache          *cache.Cache
	logger         *slog.Logger
	queue          *aq.Queue
	VerifyFn       Verify
	providers      []Provider
	proxy          chan models.Proxy
	job            chan models.Proxy
}

func (p *ProxyGenerator) isProvider(provider Provider) bool {
	for _, pr := range p.providers {
		if reflect.TypeOf(pr) == reflect.TypeOf(provider) {
			return true
		}
	}
	return false
}

func (p *ProxyGenerator) AddProvider(provider Provider) {
	if !p.isProvider(provider) {
		p.providers = append(p.providers, provider)
	}
}

func shuffle(vals []models.Proxy) {
	for len(vals) > 0 {
		n := len(vals)
		nBig, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
		if err != nil {
			panic(err)
		}
		randIndex := nBig.Int64()
		vals[n-1], vals[randIndex] = vals[randIndex], vals[n-1]
		vals = vals[:n-1]
	}
}

func (p *ProxyGenerator) createOrIgnore(ip, port, ptype string) {
	p.queue.Enqueue(models.NewProxy(ip, port, ptype))
	// todo: Unique
}
func (p *ProxyGenerator) GetProxies() []models.Proxy {
	val := p.queue.Values()
	proxies := make([]models.Proxy, 0, len(val))
	for _, s := range val {
		proxies = append(proxies, s.(models.Proxy))
	}
	return proxies
}

func (p *ProxyGenerator) load() {
	for _, provider := range p.providers {
		usedProxy.Store(p.lastValidProxy, time.Now().Hour())
		provider.SetProxy(p.lastValidProxy)

		ips, err := provider.List()
		if err != nil {
			p.lastValidProxy = models.Proxy{}
			p.logger.Error("cannot load list of proxy %s err:%s", provider.Name(), err)
			continue
		}

		usedProxy.Range(func(key, value interface{}) bool {
			if value.(int) != time.Now().Hour() {
				usedProxy.Delete(key)
			}
			return true
		})

		shuffle(ips)
		for _, proxy := range ips {
			p.createOrIgnore(proxy.IP, proxy.Port, proxy.Type)
		}
	}
}

func (p *ProxyGenerator) GetLast() models.Proxy {
	proxy := <-p.proxy
	_, ok := usedProxy.Load(proxy)
	if !ok {
		p.lastValidProxy = proxy
	}
	return proxy
}

func (p *ProxyGenerator) Count() int {

	return p.cache.ItemCount()
}
func (p *ProxyGenerator) Get() []string {
	items := make([]string, 0, len(p.cache.Items()))
	for k := range p.cache.Items() {
		items = append(items, k)
	}
	return items
}

func (p *ProxyGenerator) verifyWithCache(proxy models.Proxy) bool {
	val, found := p.cache.Get(proxy.ToString())
	if found {
		return val.(bool)
	}
	res := p.VerifyFn(p.logger, proxy)
	p.cache.Set(proxy.ToString(), res, cache.DefaultExpiration)
	return res
}

func (p *ProxyGenerator) do(proxy models.Proxy) {
	if p.verifyWithCache(proxy) {
		p.proxy <- proxy
	}
}

func (p *ProxyGenerator) worker() {
	for proxy := range p.job {
		p.do(proxy)
	}
}

func (p *ProxyGenerator) deleteOld(hour int) {
	val := p.queue.Values()
	for _, s := range val {
		proxy := s.(models.Proxy)
		if proxy.CreatedAt.Hour() < hour {
			p.queue.Dequeue()
		}
	}
	return
}

func (p *ProxyGenerator) Run(workers int, hours int) {
	go func() {
		instance.deleteOld(hours + 12)
	}()
	go p.load()

	for w := 1; w <= workers; w++ {
		go p.worker()
	}
}

func New(workers int, cacheminutes time.Duration, hoursRun int, feedList []string) *ProxyGenerator {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	queue := aq.New()
	once.Do(func() {
		instance = &ProxyGenerator{
			cache:    cache.New(cacheminutes*time.Minute, 5*time.Minute),
			VerifyFn: verifyProxy,
			proxy:    make(chan models.Proxy),
			logger:   logger,
			queue:    queue,
			job:      make(chan models.Proxy, 100),
		}

		// add providers to generator
		instance.AddProvider(providers.NewFreeProxyList())
		instance.AddProvider(providers.NewXseoIn())
		instance.AddProvider(providers.NewProxyList())
		instance.AddProvider(providers.NewTxtDomains(logger, feedList))
		instance.AddProvider(providers.NewHidemyName())
		instance.AddProvider(providers.NewCoolProxy())
		instance.AddProvider(providers.NewPubProxy())
		// run workers
		go instance.Run(workers, hoursRun)

	})
	return instance
}
