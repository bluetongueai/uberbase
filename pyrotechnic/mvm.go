package main

import (
	"context"

	firecracker "github.com/firecracker-microvm/firecracker-go-sdk"
)

func Create(opts FirecrackerOptions, ctx context.Context) (nil, error) {
	vmmCtx, vmmCancel := context.WithCancel(ctx)
	rootImagePath, err := copyImage(opts.Request.RootDrivePath)
	opts.Request.RootDrivePath = rootImagePath
	if err != nil {
		return nil, fmt.Errorf("failed copying root path: %s", err)
	}
	fcCfg, err := getConfig(&opts)
	if err != nil {
		return nil, err
	}

	cmd := firecracker.VMCommandBuilder{}.
		WithBin(opts.FcBinary).
		WithSocketPath(fcCfg.SocketPath).
		WithStdin(os.Stdin).
		WithStdout(os.Stdout).
		WithStderr(os.Stderr).
		Build(ctx)

	machineOpts := []firecracker.Opt{
		firecracker.WithProcessRunner(cmd),
	}
	exec.Command("ip", "link", "del", opts.TapDev).Run()
	if err := exec.Command("ip", "tuntap", "add", "dev", opts.TapDev, "mode", "tap").Run(); err != nil {
		return nil, fmt.Errorf("failed creating ip link: %s", err)
	}
	if err := exec.Command("rm", "-f", opts.FcSocketPath).Run(); err != nil {
		return nil, fmt.Errorf("failed to delete old socket path: %s", err)
	}
	if err := exec.Command("ip", "link", "set", opts.TapDev, "master", "firecracker0").Run(); err != nil {
		return nil, fmt.Errorf("failed adding tap device to bridge: %s", err)
	}
	if err := exec.Command("ip", "link", "set", opts.TapDev, "up").Run(); err != nil {
		return nil, fmt.Errorf("failed creating ip link: %s", err)
	}
	if err := exec.Command("sysctl", "-w", fmt.Sprintf("net.ipv4.conf.%s.proxy_arp=1", opts.TapDev)).Run(); err != nil {
		return nil, fmt.Errorf("failed doing first sysctl: %s", err)
	}
	if err := exec.Command("sysctl", "-w", fmt.Sprintf("net.ipv6.conf.%s.disable_ipv6=1", opts.TapDev)).Run(); err != nil {
		return nil, fmt.Errorf("failed doing second sysctl: %s", err)
	}
	m, err := firecracker.NewMachine(vmmCtx, *fcCfg, machineOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed creating machine: %s", err)
	}
	if err := m.Start(vmmCtx); err != nil {
		return nil, fmt.Errorf("failed to start machine: %v", err)
	}
	installSignalHandlers(vmmCtx, m)
	return &RunningFirecracker{
		ctx:       vmmCtx,
		image:     rootImagePath,
		cancelCtx: vmmCancel,
		machine:   m,
	}, nil
}

func copyImage(src string) (string, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return "", err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return "", fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer source.Close()

	destination, err := os.CreateTemp("/images", "image")
	if err != nil {
		return "", err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return destination.Name(), err
}

func installSignalHandlers(ctx context.Context, m *firecracker.Machine) {
	// not sure if this is actually really helping with anything
	go func() {
		// Clear some default handlers installed by the firecracker SDK:
		signal.Reset(os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

		for {
			switch s := <-c; {
			case s == syscall.SIGTERM || s == os.Interrupt:
				log.Printf("Caught SIGINT, requesting clean shutdown")
				m.Shutdown(ctx)
			case s == syscall.SIGQUIT:
				log.Printf("Caught SIGTERM, forcing shutdown")
				m.StopVMM()
			}
		}
	}()
}

func getConfig(opts *FirecrackerOptions) (*firecracker.Config, error) {
	drives := []models.Drive{
		models.Drive{
			DriveID:      firecracker.String("1"),
			PathOnHost:   &opts.Request.RootDrivePath,
			IsRootDevice: firecracker.Bool(true),
			IsReadOnly:   firecracker.Bool(false),
		},
	}
	if opts.Request.CloudInitPath != "" {
		isoPath, err := makeIso(opts.Request.CloudInitPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create iso: %s", err)
		}
		drives = append(drives, models.Drive{
			DriveID:      firecracker.String("2"),
			PathOnHost:   &isoPath,
			IsRootDevice: firecracker.Bool(false),
			IsReadOnly:   firecracker.Bool(true),
		})
	}

	return &firecracker.Config{
		VMID:            opts.Id,
		SocketPath:      opts.FcSocketPath,
		KernelImagePath: opts.Request.KernelPath,
		KernelArgs:      opts.FcKernelCmdLine,
		Drives:          drives,
		NetworkInterfaces: []firecracker.NetworkInterface{
			firecracker.NetworkInterface{
				StaticConfiguration: &firecracker.StaticNetworkConfiguration{
					MacAddress:  opts.TapMacAddr,
					HostDevName: opts.TapDev,
				},
				//AllowMMDS: allowMMDS,
			},
		},
		MachineCfg: models.MachineConfiguration{
			VcpuCount:  firecracker.Int64(opts.FcCPUCount),
			MemSizeMib: firecracker.Int64(opts.FcMemSz),
			//CPUTemplate: models.CPUTemplate(opts.FcCPUTemplate),
		},
		//JailerCfg: jail,
		//VsockDevices:      vsocks,
		//LogFifo:           opts.FcLogFifo,
		//LogLevel:          opts.FcLogLevel,
		//MetricsFifo:       opts.FcMetricsFifo,
		//FifoLogWriter:     fifo,
	}, nil
}
