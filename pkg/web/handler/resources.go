package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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

	"github.com/cockroachdb/errors"
	"github.com/emicklei/go-restful/v3"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/executor"
	"github.com/kubesphere/kubekey/v4/pkg/utils"
	"github.com/kubesphere/kubekey/v4/pkg/web/api"
	"github.com/kubesphere/kubekey/v4/pkg/web/query"
)

// ResourceHandler handles resource-related HTTP requests.
type ResourceHandler struct {
	rootPath string
	workdir  string
	client   ctrlclient.Client
}

// NewResourceHandler creates a new ResourceHandler instance.
func NewResourceHandler(rootPath string, workdir string, client ctrlclient.Client) *ResourceHandler {
	return &ResourceHandler{rootPath: rootPath, workdir: workdir, client: client}
}

// ConfigInfo serves the config file content as the HTTP response.
func (h ResourceHandler) ConfigInfo(request *restful.Request, response *restful.Response) {
	file, err := os.Open(filepath.Join(h.rootPath, api.SchemaConfigFile))
	if err != nil {
		if os.IsNotExist(err) {
			_ = response.WriteEntity(api.SUCCESS.SetResult("waiting for config to be created"))
		} else {
			_ = response.WriteError(http.StatusInternalServerError, err)
		}
		return
	}
	defer file.Close()

	_, _ = io.Copy(response.ResponseWriter, file)
}

// PostConfig updates the config file and triggers precheck playbooks if needed.
func (h ResourceHandler) PostConfig(request *restful.Request, response *restful.Response) {
	var (
		oldConfig map[string]map[string]any
		newConfig map[string]map[string]any
	)
	bodyBytes, err := io.ReadAll(request.Request.Body)
	if err != nil {
		_ = response.WriteError(http.StatusInternalServerError, err)
		return
	}
	// Read new config from request body.
	if err := json.Unmarshal(bodyBytes, &newConfig); err != nil {
		_ = response.WriteError(http.StatusInternalServerError, err)
		return
	}

	// Open config file for reading and writing.
	if oldConfigFile, err := os.ReadFile(filepath.Join(h.rootPath, api.SchemaConfigFile)); err == nil {
		// Decode old config if present.
		if err := json.Unmarshal(oldConfigFile, &oldConfig); err != nil && !errors.Is(err, io.EOF) {
			_ = response.WriteError(http.StatusInternalServerError, err)
			return
		}
	}

	namespace := query.DefaultString(request.QueryParameter("cluster"), "default")
	inventory := query.DefaultString(request.QueryParameter("inventory"), "default")
	playbooks := make(map[string]*kkcorev1.Playbook)
	wg := wait.Group{}

	// Iterate over new config and trigger precheck playbooks if config changed.
	for fileName, newVal := range newConfig {
		// if playbook has created should skip it.
		playbookList := &kkcorev1.PlaybookList{}
		if err := h.client.List(request.Request.Context(), playbookList, ctrlclient.InNamespace(namespace),
			ctrlclient.MatchingLabels{"install." + api.SchemaLabelSubfix: fileName}); err != nil {
			_ = response.WriteError(http.StatusInternalServerError, err)
			return
		}
		if len(playbookList.Items) > 0 {
			continue
		}
		schemaInfo, err := os.ReadFile(filepath.Join(h.rootPath, fileName))
		if err != nil {
			_ = response.WriteError(http.StatusInternalServerError, err)
			return
		}
		var schemaFile api.SchemaFile
		if err := json.Unmarshal(schemaInfo, &schemaFile); err != nil {
			_ = response.WriteError(http.StatusInternalServerError, err)
			return
		}
		// If a precheck playbook is defined, create and execute it.
		if pbpath := schemaFile.PlaybookPath["precheck."+api.SchemaLabelSubfix]; pbpath != "" {
			playbook := &kkcorev1.Playbook{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "precheck-" + strings.TrimSuffix(fileName, filepath.Ext(fileName)) + "-",
					Namespace:    namespace,
				},
				Spec: kkcorev1.PlaybookSpec{
					Config: kkcorev1.Config{
						Spec: runtime.RawExtension{Object: &unstructured.Unstructured{Object: newVal}},
					},
					InventoryRef: &corev1.ObjectReference{
						Kind:      "Inventory",
						Namespace: namespace,
						Name:      inventory,
					},
					Playbook: pbpath,
				},
				Status: kkcorev1.PlaybookStatus{
					Phase: kkcorev1.PlaybookPhasePending,
				},
			}
			// Set the workdir in the playbook's spec config
			if err := unstructured.SetNestedField(playbook.Spec.Config.Value(), h.workdir, _const.Workdir); err != nil {
				api.HandleError(response, request, err)
				return
			}
			if err := h.client.Create(context.TODO(), playbook); err != nil {
				api.HandleError(response, request, errors.Wrap(err, "failed to create precheck playbook"))
				return
			}
			playbooks[fileName] = playbook
			wg.Start(func() {
				// Execute the playbook asynchronously.
				if err := executor.PlaybookManager.Executor(playbook, h.client, "false"); err != nil {
					klog.ErrorS(err, "failed to executor precheck playbook", "schema", fileName)
				}
			})
		}
	}
	wg.Wait()

	// Collect precheck results.
	preCheckResult := make(map[string]string)
	for fileName, playbook := range playbooks {
		if playbook.Status.FailureMessage != "" {
			preCheckResult[fileName] = playbook.Status.FailureMessage
			delete(newConfig, fileName)
		}
	}

	// Write new config to file.
	if err := os.WriteFile(filepath.Join(h.rootPath, api.SchemaConfigFile), bodyBytes, 0644); err != nil {
		_ = response.WriteError(http.StatusInternalServerError, err)
		return
	}

	// Respond with precheck results if any failures, otherwise success.
	if len(preCheckResult) > 0 {
		_ = response.WriteHeaderAndEntity(http.StatusUnprocessableEntity, api.Result{Message: api.ResultFailed, Result: preCheckResult})
	} else {
		_ = response.WriteEntity(api.SUCCESS.SetResult(newConfig))
	}
}

