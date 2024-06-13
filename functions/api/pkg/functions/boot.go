package functions

import "log"

func initLima() {
	log.Printf("initializing lima VM platform")
	client, err := newClient()
	if err != nil {
		panic("failed to initialize lima/containerd client")
	}

	stdOut, stdErr, err := client.limactl("ls")
	// output looks like
	// NAME       STATUS     SSH                VMTYPE    ARCH       CPUS    MEMORY    DISK      DIR
	// default    Running    127.0.0.1:60022    qemu      aarch64    4       4GiB      100GiB    ~/.lima/default
	if err != nil {
		log.Fatalf("failed to list lima VMs: %v\n%s\n%s", err, stdOut, stdErr)
	}

	if stdOut == "" {
		log.Println("no lima containers found, creating one")
		stdOut, stdErr, err = client.limactl("--tty=false", "create", "default")
		if err != nil {
			log.Fatalf("failed to create lima VM: %v\n%s\n%s", err, stdOut, stdErr)
		}
	}

	stdOut, stdErr, err = client.limactl("start", "default")
	if err != nil {
		log.Fatalf("failed to start lima VM: %v\n%s\n%s", err, stdOut, stdErr)
	}
}
