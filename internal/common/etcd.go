package common

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/netx"
	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceEndpoint is the JSON structure stored in etcd values.
type ServiceEndpoint struct {
	Addr string            `json:"addr"`
	Meta map[string]string `json:"meta,omitempty"`
}

// ApplyEnvAndTagRouting applies environment prefix to etcd key for non-prod environments.
func ApplyEnvAndTagRouting(conf *zrpc.RpcServerConf, _ map[string]string) {
	if conf.Etcd.Key == "" {
		return
	}

	mode := strings.ToLower(conf.Mode)
	if mode != "" && mode != "prod" && mode != "production" {
		if !strings.HasPrefix(conf.Etcd.Key, "/"+mode+"/") {
			conf.Etcd.Key = "/" + mode + "/" + strings.TrimPrefix(conf.Etcd.Key, "/")
		}
	}
}

// RegisterServiceJSON registers the service to etcd with a JSON value containing addr and meta.
// Returns a stop function to deregister.
func RegisterServiceJSON(conf zrpc.RpcServerConf, meta map[string]string) func() {
	if conf.Etcd.Key == "" {
		return func() {}
	}

	endpoint := ServiceEndpoint{
		Addr: figureOutListenOn(conf.ListenOn),
		Meta: meta,
	}
	value, err := json.Marshal(endpoint)
	if err != nil {
		logx.Errorf("failed to marshal service endpoint: %v", err)
		return func() {}
	}

	var pubOpts []discov.PubOption
	if conf.Etcd.HasAccount() {
		pubOpts = append(pubOpts, discov.WithPubEtcdAccount(conf.Etcd.User, conf.Etcd.Pass))
	}
	if conf.Etcd.HasTLS() {
		pubOpts = append(pubOpts, discov.WithPubEtcdTLS(
			conf.Etcd.CertFile, conf.Etcd.CertKeyFile, conf.Etcd.CACertFile, conf.Etcd.InsecureSkipVerify))
	}
	if conf.Etcd.HasID() {
		pubOpts = append(pubOpts, discov.WithId(conf.Etcd.ID))
	}

	pub := discov.NewPublisher(conf.Etcd.Hosts, conf.Etcd.Key, string(value), pubOpts...)
	if err := pub.KeepAlive(); err != nil {
		logx.Errorf("failed to register service to etcd: %v", err)
		return func() {}
	}

	logx.Infof("registered service [%s] with value: %s", conf.Etcd.Key, string(value))

	return func() {
		pub.Stop()
	}
}

func figureOutListenOn(listenOn string) string {
	fields := strings.Split(listenOn, ":")
	if len(fields) == 0 {
		return listenOn
	}

	host := fields[0]
	if len(host) > 0 && host != "0.0.0.0" {
		return listenOn
	}

	ip := os.Getenv("POD_IP")
	if len(ip) == 0 {
		ip = netx.InternalIp()
	}
	if len(ip) == 0 {
		return listenOn
	}

	return strings.Join(append([]string{ip}, fields[1:]...), ":")
}
