package main

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/firecracker-microvm/firecracker-go-sdk"
)

type RunningFirecracker struct {
	ctx       context.Context
	cancelCtx context.CancelFunc
	image     string
	machine   *firecracker.Machine
}

type FirecrackerOptions struct {
	Id string `long:"id" description:"Jailer VMM id"`
	// maybe make this an int instead
	IpId            byte   `byte:"id" description:"an ip we use to generate an ip address"`
	FcBinary        string `long:"firecracker-binary" description:"Path to firecracker binary"`
	FcKernelCmdLine string `long:"kernel-opts" description:"Kernel commandline"`
	Request         CreateRequest
	FcSocketPath    string `long:"socket-path" short:"s" description:"path to use for firecracker socket"`
	TapMacAddr      string `long:"tap-mac-addr" description:"tap macaddress"`
	TapDev          string `long:"tap-dev" description:"tap device"`
	FcCPUCount      int64  `long:"ncpus" short:"c" description:"Number of CPUs"`
	FcMemSz         int64  `long:"memory" short:"m" description:"VM memory, in MiB"`
	FcIP            string `long:"fc-ip" description:"IP address of the VM"`
}

// build firecracker vm options from a create request
// numbers the socket
func GetOptions(id byte, req CreateRequest) FirecrackerOptions {
	fc_ip := net.IPv4(172, 102, 0, id).String()
	gateway_ip := "172.102.0.1"
	docker_mask_long := "255.255.255.0"
	bootArgs := "ro console=ttyS0 noapic reboot=k panic=1 pci=off nomodules random.trust_cpu=on "
	bootArgs = bootArgs + fmt.Sprintf("ip=%s::%s:%s::eth0:off", fc_ip, gateway_ip, docker_mask_long)
	socket_dir := os.Getenv("FIRECRACKER_SOCKET_DIR")
	return FirecrackerOptions{
		FcBinary:        "/usr/bin/firecracker",
		Request:         req,
		FcKernelCmdLine: bootArgs,
		FcSocketPath:    fmt.Sprintf("%s/firecracker-%d.sock", socket_dir, id),
		TapMacAddr:      fmt.Sprintf("02:FC:00:00:00:%02x", id),
		TapDev:          fmt.Sprintf("fc-tap-%d", id),
		FcIP:            fc_ip,
		FcCPUCount:      1,
		FcMemSz:         512,
	}
}
