package web

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/netip"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/web/api"
	"github.com/kubesphere/kubekey/v4/pkg/web/query"
)

// NewSchemaService creates a new WebService that serves schema files from the specified root path.
// It sets up a route that handles GET requests to /resources/schema/{subpath} and serves files from the rootPath directory.
// The {subpath:*} parameter allows for matching any path under /resources/schema/.
func NewSchemaService(rootPath string, workdir string, client ctrlclient.Client) *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/resources").
		Produces(restful.MIME_JSON, "text/plain")

	h := newSchemaHandler(rootPath, workdir, client)

	ws.Route(ws.GET("/ip").To(h.listIP).
		Doc("list available ip from ip cidr").
		Metadata(restfulspec.KeyOpenAPITags, []string{_const.ResourceTag}).
		Param(ws.QueryParameter("cidr", "the cidr for ip").Required(true)).
		Param(ws.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d")).
		Param(ws.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(ws.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("false")).
		Param(ws.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=ip").Required(false).DefaultValue("ip")).
		Returns(http.StatusOK, _const.StatusOK, api.ListResult[api.IPTable]{}))

	ws.Route(ws.GET("/schema/{subpath:*}").To(h.schemaInfo).
		Metadata(restfulspec.KeyOpenAPITags, []string{_const.ResourceTag}))

	ws.Route(ws.GET("/schema").To(h.allSchema).
		Doc("list all schema as table").
		Metadata(restfulspec.KeyOpenAPITags, []string{_const.ResourceTag}).
		Param(ws.QueryParameter("schemaType", "the type of schema json").Required(false)).
		Param(ws.QueryParameter("playbookLabel", "the reference playbook of schema. eg: install.kubekey.kubesphere.io/schema,check.kubekey.kubesphere.io/schema"+
			"if empty will not return any reference playbook").Required(false)).
		Param(ws.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d")).
		Param(ws.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(ws.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("false")).
		Param(ws.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=priority")).
		Returns(http.StatusOK, _const.StatusOK, api.ListResult[api.SchemaTable]{}))

	return ws
}

// newSchemaHandler creates a new schemaHandler instance with the given rootPath, workdir, and client.
func newSchemaHandler(rootPath string, workdir string, client ctrlclient.Client) *schemaHandler {
	return &schemaHandler{rootPath: rootPath, workdir: workdir, client: client}
}

// schemaHandler handles schema-related HTTP requests.
type schemaHandler struct {
	rootPath string
	workdir  string
	client   ctrlclient.Client
}

func (h schemaHandler) listIP(request *restful.Request, response *restful.Response) {
	queryParam := query.ParseQueryParameter(request)
	cidr, ok := queryParam.Filters["cidr"]
	if !ok || len(cidr) == 0 {
		api.HandleBadRequest(response, request, errors.New("cidr parameter is required"))
		return
	}
	ips := _const.ParseIP(string(cidr))
	ipTable := make([]api.IPTable, 0, len(ips))
	maxConcurrency := 20
	mu := sync.Mutex{}
	jobChannel := make(chan string, 20)
	wg := sync.WaitGroup{}
	for range maxConcurrency {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ip := range jobChannel {
				if ifLocalhostIP(ip) {
					mu.Lock()
					ipTable = append(ipTable, api.IPTable{
						IP:            ip,
						Localhost:     true,
						SSHReachable:  true,
						SSHAuthorized: true,
					})
					mu.Unlock()
					continue
				}

				// Check if the host is online using the ICMP protocol (ping).
				// Requires root privileges or CAP_NET_RAW capability.
				if !ifIPOnline(ip) {
					continue
				}
				reachable, authorized := ifIPSSHAuthorized(ip)

				mu.Lock()
				ipTable = append(ipTable, api.IPTable{
					IP:            ip,
					SSHReachable:  reachable,
					SSHAuthorized: authorized,
				})
				mu.Unlock()
			}
		}()
	}

	for _, ip := range ips {
		jobChannel <- ip
	}

	close(jobChannel)
	wg.Wait()

	// less is a comparison function for sorting SchemaTable items by a given field.
	less := func(left, right api.IPTable, sortBy query.Field) bool {
		leftVal := query.GetFieldByJSONTag(reflect.ValueOf(left), sortBy)
		rightVal := query.GetFieldByJSONTag(reflect.ValueOf(right), sortBy)
		switch leftVal.Kind() {
		case reflect.String:
			if sortBy == "ip" {
				leftIP, err := netip.ParseAddr(leftVal.String())
				if err != nil {
					return true
				}
				rightIP, err := netip.ParseAddr(rightVal.String())
				if err != nil {
					return true
				}
				return leftIP.Compare(rightIP) > 0
			}
			return leftVal.String() > rightVal.String()
		case reflect.Int, reflect.Int64:
			return leftVal.Int() > rightVal.Int()
		default:
			return true
		}
	}
	// filter is a function for filtering SchemaTable items by a given field and value.
	filter := func(o api.IPTable, f query.Filter) bool {
		val := query.GetFieldByJSONTag(reflect.ValueOf(o), f.Field)
		switch val.Kind() {
		case reflect.String:
			return val.String() == string(f.Value)
		case reflect.Int:
			v, err := strconv.Atoi(string(f.Value))
			if err != nil {
				return false
			}
			return v == int(val.Int())
		default:
			return true
		}
	}

	// Use the DefaultList function to apply filtering, sorting, and pagination.
	results := query.DefaultList(ipTable, queryParam, less, filter)
	_ = response.WriteEntity(results)
}

// schemaInfo serves static schema files from the rootPath directory.
// It strips the /resources/schema/ prefix and serves files using http.FileServer.
func (h schemaHandler) schemaInfo(request *restful.Request, response *restful.Response) {
	http.StripPrefix("/resources/schema/", http.FileServer(http.Dir(h.rootPath))).ServeHTTP(response.ResponseWriter, request.Request)
}

// allSchema lists all schema JSON files in the rootPath directory as a table.
// It supports filtering, sorting, and pagination via query parameters.
func (h schemaHandler) allSchema(request *restful.Request, response *restful.Response) {
	queryParam := query.ParseQueryParameter(request)
	playbookLabel := string(queryParam.Filters["playbookLabel"])
	// Get all entries in the rootPath directory.
	entries, err := os.ReadDir(h.rootPath)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}
	schemaTable := make([]api.SchemaTable, 0)
	for _, entry := range entries {
		if entry.IsDir() || // Skip directories.
			!strings.HasSuffix(entry.Name(), ".json") || // Only process files with .json suffix.
			entry.Name() == "product.json" { // "product.json" is agreed file name
			continue
		}
		// Read the JSON file.
		data, err := os.ReadFile(filepath.Join(h.rootPath, entry.Name()))
		if err != nil {
			api.HandleBadRequest(response, request, err)
			return
		}
		schema := api.SchemaTable{Name: entry.Name()}
		// Unmarshal the JSON data into a SchemaTable struct.
		if err := json.Unmarshal(data, &schema); err != nil {
			api.HandleBadRequest(response, request, err)
			return
		}
		// get reference playbook
		if playbookLabel != "" {
			playbookList := &kkcorev1.PlaybookList{}
			if err := h.client.List(request.Request.Context(), playbookList, ctrlclient.MatchingLabels{
				playbookLabel: entry.Name(),
			}); err != nil {
				api.HandleBadRequest(response, request, err)
				return
			}
			schema.Playbook = make([]api.SchemaTablePlaybook, len(playbookList.Items))
			for i, playbook := range playbookList.Items {
				schema.Playbook[i] = api.SchemaTablePlaybook{
					Name:      playbook.Name,
					Namespace: playbook.Namespace,
					Phase:     string(playbook.Status.Phase),
				}
			}
		}
		schemaTable = append(schemaTable, schema)
	}
	// less is a comparison function for sorting SchemaTable items by a given field.
	less := func(left, right api.SchemaTable, sortBy query.Field) bool {
		leftVal := query.GetFieldByJSONTag(reflect.ValueOf(left), sortBy)
		rightVal := query.GetFieldByJSONTag(reflect.ValueOf(right), sortBy)
		switch leftVal.Kind() {
		case reflect.String:
			return leftVal.String() > rightVal.String()
		case reflect.Int, reflect.Int64:
			return leftVal.Int() > rightVal.Int()
		default:
			return left.Priority > right.Priority
		}
	}
	// filter is a function for filtering SchemaTable items by a given field and value.
	filter := func(o api.SchemaTable, f query.Filter) bool {
		val := query.GetFieldByJSONTag(reflect.ValueOf(o), f.Field)
		switch val.Kind() {
		case reflect.String:
			return val.String() == string(f.Value)
		case reflect.Int:
			v, err := strconv.Atoi(string(f.Value))
			if err != nil {
				return false
			}
			return v == int(val.Int())
		default:
			return true
		}
	}

	// Use the DefaultList function to apply filtering, sorting, and pagination.
	results := query.DefaultList(schemaTable, queryParam, less, filter)
	_ = response.WriteEntity(results)
}

// ifLocalhostIP checks if the given IP address string (ipStr) is bound to any local network interface.
// It returns true if the IP is found on any interface, false otherwise.
func ifLocalhostIP(ipStr string) bool {
	targetIP := net.ParseIP(ipStr)
	if targetIP == nil {
		return false
	}
	ifaces, err := net.Interfaces()
	if err != nil {
		return false
	}
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				if v.IP.Equal(targetIP) {
					return true
				}
			case *net.IPAddr:
				if v.IP.Equal(targetIP) {
					return true
				}
			}
		}
	}
	return false
}

