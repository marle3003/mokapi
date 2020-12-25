package middlewares

import (
	"math"
	"math/rand"
	"mokapi/models"
	"mokapi/server/web"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type delay struct {
	config *models.Delay
	next   Middleware
	random *rand.Rand
}

func NewDelay(config *models.Delay, next Middleware) Middleware {
	return &delay{config: config, next: next, random: rand.New(rand.NewSource(time.Now().UnixNano()))}
}

func (d *delay) ServeData(request *Request, context *web.HttpContext) {
	switch strings.ToLower(d.config.Distribution.Type) {
	case "lognormal":
		number := math.Round(math.Exp(d.random.NormFloat64()*d.config.Distribution.Sigma) * d.config.Distribution.Median)
		ms := time.Duration(number) * time.Millisecond
		log.Infof("Delay request for %v", ms)
		time.Sleep(ms)
	case "uniform":
		number := rand.Intn(d.config.Distribution.Upper-d.config.Distribution.Lower) + d.config.Distribution.Lower
		ms := time.Duration(number) * time.Millisecond
		log.Infof("Delay request for %v", ms)
		time.Sleep(ms)
	case "":
		log.Infof("Delay request for %v", d.config.Fixed)
		time.Sleep(d.config.Fixed)
	}

	d.next.ServeData(request, context)
}
