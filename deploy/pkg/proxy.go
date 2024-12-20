package pkg

import (
	"fmt"
	"strings"
)

type ProxyService struct {
	Name    string
	Domains []string
	Port    string // Format like "80:3000"
	SSL     bool
	Networks []string
}

type ProxyManager struct {
	ssh      *SSHClient
	proxyBin string
}

func NewProxyManager(ssh *SSHClient, proxyBin string) *ProxyManager {
	return &ProxyManager{
		ssh:      ssh,
		proxyBin: proxyBin,
	}
}

func (p *ProxyManager) DeployService(service ProxyService) error {
	ports := strings.Split(service.Port, ":")
	if len(ports) != 2 {
		return fmt.Errorf("invalid port mapping format: %s", service.Port)
	}

	// For each domain, deploy to kamal-proxy
	for _, domain := range service.Domains {
		// Build deploy command with our service settings
		cmd := strings.Builder{}
		cmd.WriteString(fmt.Sprintf("%s deploy %s --target localhost:%s --host %s", 
			p.proxyBin, 
			service.Name,
			ports[1],  // Use target port
			domain,
		))

		if service.SSL {
			cmd.WriteString(" --tls")
		}

		// Execute the deploy command
		execCmd := NewRemoteCommand(p.ssh, cmd.String())
		if err := execCmd.Run(); err != nil {
			return fmt.Errorf("failed to deploy proxy config for %s: %w", domain, err)
		}
	}

	return nil
}

func (p *ProxyManager) RemoveService(serviceName string) error {
	cmd := NewRemoteCommand(p.ssh, fmt.Sprintf(
		"%s remove %s",
		p.proxyBin,
		serviceName,
	))

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove service %s: %w", serviceName, err)
	}

	return nil
}
