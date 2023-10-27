package virter_test

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	libvirt "github.com/digitalocean/go-libvirt"
	libvirtxml "github.com/libvirt/libvirt-go-xml"

	"github.com/LINBIT/virter/internal/virter"
)

type FakeLibvirtConnection struct {
	networks map[string]*FakeLibvirtNetwork
	domains  map[string]*FakeLibvirtDomain
	pools    map[string]*FakeLibvirtStoragePool
}

func (l *FakeLibvirtConnection) ConnectSupportsFeature(Feature int32) (int32, error) {
	return 1, nil
}

func (l *FakeLibvirtConnection) ConnectListAllNetworks(NeedResults int32, Flags libvirt.ConnectListAllNetworksFlags) ([]libvirt.Network, uint32, error) {
	nets := make([]libvirt.Network, 0, len(l.networks))
	for k := range l.networks {
		nets = append(nets, libvirt.Network{Name: k})
	}

	return nets, 0, nil
}

func (l *FakeLibvirtConnection) NetworkDefineXML(XML string) (libvirt.Network, error) {
	var parsed libvirtxml.Network
	err := xml.Unmarshal([]byte(XML), &parsed)
	if err != nil {
		return libvirt.Network{}, err
	}

	_, ok := l.networks[parsed.Name]
	if ok {
		return libvirt.Network{}, fmt.Errorf("network already exists")
	}

	l.networks[parsed.Name] = &FakeLibvirtNetwork{description: &parsed}
	return libvirt.Network{Name: parsed.Name}, nil
}

func (l *FakeLibvirtConnection) NetworkSetAutostart(Net libvirt.Network, Autostart int32) (err error) {
	return nil
}

func (l *FakeLibvirtConnection) NetworkCreate(Net libvirt.Network) (err error) {
	return nil
}

func (l *FakeLibvirtConnection) NetworkDestroy(Net libvirt.Network) (err error) {
	return nil
}

func (l *FakeLibvirtConnection) NetworkUndefine(Net libvirt.Network) error {
	_, ok := l.networks[Net.Name]
	if !ok {
		return libvirt.Error{Code: uint32(libvirt.ErrNoNetwork)}
	}
	delete(l.networks, Net.Name)
	return nil
}

func (l *FakeLibvirtConnection) NetworkGetDhcpLeases(Net libvirt.Network, Mac libvirt.OptString, NeedResults int32, Flags uint32) (rLeases []libvirt.NetworkDhcpLease, rRet uint32, err error) {
	return nil, 0, nil
}

type FakeLibvirtStorageVol struct {
	description *libvirtxml.StorageVolume
	content     []byte
}

type FakeLibvirtNetwork struct {
	description *libvirtxml.Network
}

type FakeLibvirtDomain struct {
	description *libvirtxml.Domain
	persistent  bool
	active      bool
}

type FakeLibvirtStoragePool struct {
	description *libvirtxml.StoragePool
	vols        map[string]*FakeLibvirtStorageVol
}

func newFakeLibvirtConnection() *FakeLibvirtConnection {
	l := &FakeLibvirtConnection{
		networks: map[string]*FakeLibvirtNetwork{networkName: fakeLibvirtNetwork()},
		domains:  make(map[string]*FakeLibvirtDomain),
		pools:    make(map[string]*FakeLibvirtStoragePool),
	}
	l.addFakeStoragePool(poolName)
	return l
}

func (l *FakeLibvirtConnection) Disconnect() error {
	return nil
}

func (l *FakeLibvirtConnection) ConnectListAllDomains(NeedResults int32, Flags libvirt.ConnectListAllDomainsFlags) (rDomains []libvirt.Domain, rRet uint32, err error) {
	domains := []libvirt.Domain{}
	for _, domain := range l.domains {
		domains = append(domains, libvirt.Domain{Name: domain.description.Name})
	}
	return domains, uint32(len(domains)), nil
}