// ifIPOnline checks if the given IP address is online by sending an ICMP Echo Request (ping).
// It returns true if a reply is received, false otherwise.
// The timeout for the ICMP connection and reply is set to 1 second.
func ifIPOnline(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	var (
		network     string
		icmpType    icmp.Type
		protocol    int
		listenAddr  string
		replyFilter func(icmp.Type) bool
		deadline    = time.Now().Add(1 * time.Second)
	)
	// Determine if the IP is IPv4 or IPv6 and set appropriate values.
	if ip.To4() != nil {
		network = "ip4:icmp"
		icmpType = ipv4.ICMPTypeEcho
		protocol = 1 // ICMP for IPv4
		listenAddr = "0.0.0.0"
		replyFilter = func(t icmp.Type) bool {
			return t == ipv4.ICMPTypeEchoReply || t == ipv4.ICMPTypeDestinationUnreachable
		}
	} else if ip.To16() != nil {
		network = "ip6:ipv6-icmp"
		icmpType = ipv6.ICMPTypeEchoRequest
		protocol = 58 // ICMPv6
		listenAddr = "::"
		replyFilter = func(t icmp.Type) bool {
			return t == ipv6.ICMPTypeEchoReply || t == ipv6.ICMPTypeDestinationUnreachable
		}
	} else {
		// Not a valid IP address.
		return false
	}

	// Listen for ICMP packets on the specified network.
	conn, err := icmp.ListenPacket(network, listenAddr)
	if err != nil {
		klog.V(6).Infof("connect to %q use icmp failed, error: %v", ip, err)
		return false
	}
	defer conn.Close()

	// Set a deadline for the entire operation (write + read).
	err = conn.SetDeadline(deadline)
	if err != nil {
		klog.V(6).Infof("set deadline for %q use icmp failed, error: %v", ip, err)
		return false
	}

	pid := os.Getpid() & 0xffff
	seq := int(time.Now().UnixNano() & 0xffff)
	// Construct the ICMP Echo Request message.
	msg := icmp.Message{
		Type: icmpType,
		Code: 0,
		Body: &icmp.Echo{
			ID:   pid,
			Seq:  seq,
			Data: []byte("PING"),
		},
	}
	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		klog.V(6).Infof("marshal msg to %q use icmp failed, error: %v", ip, err)
		return false
	}

	// Send the ICMP Echo Request to the target IP address.
	_, err = conn.WriteTo(msgBytes, &net.IPAddr{IP: ip})
	if err != nil {
		klog.V(6).Infof("write msg to %q use icmp failed, error: %v", ip, err)
		return false
	}

	reply := make([]byte, 1500)
	for time.Now().Before(deadline) {
		if time.Until(deadline) <= 0 {
			break
		}

		if err := conn.SetDeadline(time.Now().Add(100 * time.Millisecond)); err != nil {
			klog.V(6).Infof("set reply deadline for %q use icmp failed, error: %v", ip, err)
			continue
		}

		n, src, err := conn.ReadFrom(reply)
		if err != nil {
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				continue
			}
			klog.V(6).Infof("read msg from %q use icmp timeout, error: %v", ip, err)
			return false
		}

		if isValidICMPReply(n, reply, src, ip, protocol, pid, seq, replyFilter) {
			return true
		}
	}

	return false
}

