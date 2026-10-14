package netutil

import (
	"fmt"
	"golang-share-files/internal/apperror"
	"net"
	"strings"
)

var privateIPBlocks = parseCIDRs([]string{
	"127.0.0.0/8",
	"10.0.0.0/8",
	"172.16.0.0/12",
	"192.168.0.0/16",
	"169.254.0.0/16",
	"::1/128",
	"fe80::/10",
	"fc00::/7",
})

func parseCIDRs(cidrs []string) []*net.IPNet {
	var blocks []*net.IPNet
	for _, cidr := range cidrs {
		_, block, err := net.ParseCIDR(cidr)
		custErr := &apperror.Error{
			Kind: apperror.KindParse,
			Msg:  "Error Parsing",
			Err:  err,
		}
		if err != nil {
			panic(fmt.Errorf("Error on %q with detail : %w", cidr, custErr))
		}
		blocks = append(blocks, block)
	}
	return blocks
}

func isPrivateIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}

	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

func isVirtualName(name string) bool {
	prefixes := []string{
		"docker",  // docker0
		"br-",     // br-abc123 (Docker Compose bridge)
		"veth",    // veth1a2b3c (virtual ethernet pair)
		"virbr",   // virbr0 (libvirt bridge)
		"vmbr",    // vmbr0 (Proxmox bridge)
		"tun",     // tun0 (OpenVPN)
		"tap",     // tap0
		"wg",      // wg0 (WireGuard)
		"utun",    // utun0 (macOS VPN)
		"vmnet",   // vmnet1 (VMware)
		"vboxnet", // vboxnet0 (VirtualBox)
		"ppp",     // ppp0 (PPP dial-up)
		"dummy",   // dummy0
	}

	substrings := []string{
		"vethernet",  // vEthernet (WSL), vEthernet (Default Switch)
		"wsl",        // interface WSL
		"docker",     // kadang muncul di tengah nama
		"virtual",    // "Virtual Adapter", "Hyper-V Virtual Ethernet"
		"vpn",        // "VPN Connection", "Cisco AnyConnect VPN"
		"hyper-v",    // Hyper-V virtual switch
		"vmware",     // VMware Network Adapter
		"virtualbox", // VirtualBox Host-Only
	}

	lower := strings.ToLower(name)

	for _, p := range prefixes {
		if strings.HasPrefix(lower, p) {
			return true
		}
	}

	for _, s := range substrings {
		if strings.Contains(lower, s) {
			return true
		}
	}

	return false
}

func pickFrom(filtered []IPCandidate) error {
	if len(filtered) == 0 {
		return &apperror.Error{
			Kind: apperror.KindNoPrivateIP,
			Msg:  "tidak ada IP privat yang cocok",
		}
	}
	if len(filtered) > 1 {
		return &apperror.Error{
			Kind: apperror.KindAmbiguousIface,
			Msg:  "lebih dari satu interface layak; tentukan dengan --iface",
		}
	}
	return nil
}

func PickByName(cands []IPCandidate, name string) (IPCandidate, error) {
	var fM *IPCandidate
	var bM *IPCandidate

	for _, cand := range cands {
		if cand.Name != name {
			continue
		}

		if fM == nil {
			c := cand
			fM = &c
		}

		if cand.IsPrivate && !cand.IsVirtual {
			c := cand
			bM = &c
			break
		}
	}

	if bM != nil {
		return *bM, nil
	}

	if fM != nil {
		return *fM, nil
	}

	return IPCandidate{}, &apperror.Error{
		Kind: apperror.KindIfaceNotFound,
		Msg:  "interface tidak ditemukan: " + name,
	}
}

func BestCandidate(cands []IPCandidate) (IPCandidate, error) {
	var strict []IPCandidate
	for _, c := range cands {
		if c.IsPrivate && !c.IsVirtual {
			strict = append(strict, c)
		}
	}

	err := pickFrom(strict)
	if err == nil {
		return strict[0], nil
	}
	if !apperror.IsKind(err, apperror.KindNoPrivateIP) {
		return IPCandidate{}, err
	}

	var loose []IPCandidate
	for _, c := range cands {
		if c.IsPrivate {
			loose = append(loose, c)
		}
	}

	err = pickFrom(loose)
	if err == nil {
		return loose[0], nil
	}

	return IPCandidate{}, err
}

func FindCandidates() ([]IPCandidate, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, &apperror.Error{
			Kind: apperror.KindNoNetwork,
			Msg:  "gagal membaca interface jaringan",
			Err:  err,
		}
	}

	var out []IPCandidate

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, a := range addrs {
			ipnet, ok := a.(*net.IPNet)
			if !ok {
				continue
			}
			ip4 := ipnet.IP.To4()
			if ip4 == nil {
				continue
			}

			out = append(out, IPCandidate{
				Name:      iface.Name,
				IP:        ip4,
				IsPrivate: isPrivateIP(ip4),
				IsVirtual: isVirtualName(iface.Name),
			})
		}
	}

	if len(out) == 0 {
		return nil, &apperror.Error{
			Kind: apperror.KindNoPrivateIP,
			Msg:  "tidak ada interface dengan IPv4",
		}
	}

	return out, nil
}
