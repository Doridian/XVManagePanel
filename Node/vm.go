package main
import (
	"github.com/alexzorin/libvirt-go"
	"log"
	"time"
)

type VMDomain struct {
	name string
	vmid uint32
	
	poweredOn bool

	cpuUsage float64
	ramUsage float64
	vcpus int64
	
	removePossible bool
	
	lastCpuTime uint64
	lastCheck time.Time
}

func libvirtClose(virConn libvirt.VirConnection) {
	closeRes := 1
	var err error
	for closeRes > 0 {
		closeRes, err = virConn.CloseConnection()
		if err != nil {
			log.Printf("close: %v", err)
		}
	}
}

func initializeLibvirt() {
	vmDomains.m = make(map[string]VMDomain)
	maintainVMList()
	go maintainVMListTicker()
}

func getLibvirtConnection() libvirt.VirConnection {
	virConn, err := libvirt.NewVirConnection("qemu:///system")
	if err != nil {
		log.Printf("Libvirt load: error: %v", err)
	}
	return virConn
}

func getLibvirtDomain(virConn libvirt.VirConnection, name string) libvirt.VirDomain {
	virDomain, _ := virConn.LookupDomainByName(name)
	return virDomain
}

func findVMDomainByName(name string) *VMDomain {
	vmDomains.RLock()
	vmDomain, found :=  vmDomains.m[name]
	vmDomains.RUnlock()
	if !found {
		return nil
	}
	return &vmDomain
}

func vmGetStatus(name string) VMStatus {
	return findVMDomainByName(name).makeVMStatus()
}

func vmProcessCommand(name string, command string) {
	virConn := getLibvirtConnection()
	defer libvirtClose(virConn)	

	virDomain := getLibvirtDomain(virConn, name)
	switch command {
		case "reset":
			virDomain.Destroy()
			virDomain.Create()
		case "start":
			virDomain.Create()
		case "destroy":
			virDomain.Destroy()
		case "shutdown":
			virDomain.Shutdown()
		case "reboot":
			virDomain.Reboot()
	}
}

func vmList() []VMStatus {
	statusRes := make([]VMStatus, 0)
	vmDomains.RLock()
	for _, vmDomain := range vmDomains.m {
		statusRes = append(statusRes, vmDomain.makeVMStatus())
	}
	vmDomains.RUnlock()
	return statusRes
}