func isValidICMPReply(n int, reply []byte, src net.Addr, expectedIP net.IP, protocol int, pid, seq int, replyFilter func(icmp.Type) bool) bool {
	srcIP, ok := src.(*net.IPAddr)
	if !ok || !srcIP.IP.Equal(expectedIP) {
		klog.V(6).Infof("Ignore response from non-target IP: %s (expected: %s)", srcIP, expectedIP)
		return false
	}

	recvMsg, err := icmp.ParseMessage(protocol, reply[:n])
	if err != nil {
		klog.V(6).Infof("parse msg from %q failed, error: %v", expectedIP, err)
		return false
	}

	if !replyFilter(recvMsg.Type) {
		klog.V(6).Infof("Ignore unrelated ICMP type: %v", recvMsg.Type)
		return false
	}

	switch recvMsg.Type {
	case ipv4.ICMPTypeEchoReply, ipv6.ICMPTypeEchoReply:
		if echo, ok := recvMsg.Body.(*icmp.Echo); ok && echo.ID == pid && echo.Seq == seq {
			return true
		}
	case ipv4.ICMPTypeDestinationUnreachable, ipv6.ICMPTypeDestinationUnreachable:
		return false
	}

	return false
}

// ifIPSSHAuthorized checks if SSH authorization to the given IP is possible using the local private key.
// It returns two booleans: the first indicates if the SSH port (22) is reachable, and the second indicates if SSH authorization using the local private key is successful.
// The function attempts to find the user's private key, read and parse it, and then connect via SSH.
func ifIPSSHAuthorized(ipStr string) (bool, bool) {
	// First check if port 22 is reachable on the target IP address.
	conn, err := net.DialTimeout("tcp", ipStr+":22", time.Second)
	if err != nil {
		klog.V(6).Infof("port 22 not reachable on ip %q, error %v", ipStr, err)
		return false, false
	}
	defer conn.Close()

	// Set default SSH user and private key path.
	sshUser := "root"
	sshPrivateKey := "/root/.ssh/id_rsa"
	// Try to get the current user and set the SSH user and private key path accordingly.
	if currentUser, err := user.Current(); err == nil {
		sshUser = currentUser.Username
		sshPrivateKey = filepath.Join(currentUser.HomeDir, ".ssh/id_rsa")
	}

	// Check if the private key file exists.
	if _, err := os.Stat(sshPrivateKey); err != nil {
		// Port 22 is reachable, but private key is not found.
		klog.V(6).Infof("cannot found private key %q local in ip %q, error %v", sshPrivateKey, ipStr, err)
		return true, false
	}

	// Read the private key file.
	key, err := os.ReadFile(sshPrivateKey)
	if err != nil {
		// Port 22 is reachable, but private key cannot be read.
		klog.V(6).Infof("cannot read private key %q local in ip %q, error %v", sshPrivateKey, ipStr, err)
		return true, false
	}

	// Parse the private key.
	privateKey, err := ssh.ParsePrivateKey(key)
	if err != nil {
		// Port 22 is reachable, but private key cannot be parsed.
		klog.V(6).Infof("cannot parse private key %q local in ip %q, error %v", sshPrivateKey, ipStr, err)
		return true, false
	}

	// Prepare SSH authentication method.
	auth := []ssh.AuthMethod{ssh.PublicKeys(privateKey)}

	// Attempt to establish an SSH connection to the target IP.
	sshClient, err := ssh.Dial("tcp", ipStr+":22", &ssh.ClientConfig{
		User:            sshUser,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Second,
	})
	if err != nil {
		// Port 22 is reachable, but SSH authentication failed.
		klog.V(6).Infof("failed to connect ip %q by ssh, error %v", ipStr, err)
		return true, false
	}
	defer sshClient.Close()

	// Port 22 is reachable and SSH authentication succeeded.
	return true, true
}
