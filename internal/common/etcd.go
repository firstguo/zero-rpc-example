package common

import (
	"net/url"
	"strings"

	"github.com/zeromicro/go-zero/zrpc"
)

// ApplyEnvAndTagRouting applies environment prefix and tag routing to etcd configuration
// 1. For non-prod environments (dev/test), adds environment prefix to etcd key: /test/...
// 2. Appends Meta as query parameters to ListenOn: 127.0.0.1:8080?pdb=tripo_studio_enhance_v3
func ApplyEnvAndTagRouting(conf *zrpc.RpcServerConf, meta map[string]string) {
	if conf.Etcd.Key == "" {
		return
	}

	// 1. Apply environment prefix for non-prod environments
	mode := strings.ToLower(conf.Mode)
	if mode != "" && mode != "prod" && mode != "production" {
		// Add environment prefix if not already present
		if !strings.HasPrefix(conf.Etcd.Key, "/"+mode+"/") {
			conf.Etcd.Key = "/" + mode + "/" + strings.TrimPrefix(conf.Etcd.Key, "/")
		}
	}

	// 2. Apply tag routing by appending Meta as query parameters to ListenOn
	if len(meta) > 0 && conf.ListenOn != "" {
		params := url.Values{}
		for k, v := range meta {
			params.Add(k, v)
		}

		// Check if ListenOn already has query parameters
		if strings.Contains(conf.ListenOn, "?") {
			conf.ListenOn = conf.ListenOn + "&" + params.Encode()
		} else {
			conf.ListenOn = conf.ListenOn + "?" + params.Encode()
		}
	}
}
