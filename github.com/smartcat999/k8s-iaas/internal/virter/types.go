package virter

import (
	libvirtxml "github.com/libvirt/libvirt-go-xml"
)

type Domain struct {
	libvirtxml.Domain
	Status int32
}
