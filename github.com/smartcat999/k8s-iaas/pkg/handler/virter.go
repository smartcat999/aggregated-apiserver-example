package handler

import (
	"fmt"
	"net"
	"time"

	"github.com/digitalocean/go-libvirt"
	"github.com/smartcat999/k8s-iaas/internal/virter"

	"github.com/LINBIT/virter/pkg/sshkeys"
)

const (
	libvirt_pool                 = "default"
	libvirt_network              = "default"
	libvirt_static_dhcp          = false
	libvirt_disk_cache           = ""
	time_ssh_ping_count          = 300
	time_ssh_ping_period         = time.Second
	time_shutdown_timeout        = 20 * time.Second
	auth_virter_public_key_path  = "/root/.config/virter/id_rsa.pub"
	auth_virter_private_key_path = "/root/.config/virter/id_rsa"
)

// InitVirter initializes virter by connecting to the local libvirt instance and configures the ssh keystore.
func InitVirter() (*virter.Virter, error) {
	c, err := net.DialTimeout("unix", "/var/run/libvirt/libvirt-sock", 2*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to dial libvirt: %w", err)
	}

	l := libvirt.New(c)
	if err := l.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to libvirt socket: %w", err)
	}

	pool := libvirt_pool
	network := libvirt_network
	privateKeyPath := auth_virter_private_key_path
	publicKeyPath := auth_virter_public_key_path

	keyStore, err := sshkeys.NewKeyStore(privateKeyPath, publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load ssh key store: %w", err)
	}

	return virter.New(l, pool, network, keyStore), nil
}