// ListIP lists all IPs in the given CIDR, checks their online and SSH status, and returns the result.
func (h ResourceHandler) ListIP(request *restful.Request, response *restful.Response) {
	queryParam := query.ParseQueryParameter(request)
	cidr := request.QueryParameter("cidr")
	sshPort := query.DefaultString(request.QueryParameter("sshPort"), "22")

	var existsInventoryList kkcorev1.InventoryList
	err := h.client.List(request.Request.Context(), &existsInventoryList)

	if err != nil && ctrlclient.IgnoreNotFound(err) != nil {
		klog.Errorf("failed to list inventory objects: %v", err)
		_ = response.WriteError(http.StatusInternalServerError, err)
		return
	}

	var addedPorts = make(map[string]struct{})
	for _, item := range existsInventoryList.Items {
		for _, iData := range item.Spec.Hosts {
			var iHost api.InventoryConnect
			err = json.Unmarshal(iData.Raw, &iHost)
			if err != nil {
				continue
			}
			if iHost.Connector.Port == "" {
				iHost.Connector.Port = "22"
			}
			addedPorts[iHost.Connector.Host+":"+iHost.Connector.Port] = struct{}{}
		}
	}

	ips := utils.ParseIP(cidr)
	ipTable := make([]api.IPTable, 0, len(ips))
	maxConcurrency := 20
	mu := sync.Mutex{}
	jobChannel := make(chan string, 20)
	wg := wait.Group{}
	// Start worker goroutines for concurrent IP checking.
	for range maxConcurrency {
		wg.Start(func() {
			for ip := range jobChannel {
				var added = false
				if _, ok := addedPorts[ip+":"+sshPort]; ok {
					added = true
				}
				if utils.IsLocalhostIP(ip) {
					mu.Lock()
					ipTable = append(ipTable, api.IPTable{
						IP:            ip,
						SSHPort:       sshPort,
						Localhost:     true,
						SSHReachable:  true,
						SSHAuthorized: true,
						Added:         added,
					})
					mu.Unlock()
					continue
				}

				// Check if the host is online using the ICMP protocol (ping).
				// Requires root privileges or CAP_NET_RAW capability.
				if !isIPOnline(ip) {
					continue
				}
				reachable, authorized := isSSHAuthorized(ip, sshPort)

				mu.Lock()
				ipTable = append(ipTable, api.IPTable{
					IP:            ip,
					SSHPort:       sshPort,
					SSHReachable:  reachable,
					SSHAuthorized: authorized,
					Added:         added,
				})
				mu.Unlock()
			}
		})
	}

	// Send IPs to job channel for processing.
	for _, ip := range ips {
		jobChannel <- ip
	}

	close(jobChannel)
	wg.Wait()

	// less is a comparison function for sorting IPTable items by a given field.
	less := func(left, right api.IPTable, sortBy string) bool {
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
	// filter is a function for filtering IPTable items by a given field and value.
	filter := func(o api.IPTable, f query.Filter) bool {
		val := query.GetFieldByJSONTag(reflect.ValueOf(o), f.Field)
		switch val.Kind() {
		case reflect.String:
			return strings.Contains(val.String(), f.Value)
		default:
			return true
		}
	}

	// Use the DefaultList function to apply filtering, sorting, and pagination.
	results := query.DefaultList(ipTable, queryParam, less, filter)
	_ = response.WriteEntity(results)
}

// SchemaInfo serves static schema files from the rootPath directory.
// It strips the /resources/schema/ prefix and serves files using http.FileServer.
func (h ResourceHandler) SchemaInfo(request *restful.Request, response *restful.Response) {
	http.StripPrefix("/resources/schema/", http.FileServer(http.Dir(h.rootPath))).ServeHTTP(response.ResponseWriter, request.Request)
}

// ListSchema lists all schema JSON files in the rootPath directory as a table.
// It supports filtering, sorting, and pagination via query parameters.
func (h ResourceHandler) ListSchema(request *restful.Request, response *restful.Response) {
	queryParam := query.ParseQueryParameter(request)
	// Read all entries in the rootPath directory.
	entries, err := os.ReadDir(h.rootPath)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}
	schemaTable := make([]api.SchemaTable, 0)
	for _, entry := range entries {
		// Skip directories, non-JSON files, and special schema files.
		if entry.IsDir() ||
			!strings.HasSuffix(entry.Name(), ".json") ||
			entry.Name() == api.SchemaProductFile || entry.Name() == api.SchemaConfigFile {
			continue
		}
		// Read the JSON file.
		data, err := os.ReadFile(filepath.Join(h.rootPath, entry.Name()))
		if err != nil {
			api.HandleError(response, request, errors.Wrapf(err, "failed to read file for schema %q", entry.Name()))
			return
		}
		var schemaFile api.SchemaFile
		// Unmarshal the JSON data into a SchemaTable struct.
		if err := json.Unmarshal(data, &schemaFile); err != nil {
			api.HandleError(response, request, errors.Wrapf(err, "failed to unmarshal file for schema %q", entry.Name()))
			return
		}
		// Get all playbooks in the given namespace.
		playbookList := &kkcorev1.PlaybookList{}
		if err := h.client.List(request.Request.Context(), playbookList, ctrlclient.InNamespace(request.PathParameter("cluster"))); err != nil {
			api.HandleError(response, request, err)
			return
		}

		schema := api.SchemaFile2Table(schemaFile, filepath.Join(h.rootPath, api.SchemaConfigFile), entry.Name())
		// For each playbook, if it matches a label in schema.Playbook and the label value equals schema.Name, add its info.
		for _, playbook := range playbookList.Items {
			for label, schemaName := range playbook.Labels {
				// Only process playbooks whose label is defined in schema.Playbook and value matches schema.Name.
				if _, ok := schema.Playbook[label]; ok && schemaName == schema.Name {
					// If a playbook for this label already exists, return an error.
					if schema.Playbook[label].Name != "" {
						api.HandleError(response, request, errors.Errorf("schema %q has multiple playbooks of label %q", entry.Name(), label))
						return
					}
					var result any
					// If the playbook has a result, unmarshal it.
					if len(playbook.Status.Result.Raw) != 0 {
						if err := json.Unmarshal(playbook.Status.Result.Raw, &result); err != nil {
							api.HandleError(response, request, errors.Errorf("failed to unmarshal result from playbook of schema %q", schema.Name))
							return
						}
					}
					// Fill in playbook info for this label.
					schema.Playbook[label] = api.SchemaTablePlaybook{
						Path:      schema.Playbook[label].Path,
						Name:      playbook.Name,
						Namespace: playbook.Namespace,
						Phase:     string(playbook.Status.Phase),
						Result:    result,
					}
				}
			}
		}
		// Add the processed schema to the schemaTable slice.
		schemaTable = append(schemaTable, schema)
	}
	// less is a comparison function for sorting SchemaTable items by a given field.
	less := func(left, right api.SchemaTable, sortBy string) bool {
		leftVal := query.GetFieldByJSONTag(reflect.ValueOf(left), sortBy)
		rightVal := query.GetFieldByJSONTag(reflect.ValueOf(right), sortBy)
		switch leftVal.Kind() {
		case reflect.String:
			return leftVal.String() > rightVal.String()
		case reflect.Int, reflect.Int64:
			return leftVal.Int() > rightVal.Int()
		default:
			// If the field is not a string or int, sort by Priority as a fallback.
			return left.Priority > right.Priority
		}
	}
	// filter is a function for filtering SchemaTable items by a given field and value.
	filter := func(o api.SchemaTable, f query.Filter) bool {
		val := query.GetFieldByJSONTag(reflect.ValueOf(o), f.Field)
		switch val.Kind() {
		case reflect.String:
			return strings.Contains(val.String(), f.Value)
		case reflect.Int:
			v, err := strconv.Atoi(f.Value)
			if err != nil {
				return false
			}
			return v == int(val.Int())
		default:
			return true
		}
	}

	// Use the DefaultList function to apply filtering, sorting, and pagination.
	// The results variable contains the filtered, sorted, and paginated schemaTable.
	results := query.DefaultList(schemaTable, queryParam, less, filter)
	_ = response.WriteEntity(results)
}

