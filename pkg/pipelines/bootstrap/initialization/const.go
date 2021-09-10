package initialization

const (
	// command software
	sudo      = "sudo"
	curl      = "curl"
	openssl   = "openssl"
	ebtables  = "ebtables"
	socat     = "socat"
	ipset     = "ipset"
	conntrack = "conntrack"
	docker    = "docker"
	showmount = "showmount"
	rbd       = "rbd"
	glusterfs = "glusterfs"

	// extra command tools
	nfs  = "nfs"
	ceph = "ceph"

	UnknownVersion = "UnknownVersion"
)

// defines the base software to be checked.
var baseSoftware = []string{
	sudo,
	curl,
	openssl,
	ebtables,
	socat,
	ipset,
	conntrack,
	docker,
	showmount,
	rbd,
	glusterfs,
}
