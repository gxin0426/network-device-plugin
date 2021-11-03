package server

import (
	"context"
	"net"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
)

var (
	TotalBytes int
)

const (
	resourceName   string = "9nml.com/netIO"
	easyalgoSocket string = "easyalgo.sock"
	// KubeletSocket kubelet 监听 unix 的名称
	KubeletSocket string = "kubelet.sock"
	// DevicePluginPath 默认位置
	DevicePluginPath string = "/var/lib/kubelet/device-plugins/"
)

// EasyalgoServer 是一个 device plugin server
type EasyalgoServer struct {
	srv         *grpc.Server
	devices     map[string]*pluginapi.Device
	notify      chan bool
	ctx         context.Context
	cancel      context.CancelFunc
	restartFlag bool // 本次是否是重启
}

// NewEasyalgoServer 实例化 easyalgoServer
func NewEasyalgoServer() *EasyalgoServer {
	ctx, cancel := context.WithCancel(context.Background())
	return &EasyalgoServer{
		devices:     make(map[string]*pluginapi.Device),
		srv:         grpc.NewServer(grpc.EmptyServerOption{}),
		notify:      make(chan bool),
		ctx:         ctx,
		cancel:      cancel,
		restartFlag: false,
	}
}

// Run 运行服务
func (s *EasyalgoServer) Run() error {

	pluginapi.RegisterDevicePluginServer(s.srv, s)
	//删除easyalgoSocket
	err := syscall.Unlink(DevicePluginPath + easyalgoSocket)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	//监听
	l, err := net.Listen("unix", DevicePluginPath+easyalgoSocket)
	if err != nil {
		return err
	}

	go func() {
		lastCrashTime := time.Now()
		restartCount := 0
		for {
			log.Printf("start GPPC server for '%s'", resourceName)
			//grpc
			err = s.srv.Serve(l)
			if err == nil {
				break
			}

			log.Printf("GRPC server for '%s' crashed with error: $v", resourceName, err)

			if restartCount > 5 {
				log.Fatal("GRPC server for '%s' has repeatedly crashed recently. Quitting", resourceName)
			}
			timeSinceLastCrash := time.Since(lastCrashTime).Seconds()
			lastCrashTime = time.Now()
			if timeSinceLastCrash > 3600 {
				restartCount = 1
			} else {
				restartCount++
			}
		}
	}()

	// Wait for server to start by lauching a blocking connection
	conn, err := s.dial(easyalgoSocket, 5*time.Second)
	if err != nil {
		return err
	}
	conn.Close()

	return nil
}

// RegisterToKubelet 向kubelet注册device plugin
func (s *EasyalgoServer) RegisterToKubelet() error {
	socketFile := filepath.Join(DevicePluginPath + KubeletSocket)

	conn, err := s.dial(socketFile, 5*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pluginapi.NewRegistrationClient(conn)
	req := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     path.Base(DevicePluginPath + easyalgoSocket),
		ResourceName: resourceName,
	}
	log.Infof("Register to kubelet with endpoint %s", req.Endpoint)
	_, err = client.Register(context.Background(), req)
	if err != nil {
		return err
	}

	return nil
}

// GetDevicePluginOptions returns options to be communicated with Device
// Manager
func (s *EasyalgoServer) GetDevicePluginOptions(ctx context.Context, e *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	log.Infoln("GetDevicePluginOptions called")
	return &pluginapi.DevicePluginOptions{PreStartRequired: true}, nil
}

// ListAndWatch returns a stream of List of Devices
// Whenever a Device state change or a Device disappears, ListAndWatch
// returns the new list
func (s *EasyalgoServer) ListAndWatch(e *pluginapi.Empty, srv pluginapi.DevicePlugin_ListAndWatchServer) error {

	log.Infoln("ListAndWatch called")

	devs := make([]*pluginapi.Device, TotalBytes)
	for i := 0; i < TotalBytes; i++ {
		devs[i] = &pluginapi.Device{
			ID:	"networkdevice"  + strconv.Itoa(i + 500),
			Health: pluginapi.Healthy,
		}
	}

	err := srv.Send(&pluginapi.ListAndWatchResponse{Devices: devs})
	if err != nil {
		log.Errorf("ListAndWatch send device error: %v", err)
		return err
	}
	ticker := time.NewTicker(120 * time.Second)
	// 更新 device list
	for {
		log.Infoln("waiting for device change")
		select {
		case <-ticker.C:
			log.Infoln("当前节点可用流量 : ", TotalBytes, "MB/s")
			devs := make([]*pluginapi.Device, TotalBytes)
			for i := 0; i < TotalBytes; i++ {
				devs[i] = &pluginapi.Device{
					ID:     "networkdevice"  + strconv.Itoa(i + 500),
					Health: pluginapi.Healthy,
				}
			}

			srv.Send(&pluginapi.ListAndWatchResponse{Devices: devs})

		case <-s.ctx.Done():
			log.Info("ListAndWatch exit")
			return nil
		}
	}
}

// Allocate is called during container creation so that the Device
// Plugin can run device specific operations and instruct Kubelet
// of the steps to make the Device available in the container
func (s *EasyalgoServer) Allocate(ctx context.Context, reqs *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	log.Infoln("Allocate called")
	resps := &pluginapi.AllocateResponse{}
	return resps, nil
}

// PreStartContainer is called, if indicated by Device Plugin during registeration phase,
// before each container start. Device plugin can run device specific operations
// such as reseting the device before making devices available to the container
func (s *EasyalgoServer) PreStartContainer(ctx context.Context, req *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	log.Infoln("PreStartContainer called")
	return &pluginapi.PreStartContainerResponse{}, nil
}

func (s *EasyalgoServer) dial(unixSocketPath string, timeout time.Duration) (*grpc.ClientConn, error) {
	c, err := grpc.Dial(unixSocketPath, grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithTimeout(timeout),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}),
	)

	if err != nil {
		return nil, err
	}
	return c, nil
}
