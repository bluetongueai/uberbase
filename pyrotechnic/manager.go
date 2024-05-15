package main

import "os"

type FirecrackerMVMManager struct {
	RunningVMs map[string]RunningFirecracker
}

func (vmm FirecrackerMVMManager) Get(id string) (RunningFirecracker, error) {
	return vmm.RunningVMs[id], nil
}

func (vmm FirecrackerMVMManager) Add(id string, mvm RunningFirecracker) error {
	vmm.RunningVMs[id] = mvm
	return nil
}

func (vmm FirecrackerMVMManager) Remove(id string) error {
	vmm.Shutdown(id)
	delete(vmm.RunningVMs, id)
	return nil
}

func (vmm FirecrackerMVMManager) Shutdown(id string) error {
	vm := vmm.RunningVMs[id]
	vm.machine.StopVMM()
	os.Remove(vm.image)
	return nil
}

func (vmm FirecrackerMVMManager) ShutdownAll() error {
	for id := range vmm.RunningVMs {
		vmm.Shutdown(id)
	}
	return nil
}