func (l *FakeLibvirtConnection) StoragePoolLookupByName(Name string) (rPool libvirt.StoragePool, err error) {
	if _, ok := l.pools[Name]; !ok {
		return libvirt.StoragePool{}, errors.New("unknown pool")
	}
	return libvirt.StoragePool{
		Name: Name,
	}, nil
}

func (l *FakeLibvirtConnection) StoragePoolListAllVolumes(Pool libvirt.StoragePool, NeedResults int32, Flags uint32) (rVols []libvirt.StorageVol, rRet uint32, err error) {
	for _, v := range l.pools[Pool.Name].vols {
		rVols = append(rVols, libvirt.StorageVol{Name: v.description.Name, Pool: Pool.Name})
	}
	return
}

func (l *FakeLibvirtConnection) StorageVolCreateXML(Pool libvirt.StoragePool, XML string, Flags libvirt.StorageVolCreateFlags) (rVol libvirt.StorageVol, err error) {
	description := &libvirtxml.StorageVolume{}
	if err := description.Unmarshal(XML); err != nil {
		return libvirt.StorageVol{}, fmt.Errorf("invalid storage vol XML: %w", err)
	}
	l.pools[Pool.Name].vols[description.Name] = &FakeLibvirtStorageVol{
		description: description,
	}
	return libvirt.StorageVol{
		Name: description.Name,
		Pool: Pool.Name,
	}, nil
}

func (l *FakeLibvirtConnection) StorageVolDelete(Vol libvirt.StorageVol, Flags libvirt.StorageVolDeleteFlags) (err error) {
	_, ok := l.pools[Vol.Pool].vols[Vol.Name]
	if !ok {
		return libvirt.Error{Code: uint32(libvirt.ErrNoStorageVol)}
	}

	delete(l.pools[Vol.Pool].vols, Vol.Name)
	return nil
}

func (l *FakeLibvirtConnection) StorageVolLookupByName(Pool libvirt.StoragePool, Name string) (rVol libvirt.StorageVol, err error) {
	_, ok := l.pools[Pool.Name].vols[Name]
	if !ok {
		return libvirt.StorageVol{}, libvirt.Error{Code: uint32(libvirt.ErrNoStorageVol)}
	}

	return libvirt.StorageVol{
		Name: Name,
		Pool: Pool.Name,
	}, nil
}

func (l *FakeLibvirtConnection) StorageVolUpload(Vol libvirt.StorageVol, outStream io.Reader, Offset, Length uint64, Flags libvirt.StorageVolUploadFlags) (err error) {
	vol, ok := l.pools[Vol.Pool].vols[Vol.Name]
	if !ok {
		return libvirt.Error{Code: uint32(libvirt.ErrNoStorageVol)}
	}

	vol.content, err = ioutil.ReadAll(outStream)
	if err != nil {
		return errors.New("error reading upload data")
	}

	return nil
}

func (l *FakeLibvirtConnection) StorageVolGetXMLDesc(Vol libvirt.StorageVol, Flags uint32) (rXML string, err error) {
	v, ok := l.pools[Vol.Pool].vols[Vol.Name]
	if !ok {
		return "", fmt.Errorf("unknown volume %s", Vol.Name)
	}

	encoded, err := v.description.Marshal()
	if err != nil {
		return "", err
	}
	return encoded, nil
}

func (l *FakeLibvirtConnection) StorageVolCreateXMLFrom(Pool libvirt.StoragePool, XML string, Clonevol libvirt.StorageVol, Flags libvirt.StorageVolCreateFlags) (rVol libvirt.StorageVol, err error) {
	newDescription := &libvirtxml.StorageVolume{}
	if err := newDescription.Unmarshal(XML); err != nil {
		return libvirt.StorageVol{}, fmt.Errorf("invalid storage vol XML: %w", err)
	}

	oldVol, ok := l.pools[Pool.Name].vols[Clonevol.Name]
	if !ok {
		panic("nonexistent Clonevol specified")
	}

	// start off with existing definition, using only name and permissions from new XML
	description := oldVol.description
	description.Name = newDescription.Name
	description.Target = newDescription.Target
	l.pools[Pool.Name].vols[description.Name] = &FakeLibvirtStorageVol{
		description: description,
		content:     oldVol.content,
	}
	return libvirt.StorageVol{
		Name: description.Name,
		Pool: Pool.Name,
	}, nil
}

