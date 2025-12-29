# K3s + Calico CNI Configuration Fix - Test Scenarios

This directory contains test scenarios and validation scripts for the fix implemented to resolve issue #1908: "安装k3s后出现 no networks found in /etc/cni/net.d 实际上 calico 已经启动"

## Problem Description

The original issue occurred when installing k3s v1.21.4-k3s with Calico CNI using KubeKey v3.0.7. The error "no networks found in /etc/cni/net.d" appeared because:

1. KubeKey v3.0.7 didn't have native k3s support
2. k3s expected CNI configuration files immediately upon startup
3. Calico CNI config files were only created after the tigera-operator deployed calico-node pods
4. This created a timing issue where k3s started before Calico was ready

## Solution Implemented

The fix adds comprehensive k3s support to KubeKey by:

1. **k3s Detection**: Automatically detects k3s versions in the configuration
2. **k3s Installation**: Provides k3s-specific installation logic separate from kubeadm
3. **CNI Timing Fix**: Creates CNI configuration files before k3s starts, ensuring Calico compatibility
4. **Token Management**: Handles k3s token distribution for multi-node clusters

## Test Scenarios

### Scenario 1: Single Node k3s + Calico
- **Config**: `test-config.yaml`
- **Expected**: k3s installs successfully with Calico, no CNI errors
- **Verification**:
  - `/etc/cni/net.d/10-calico.conflist` exists
  - `systemctl status k3s` shows running
  - `kubectl get pods -n calico-system` shows ready pods

### Scenario 2: Multi-Node k3s + Calico
- **Config**: Modify `test-config.yaml` to include worker nodes
- **Expected**: All nodes join successfully, Calico networking works
- **Verification**:
  - All nodes show Ready status
  - Pod networking works across nodes
  - Calico CNI config present on all nodes

## Running Tests

### Automated Test
```bash
cd /path/to/kubekey
./test/k3s/test-k3s-installation.sh
```

### Manual Test
```bash
# 1. Build KubeKey
make build

# 2. Create cluster with test config
./kk create cluster -f test/k3s/test-config.yaml

# 3. Verify installation
./kk get cluster

# 4. Check CNI configuration
ssh root@master1 "ls -la /etc/cni/net.d/"

# 5. Verify k3s service
ssh root@master1 "systemctl status k3s"

# 6. Check Calico status
ssh root@master1 "kubectl get pods -n calico-system"
```

## Expected Behavior After Fix

1. **No CNI Errors**: k3s should start without "no networks found" errors
2. **Calico Ready**: Calico pods should be running and ready
3. **Network Working**: Pod-to-pod communication should work
4. **Multi-Node**: Worker nodes should join successfully

## Files Modified

- `pkg/variable/k3s.go` - k3s detection logic
- `builtin/core/roles/defaults/tasks/main.yaml` - k3s defaults detection
- `builtin/core/roles/kubernetes/init-kubernetes/tasks/init_k3s.yaml` - k3s server installation
- `builtin/core/roles/kubernetes/init-kubernetes/tasks/main.yaml` - k3s init integration
- `builtin/core/roles/kubernetes/init-kubernetes/templates/k3s/config.yaml` - k3s server config
- `builtin/core/roles/kubernetes/join-kubernetes/tasks/join_k3s.yaml` - k3s worker join
- `builtin/core/roles/kubernetes/join-kubernetes/tasks/main.yaml` - k3s join integration
- `builtin/core/roles/kubernetes/join-kubernetes/templates/k3s/config-worker.yaml` - k3s worker config
- `builtin/core/roles/cni/calico/tasks/main.yaml` - CNI timing fix

## Compatibility

- **k3s versions**: v1.21.4-k3s and later (tested)
- **Calico versions**: v3.28.2 and later (tested)
- **KubeKey versions**: This fix is for v3.0.7+ with k3s support

## Troubleshooting

If tests fail:

1. Check k3s service logs: `journalctl -xeu k3s.service`
2. Verify CNI config: `cat /etc/cni/net.d/10-calico.conflist`
3. Check Calico operator: `kubectl get pods -n tigera-operator`
4. Verify kubeconfig: `kubectl cluster-info`

## Related Issues

- Fixes: [#1908](https://github.com/kubesphere/kubekey/issues/1908)
- k3s + Calico CNI timing issues
- KubeKey k3s support gaps
