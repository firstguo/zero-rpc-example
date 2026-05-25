package common

import (
	"encoding/json"
	"net"
	"reflect"
	"time"
	"unsafe"

	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
)

// HookServerRegistration uses reflection to replace go-zero's internal registerEtcd
// function with our JSON-based registration. The registration happens in a goroutine
// that waits for the port to be ready, ensuring services are discoverable only after
// they can accept connections.
func HookServerRegistration(server *zrpc.RpcServer, conf zrpc.RpcServerConf, meta map[string]string) func() {
	if conf.Etcd.Key == "" {
		return func() {}
	}

	// Access RpcServer.server (internal.Server interface)
	serverVal := reflect.ValueOf(server).Elem()
	internalServerField := serverVal.FieldByName("server")
	if !internalServerField.IsValid() {
		logx.Error("failed to access internal server field")
		return func() {}
	}

	// The interface value contains a keepAliveServer (value type, not pointer).
	// We need to get the underlying data pointer from the interface.
	// Interface layout: [type pointer, data pointer]
	type iface struct {
		typ  unsafe.Pointer
		data unsafe.Pointer
	}

	ifacePtr := (*iface)(unsafe.Pointer(internalServerField.UnsafeAddr()))
	if ifacePtr.data == nil {
		logx.Error("internal server interface data is nil")
		return func() {}
	}

	// keepAliveServer struct layout:
	//   registerEtcd func() error  (first field)
	//   Server       internal.Server (embedded, second field)
	// The registerEtcd field is at offset 0
	registerEtcdPtr := (*func() error)(ifacePtr.data)

	// Create our custom registration function
	var publisher *discov.Publisher
	customRegister := func() error {
		go func() {
			listenAddr := figureOutListenOn(conf.ListenOn)

			if err := waitForPort(listenAddr, 10*time.Second); err != nil {
				logx.Errorf("failed to wait for port %s: %v", listenAddr, err)
				return
			}

			endpoint := ServiceEndpoint{
				Addr: listenAddr,
				Meta: meta,
			}
			value, err := json.Marshal(endpoint)
			if err != nil {
				logx.Errorf("failed to marshal service endpoint: %v", err)
				return
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

			publisher = discov.NewPublisher(conf.Etcd.Hosts, conf.Etcd.Key, string(value), pubOpts...)
			if err := publisher.KeepAlive(); err != nil {
				logx.Errorf("failed to register service to etcd: %v", err)
				return
			}

			logx.Infof("registered service [%s] with JSON value: %s", conf.Etcd.Key, string(value))
		}()

		return nil
	}

	// Replace the registerEtcd function pointer
	*registerEtcdPtr = customRegister

	return func() {
		if publisher != nil {
			publisher.Stop()
		}
	}
}

func waitForPort(addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return nil
}