// ===========================================================================
// =============================   isIPOnline   ==============================
// ===========================================================================

// isIPOnline checks if the given IP address is online by sending an ICMP Echo Request (ping).
// It returns true if a reply is received, false otherwise.
// The timeout for the ICMP connection and reply is set to 1 second.
func isIPOnline(ipStr string) bool {
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

// isValidICMPReply checks if the received ICMP reply is valid and matches the expected parameters.
// n: number of bytes read
// reply: the reply buffer
// src: source address of the reply
// expectedIP: the IP we expect the reply from
// protocol: ICMP protocol number
// pid: process ID used in the Echo request
// seq: sequence number used in the Echo request
// replyFilter: function to filter valid ICMP types
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

// ===========================================================================
// =============================   isSSHAuthorized   =========================
// ===========================================================================

// isSSHAuthorized checks if SSH authorization to the given IP is possible using the local private key.
// It returns two booleans: the first indicates if the SSH port (22) is reachable, and the second indicates if SSH authorization using the local private key is successful.
// The function attempts to find the user's private key, read and parse it, and then connect via SSH.
func isSSHAuthorized(ipStr, sshPort string) (bool, bool) {
	// First check if port 22 is reachable on the target IP address.
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", ipStr, sshPort), time.Second)
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

func checkSSHConnect(ipStr, sshPort, sshUser, sshPwd, sshPrivateKeyContent string) (bool, bool) {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(ipStr, sshPort), time.Second)
	if err != nil {
		klog.V(6).Infof("port %s not reachable on ip %q, error %v", sshPort, ipStr, err)
		return false, false
	}
	defer conn.Close()

	var authMethods []ssh.AuthMethod

	if sshPwd != "" {
		authMethods = append(authMethods, ssh.Password(sshPwd))
		klog.V(6).Infof("Added password authentication for user %s", sshUser)
	}

	if sshPrivateKeyContent != "" {
		signer, err := ssh.ParsePrivateKey([]byte(sshPrivateKeyContent))
		if err != nil {
			klog.V(6).Infof("Failed to parse provided private key: %v", err)
		} else {
			authMethods = append(authMethods, ssh.PublicKeys(signer))
			klog.V(6).Infof("Added public key authentication from provided content for user %s", sshUser)
		}
	} else {
		klog.V(6).Infof("No private key content provided, checking for default private keys")
		foundKeys := findSSHPrivateKeys()
		var signers = make([]ssh.Signer, 0)
		for _, keyPath := range foundKeys {
			keyBytes, err := os.ReadFile(keyPath)
			if err != nil {
				klog.V(6).Infof("Failed to read private key file %s: %v", keyPath, err)
				continue
			}

			signer, err := ssh.ParsePrivateKey(keyBytes)
			if err != nil {
				klog.V(6).Infof("Failed to parse private key from %s: %v", keyPath, err)
				continue
			}
			signers = append(signers, signer)
		}
		if len(signers) > 0 {
			authMethods = append(authMethods, ssh.PublicKeys(signers...))
		}
	}

	klog.V(6).Infof("Using %d authentication methods for SSH connection", len(authMethods))

	config := &ssh.ClientConfig{
		User:            sshUser,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	sshClient, err := ssh.Dial("tcp", net.JoinHostPort(ipStr, sshPort), config)
	if err != nil {
		klog.V(6).Infof("SSH connection failed: %v", err)
		return true, false
	}
	defer sshClient.Close()

	session, err := sshClient.NewSession()
	if err != nil {
		klog.V(6).Infof("SSH session creation failed: %v", err)
		return true, false
	}
	defer session.Close()

	err = session.Run("echo 'SSH connection test'")
	if err != nil {
		klog.V(6).Infof("SSH command execution failed: %v", err)
		return true, false
	}

	klog.V(6).Infof("SSH connection successful for user %s", sshUser)
	return true, true
}

func findSSHPrivateKeys() []string {
	var keyFiles = make([]string, 0)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		klog.V(6).Infof("Failed to get user home directory: %v", err)
		homeDir = "/root"
	}

	sshDir := filepath.Join(homeDir, ".ssh")
	var sshDirInfo os.FileInfo

	if sshDirInfo, err = os.Stat(sshDir); os.IsNotExist(err) {
		klog.V(6).Infof("SSH directory %s does not exist", sshDir)
		return keyFiles
	}

	if !sshDirInfo.IsDir() {
		return keyFiles
	}

	dirs, _ := os.ReadDir(sshDir)
	for _, dir := range dirs {
		if dir.IsDir() {
			continue
		}
		keyFiles = append(keyFiles, filepath.Join(sshDir, dir.Name()))
	}

	return keyFiles
}

// PreCheckHost check input ssh information.
func (h ResourceHandler) PreCheckHost(request *restful.Request, response *restful.Response) {
	var hosts []api.IPHostCheckData
	if err := request.ReadEntity(&hosts); err != nil {
		api.HandleError(response, request, err)
		return
	}
	var wg = sync.WaitGroup{}
	var result = make([]api.IPHostCheckResult, len(hosts))
	wg.Add(len(hosts))
	for i, host := range hosts {
		go func(idx int, currentHost api.IPHostCheckData) {
			defer wg.Done()
			var status string
			if utils.IsLocalhostIP(currentHost.IP) {
				status = _const.SSHVerifyStatusSuccess
			}
			if !isIPOnline(currentHost.IP) {
				status = _const.SSHVerifyStatusOffline
			}
			if currentHost.SSHUser == "" {
				status = _const.SSHVerifyStatusSSHIncomplete
			}
			if status == "" {
				reachable, authorized := checkSSHConnect(currentHost.IP, currentHost.SSHPort,
					currentHost.SSHUser, currentHost.SSHPwd, currentHost.SSHPrivateKeyContent)
				klog.V(6).Infof("check ssh connect for %s:%s with result:%v,%v",
					currentHost.IP, currentHost.SSHPort, reachable, authorized)
				switch {
				case authorized && reachable:
					status = _const.SSHVerifyStatusSuccess
				case !authorized && reachable:
					status = _const.SSHVerifyStatusFailed
				case !reachable && !authorized:
					status = _const.SSHVerifyStatusUnreachable
				default:
					klog.Warningf("check ssh connect show authorized but unreachable! ip:%s,port=%s",
						currentHost.IP, currentHost.SSHPort)
					status = _const.SSHVerifyStatusFailed
				}
			}
			// if ssh_failed, it means current host can access target host but unauthorized
			// in this case,if user did not input pwd or key,it means ssh information incomplete
			if status == _const.SSHVerifyStatusFailed && currentHost.SSHPwd == "" && currentHost.SSHPrivateKeyContent == "" {
				status = _const.SSHVerifyStatusSSHIncomplete
			}
			result[idx] = api.IPHostCheckResult{
				IP:      currentHost.IP,
				SSHPort: currentHost.SSHPort,
				Status:  status,
			}
		}(i, host)
	}
	wg.Wait()
	_ = response.WriteEntity(result)
}

func (h ResourceHandler) ResourceSummary(request *restful.Request, response *restful.Response) {

}