func (l *FakeLibvirtConnection) StorageVolDownload(Vol libvirt.StorageVol, inStream io.Writer, Offset, Length uint64, Flags libvirt.StorageVolDownloadFlags) (err error) {
	vol, ok := l.pools[Vol.Pool].vols[Vol.Name]
	if !ok {
		return libvirt.Error{Code: uint32(libvirt.ErrNoStorageVol)}
	}

	_, err = inStream.Write(vol.content)
	if err != nil {
		return errors.New("error writing upload data")
	}

	return nil
}

func (l *FakeLibvirtConnection) StorageVolGetInfo(Vol libvirt.StorageVol) (rType int8, rCapacity, rAllocation uint64, err error) {
	_, ok := l.pools[Vol.Pool].vols[Vol.Name]
	if !ok {
		return 0, 0, 0, libvirt.Error{Code: uint32(libvirt.ErrNoStorageVol)}
	}

	return 0, 42, 23, nil
}

func (l *FakeLibvirtConnection) ConnectListNetworks(Maxnames int32) (rNames []string, err error) {
	return []string{networkName}, nil
}

func (l *FakeLibvirtConnection) NetworkLookupByName(Name string) (rNet libvirt.Network, err error) {
	_, ok := l.networks[Name]
	if !ok {
		return libvirt.Network{}, libvirt.Error{Code: uint32(libvirt.ErrNoNetwork)}
	}

	return libvirt.Network{Name: Name}, nil
}

func (l *FakeLibvirtConnection) NetworkGetXMLDesc(Net libvirt.Network, Flags uint32) (string, error) {
	n, ok := l.networks[Net.Name]
	if !ok {
		return "", libvirt.Error{Code: uint32(libvirt.ErrNoNetwork)}
	}

	xmldesc, err := xml.Marshal(n.description)
	if err != nil {
		return "", err
	}

	return string(xmldesc), nil
}

func (l *FakeLibvirtConnection) NetworkUpdate(Net libvirt.Network, command, section uint32, ParentIndex int32, XML string, Flags libvirt.NetworkUpdateFlags) (err error) {
	if section != uint32(libvirt.NetworkSectionIPDhcpHost) {
		return errors.New("unknown section")
	}

	var n *libvirtxml.Network
	for _, knownNet := range l.networks {
		if knownNet.description.Name == Net.Name {
			n = knownNet.description
		}
	}

	hosts := &n.IPs[0].DHCP.Hosts

	host := &libvirtxml.NetworkDHCPHost{}
	if err := host.Unmarshal(XML); err != nil {
		return fmt.Errorf("invalid network host XML: %w", err)
	}

	if command == uint32(libvirt.NetworkUpdateCommandAddLast) {
		*hosts = append(*hosts, *host)
	} else if command == uint32(libvirt.NetworkUpdateCommandDelete) {
		newHosts := []libvirtxml.NetworkDHCPHost{}
		for _, h := range *hosts {
			if h.MAC != host.MAC || h.IP != host.IP {
				newHosts = append(newHosts, h)
			}
		}
		if len(newHosts) == len(*hosts) {
			return errors.New("host for deletion not found")
		}
		if len(newHosts) < len(*hosts)-1 {
			return errors.New("host for deletion not unique")
		}
		*hosts = newHosts
	} else {
		return errors.New("unknown command")
	}

	return nil
}

func (l *FakeLibvirtConnection) DomainLookupByName(Name string) (rDom libvirt.Domain, err error) {
	_, ok := l.domains[Name]
	if !ok {
		return libvirt.Domain{}, libvirt.Error{Code: uint32(libvirt.ErrNoDomain)}
	}

	return libvirt.Domain{
		Name: Name,
	}, nil
}

