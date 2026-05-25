package common

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/syncx"
	"github.com/zeromicro/go-zero/core/timex"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/status"
)

const (
	Name = "tag_p2c"

	decayTime       = int64(time.Second * 10)
	forcePick       = int64(time.Second)
	initSuccess     = 1000
	throttleSuccess = initSuccess / 2
	penalty         = int64(math.MaxInt32)
	pickTimes       = 3
	logInterval     = time.Minute
)

var emptyPickResult balancer.PickResult

type tagsKey struct{}

// WithTags sets routing tags into context for per-request tag filtering.
func WithTags(ctx context.Context, tags map[string]string) context.Context {
	return context.WithValue(ctx, tagsKey{}, tags)
}

// tagsFromContext extracts routing tags from context.
func tagsFromContext(ctx context.Context) map[string]string {
	tags, _ := ctx.Value(tagsKey{}).(map[string]string)
	return tags
}

func init() {
	balancer.Register(base.NewBalancerBuilder(Name, &tagP2cPickerBuilder{}, base.Config{HealthCheck: true}))
}

type tagP2cPickerBuilder struct{}

func (b *tagP2cPickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	readySCs := info.ReadySCs
	if len(readySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}

	conns := make([]*subConn, 0, len(readySCs))
	for conn, connInfo := range readySCs {
		conns = append(conns, &subConn{
			addr:    connInfo.Address,
			conn:    conn,
			success: initSuccess,
			tags:    metaToValues(MetaFromAddress(connInfo.Address)),
		})
	}

	return &tagP2cPicker{
		conns: conns,
		r:     rand.New(rand.NewSource(time.Now().UnixNano())),
		stamp: syncx.NewAtomicDuration(),
	}
}

func metaToValues(meta map[string]string) url.Values {
	if len(meta) == 0 {
		return nil
	}
	values := make(url.Values)
	for k, v := range meta {
		values.Set(k, v)
	}
	return values
}

type tagP2cPicker struct {
	conns []*subConn
	r     *rand.Rand
	stamp *syncx.AtomicDuration
	lock  sync.Mutex
}

func (p *tagP2cPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	// Filter by request-level tags from context
	reqTags := tagsFromContext(info.Ctx)
	candidates := p.filterByTags(reqTags)

	if len(candidates) == 0 {
		return emptyPickResult, balancer.ErrNoSubConnAvailable
	}

	var chosen *subConn
	switch len(candidates) {
	case 1:
		chosen = p.choose(candidates[0], nil)
	case 2:
		chosen = p.choose(candidates[0], candidates[1])
	default:
		var node1, node2 *subConn
		for i := 0; i < pickTimes; i++ {
			a := p.r.Intn(len(candidates))
			b := p.r.Intn(len(candidates) - 1)
			if b >= a {
				b++
			}
			node1 = candidates[a]
			node2 = candidates[b]
			if node1.healthy() && node2.healthy() {
				break
			}
		}
		chosen = p.choose(node1, node2)
	}

	atomic.AddInt64(&chosen.inflight, 1)
	atomic.AddInt64(&chosen.requests, 1)

	return balancer.PickResult{
		SubConn: chosen.conn,
		Done:    p.buildDoneFunc(chosen),
	}, nil
}

func (p *tagP2cPicker) filterByTags(reqTags map[string]string) []*subConn {
	if len(reqTags) == 0 {
		return p.conns
	}

	filtered := make([]*subConn, 0, len(p.conns))
	for _, c := range p.conns {
		if matchTags(c.tags, reqTags) {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

func matchTags(addrTags url.Values, reqTags map[string]string) bool {
	if len(addrTags) == 0 {
		return false
	}
	for k, v := range reqTags {
		if addrTags.Get(k) != v {
			return false
		}
	}
	return true
}

func (p *tagP2cPicker) buildDoneFunc(c *subConn) func(info balancer.DoneInfo) {
	start := int64(timex.Now())
	return func(info balancer.DoneInfo) {
		atomic.AddInt64(&c.inflight, -1)
		now := timex.Now()
		last := atomic.SwapInt64(&c.last, int64(now))
		td := int64(now) - last
		if td < 0 {
			td = 0
		}

		w := math.Exp(float64(-td) / float64(decayTime))
		lag := int64(now) - start
		if lag < 0 {
			lag = 0
		}
		olag := atomic.LoadUint64(&c.lag)
		if olag == 0 {
			w = 0
		}

		atomic.StoreUint64(&c.lag, uint64(float64(olag)*w+float64(lag)*(1-w)))
		success := initSuccess
		if info.Err != nil && !acceptable(info.Err) {
			success = 0
		}
		osucc := atomic.LoadUint64(&c.success)
		atomic.StoreUint64(&c.success, uint64(float64(osucc)*w+float64(success)*(1-w)))

		stamp := p.stamp.Load()
		if now-stamp >= logInterval {
			if p.stamp.CompareAndSwap(stamp, now) {
				p.logStats()
			}
		}
	}
}

func (p *tagP2cPicker) choose(c1, c2 *subConn) *subConn {
	start := int64(timex.Now())
	if c2 == nil {
		atomic.StoreInt64(&c1.pick, start)
		return c1
	}

	if c1.load() > c2.load() {
		c1, c2 = c2, c1
	}

	pick := atomic.LoadInt64(&c2.pick)
	if start-pick > forcePick && atomic.CompareAndSwapInt64(&c2.pick, pick, start) {
		return c2
	}

	atomic.StoreInt64(&c1.pick, start)
	return c1
}

func (p *tagP2cPicker) logStats() {
	stats := make([]string, 0, len(p.conns))
	for _, conn := range p.conns {
		stats = append(stats, fmt.Sprintf("conn: %s, load: %d, reqs: %d",
			conn.addr.Addr, conn.load(), atomic.SwapInt64(&conn.requests, 0)))
	}
	logx.Statf("tag_p2c - %s", strings.Join(stats, "; "))
}

type subConn struct {
	lag      uint64
	inflight int64
	success  uint64
	requests int64
	last     int64
	pick     int64
	addr     resolver.Address
	conn     balancer.SubConn
	tags     url.Values
}

func (c *subConn) healthy() bool {
	return atomic.LoadUint64(&c.success) > throttleSuccess
}

func (c *subConn) load() int64 {
	lag := int64(math.Sqrt(float64(atomic.LoadUint64(&c.lag) + 1)))
	load := lag * (atomic.LoadInt64(&c.inflight) + 1)
	if load == 0 {
		return penalty
	}
	return load
}

func acceptable(err error) bool {
	switch status.Code(err) {
	case codes.DeadlineExceeded, codes.Internal, codes.Unavailable, codes.DataLoss,
		codes.Unimplemented, codes.ResourceExhausted:
		return false
	default:
		return true
	}
}
