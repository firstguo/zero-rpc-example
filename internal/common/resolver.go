package common

import (
	"encoding/json"
	"strings"

	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
)

const SchemeDiscovJSON = "discovj"

type metaAttrKey struct{}

func init() {
	resolver.Register(&discovJSONBuilder{})
}

type discovJSONBuilder struct{}

func (b *discovJSONBuilder) Build(target resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (resolver.Resolver, error) {
	hosts := strings.FieldsFunc(target.URL.Host, func(r rune) bool {
		return r == ','
	})
	key := target.URL.Path
	if key == "" {
		key = target.Endpoint()
	}

	sub, err := discov.NewSubscriber(hosts, key)
	if err != nil {
		return nil, err
	}

	update := func() {
		vals := sub.Values()
		addrs := make([]resolver.Address, 0, len(vals))
		for _, val := range vals {
			addr, meta := parseEndpointValue(val)
			a := resolver.Address{Addr: addr}
			if len(meta) > 0 {
				a.Attributes = attributes.New(metaAttrKey{}, meta)
			}
			addrs = append(addrs, a)
		}
		if err := cc.UpdateState(resolver.State{
			Addresses: addrs,
		}); err != nil {
			logx.Error(err)
		}
	}
	sub.AddListener(update)
	update()

	return &discovJSONResolver{sub: sub}, nil
}

func (b *discovJSONBuilder) Scheme() string {
	return SchemeDiscovJSON
}

type discovJSONResolver struct {
	sub *discov.Subscriber
}

func (r *discovJSONResolver) Close() {
	r.sub.Close()
}

func (r *discovJSONResolver) ResolveNow(_ resolver.ResolveNowOptions) {}

func parseEndpointValue(val string) (addr string, meta map[string]string) {
	val = strings.TrimSpace(val)
	if len(val) == 0 {
		return "", nil
	}

	if val[0] != '{' {
		return val, nil
	}

	var ep ServiceEndpoint
	if err := json.Unmarshal([]byte(val), &ep); err != nil {
		return val, nil
	}

	return ep.Addr, ep.Meta
}

func MetaFromAddress(addr resolver.Address) map[string]string {
	if addr.Attributes == nil {
		return nil
	}
	v := addr.Attributes.Value(metaAttrKey{})
	if v == nil {
		return nil
	}
	if m, ok := v.(map[string]string); ok {
		return m
	}
	return nil
}