func (l *FakeLibvirtConnection) DomainGetXMLDesc(Dom libvirt.Domain, Flags libvirt.DomainXMLFlags) (rXML string, err error) {
	domain, ok := l.domains[Dom.Name]
	if !ok {
		return "", libvirt.Error{Code: uint32(libvirt.ErrNoDomain)}
	}

	xml, err := domain.description.Marshal()
	if err != nil {
		panic(err)
	}
	return xml, nil
}

func (l *FakeLibvirtConnection) DomainDefineXML(XML string) (rDom libvirt.Domain, err error) {
	description := &libvirtxml.Domain{}
	if err := description.Unmarshal(XML); err != nil {
		return libvirt.Domain{}, fmt.Errorf("invalid domain XML: %w", err)
	}
	l.domains[description.Name] = &FakeLibvirtDomain{
		description: description,
		persistent:  true,
	}
	return libvirt.Domain{
		Name: description.Name,
	}, nil
}

func (l *FakeLibvirtConnection) DomainCreate(Dom libvirt.Domain) (err error) {
	domain, ok := l.domains[Dom.Name]
	if !ok {
		return libvirt.Error{Code: uint32(libvirt.ErrNoDomain)}
	}

	domain.active = true

	return nil
}

func (l *FakeLibvirtConnection) DomainIsActive(Dom libvirt.Domain) (rActive int32, err error) {
	domain, ok := l.domains[Dom.Name]
	if !ok {
		return 0, libvirt.Error{Code: uint32(libvirt.ErrNoDomain)}
	}

	return boolToInt32(domain.active), nil
}

func (l *FakeLibvirtConnection) DomainIsPersistent(Dom libvirt.Domain) (rPersistent int32, err error) {
	domain, ok := l.domains[Dom.Name]
	if !ok {
		return 0, libvirt.Error{Code: uint32(libvirt.ErrNoDomain)}
	}

	return boolToInt32(domain.persistent), nil
}

func boolToInt32(b bool) int32 {
	if b {
		return 1
	}
	return 0
}

func (l *FakeLibvirtConnection) DomainShutdown(Dom libvirt.Domain) (err error) {
	domain, ok := l.domains[Dom.Name]
	if !ok {
		return libvirt.Error{Code: uint32(libvirt.ErrNoDomain)}
	}

	domain.active = false

	gcDomain(l.domains, Dom.Name, domain)

	return nil
}

func (l *FakeLibvirtConnection) DomainDestroy(Dom libvirt.Domain) (err error) {
	domain, ok := l.domains[Dom.Name]
	if !ok {
		return libvirt.Error{Code: uint32(libvirt.ErrNoDomain)}
	}

	domain.active = false

	gcDomain(l.domains, Dom.Name, domain)

	return nil
}

func (l *FakeLibvirtConnection) DomainUndefineFlags(Dom libvirt.Domain, Flags libvirt.DomainUndefineFlagsValues) (err error) {
	domain, ok := l.domains[Dom.Name]
	if !ok {
		return libvirt.Error{Code: uint32(libvirt.ErrNoDomain)}
	}

	domain.persistent = false

	gcDomain(l.domains, Dom.Name, domain)

	return nil
}

func (l *FakeLibvirtConnection) ConnectGetDomainCapabilities(Emulatorbin libvirt.OptString, Arch libvirt.OptString, Machine libvirt.OptString, Virttype libvirt.OptString, Flags uint32) (string, error) {
	return "", nil
}

func gcDomain(domains map[string]*FakeLibvirtDomain, name string, domain *FakeLibvirtDomain) {
	if !domain.persistent && !domain.active {
		delete(domains, name)
	}
}

