package main

type FirecrackerMVMManager {
	RunningVMs map[string]RunningFirecracker
}

func (vmm *FirecrackerMVMManager) Init() {
	vmm.RunningVMs = make(map[string]*RunningFirecracker)
}

func (vmm *FirecrackerMVMManager) Add(mvm *RunningFirecracker) (nil, err) {

}

func (vmm *FirecrackerMVMManager) Remove(mvm *RunningFirecracker) (nil, err) {

}

func (vmm *FirecrackerMVMManager) Shutdown(mvm *RunningFirecracker) (nil, err) {

}

func (vmm *FirecrackerMVMManager) ShutdownAll(mvm *RunningFirecracker) (nil, err) {

}