func (l *FakeLibvirtConnection) DomainListAllSnapshots(Dom libvirt.Domain, NeedResults int32, Flags uint32) (rSnapshots []libvirt.DomainSnapshot, rRet int32, err error) {
	_, ok := l.domains[Dom.Name]
	if !ok {
		return []libvirt.DomainSnapshot{}, 0, libvirt.Error{Code: uint32(libvirt.ErrNoDomain)}
	}

	return []libvirt.DomainSnapshot{}, 0, nil
}

func (l *FakeLibvirtConnection) DomainSnapshotDelete(Snap libvirt.DomainSnapshot, Flags libvirt.DomainSnapshotDeleteFlags) (err error) {
	return nil
}

func (l *FakeLibvirtConnection) addEmptyRawVol(pool string, name string) *FakeLibvirtStorageVol {
	empty := &FakeLibvirtStorageVol{
		description: &libvirtxml.StorageVolume{
			Name: name,
		},
	}

	l.pools[pool].vols[empty.description.Name] = empty

	return empty
}

func (l *FakeLibvirtConnection) addFakeImage(pool string, name string) *FakeLibvirtStorageVol {
	empty := &FakeLibvirtStorageVol{
		description: &libvirtxml.StorageVolume{
			// Hash for empty volume
			Name: virter.LayerVolumePrefix + ExampleLayerDigest,
		},
		content: []byte(ExampleLayerContent),
	}

	l.pools[pool].vols[empty.description.Name] = empty

	tag := &FakeLibvirtStorageVol{
		description: &libvirtxml.StorageVolume{
			Name: virter.TagVolumePrefix + name,
			BackingStore: &libvirtxml.StorageVolumeBackingStore{
				Path: "./" + empty.description.Name,
			},
		},
	}

	l.pools[pool].vols[tag.description.Name] = tag

	return tag
}
func (l *FakeLibvirtConnection) addFakeStoragePool(name string) *FakeLibvirtStoragePool {
	pool := &FakeLibvirtStoragePool{
		description: &libvirtxml.StoragePool{
			Name: name,
		},
		vols: make(map[string]*FakeLibvirtStorageVol),
	}

	l.pools[pool.description.Name] = pool
	return pool
}

func fakeLibvirtNetwork() *FakeLibvirtNetwork {
	return &FakeLibvirtNetwork{
		description: &libvirtxml.Network{
			XMLName: xml.Name{Local: "network"},
			Name:    networkName,
			Domain:  &libvirtxml.NetworkDomain{Name: "fake-domain.com"},
			IPs: []libvirtxml.NetworkIP{
				libvirtxml.NetworkIP{
					Address: networkAddress,
					Netmask: networkNetmask,
					DHCP:    &libvirtxml.NetworkDHCP{},
				},
			},
		},
	}
}

func fakeNetworkAddHost(network *FakeLibvirtNetwork, mac, ip string) {
	hosts := &network.description.IPs[0].DHCP.Hosts
	host := libvirtxml.NetworkDHCPHost{
		MAC: mac,
		IP:  ip,
	}
	*hosts = append(*hosts, host)
}

const fakeVMMeta = `
<meta xmlns="https://github.com/LINBIT/virter">
	<hostkey>ssh-rsa abcdef123456789</hostkey>
</meta>
`

func newFakeLibvirtDomain(name string, mac string) *FakeLibvirtDomain {
	return &FakeLibvirtDomain{
		description: &libvirtxml.Domain{
			Metadata: &libvirtxml.DomainMetadata{XML: fakeVMMeta},
			Name:     name,
			Devices: &libvirtxml.DomainDeviceList{
				Interfaces: []libvirtxml.DomainInterface{
					libvirtxml.DomainInterface{
						Source: &libvirtxml.DomainInterfaceSource{
							Network: &libvirtxml.DomainInterfaceSourceNetwork{
								Network: networkName,
							},
						},
						MAC: &libvirtxml.DomainInterfaceMAC{
							Address: mac,
						},
					},
				},
			},
		},
	}
}

const (
	networkAddress = "192.168.122.1"
	networkNetmask = "255.255.255.0"
)
