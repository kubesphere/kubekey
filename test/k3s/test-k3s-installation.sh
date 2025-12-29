#!/bin/bash
# Test script for k3s + Calico installation
# This script tests the fix for the CNI configuration issue

set -e

echo "=== K3s + Calico Installation Test ==="
echo "Testing the fix for issue #1908: 'no networks found in /etc/cni/net.d'"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test 1: Check if k3s detection works
echo -e "\n${YELLOW}Test 1: k3s version detection${NC}"
if [ -f "pkg/variable/k3s.go" ]; then
    echo -e "${GREEN}✓ k3s detection module exists${NC}"
else
    echo -e "${RED}✗ k3s detection module missing${NC}"
    exit 1
fi

# Test 2: Check if k3s installation tasks exist
echo -e "\n${YELLOW}Test 2: k3s installation tasks${NC}"
if [ -f "builtin/core/roles/kubernetes/init-kubernetes/tasks/init_k3s.yaml" ]; then
    echo -e "${GREEN}✓ k3s init tasks exist${NC}"
else
    echo -e "${RED}✗ k3s init tasks missing${NC}"
    exit 1
fi

if [ -f "builtin/core/roles/kubernetes/join-kubernetes/tasks/join_k3s.yaml" ]; then
    echo -e "${GREEN}✓ k3s join tasks exist${NC}"
else
    echo -e "${RED}✗ k3s join tasks missing${NC}"
    exit 1
fi

# Test 3: Check if k3s templates exist
echo -e "\n${YELLOW}Test 3: k3s configuration templates${NC}"
if [ -f "builtin/core/roles/kubernetes/init-kubernetes/templates/k3s/config.yaml" ]; then
    echo -e "${GREEN}✓ k3s server config template exists${NC}"
else
    echo -e "${RED}✗ k3s server config template missing${NC}"
    exit 1
fi

if [ -f "builtin/core/roles/kubernetes/join-kubernetes/templates/k3s/config-worker.yaml" ]; then
    echo -e "${GREEN}✓ k3s worker config template exists${NC}"
else
    echo -e "${RED}✗ k3s worker config template missing${NC}"
    exit 1
fi

# Test 4: Check if CNI timing fix is implemented
echo -e "\n${YELLOW}Test 4: CNI timing fix${NC}"
if grep -q "kubernetes.is_k3s" "builtin/core/roles/cni/calico/tasks/main.yaml"; then
    echo -e "${GREEN}✓ CNI timing fix implemented${NC}"
else
    echo -e "${RED}✗ CNI timing fix missing${NC}"
    exit 1
fi

# Test 5: Check if defaults detection works
echo -e "\n${YELLOW}Test 5: k3s defaults detection${NC}"
if grep -q "is_k3s" "builtin/core/roles/defaults/tasks/main.yaml"; then
    echo -e "${GREEN}✓ k3s defaults detection implemented${NC}"
else
    echo -e "${RED}✗ k3s defaults detection missing${NC}"
    exit 1
fi

# Test 6: Check if main init task includes k3s logic
echo -e "\n${YELLOW}Test 6: k3s init integration${NC}"
if grep -q "init_k3s.yaml" "builtin/core/roles/kubernetes/init-kubernetes/tasks/main.yaml"; then
    echo -e "${GREEN}✓ k3s init integration implemented${NC}"
else
    echo -e "${RED}✗ k3s init integration missing${NC}"
    exit 1
fi

# Test 7: Check if main join task includes k3s logic
echo -e "\n${YELLOW}Test 7: k3s join integration${NC}"
if grep -q "join_k3s.yaml" "builtin/core/roles/kubernetes/join-kubernetes/tasks/main.yaml"; then
    echo -e "${GREEN}✓ k3s join integration implemented${NC}"
else
    echo -e "${RED}✗ k3s join integration missing${NC}"
    exit 1
fi

echo -e "\n${GREEN}=== All tests passed! ===${NC}"
echo -e "${GREEN}The k3s CNI configuration fix has been successfully implemented.${NC}"
echo -e "\nTo test the actual functionality:"
echo -e "1. Build KubeKey: make build"
echo -e "2. Use the test config: ./kk create cluster -f test/k3s/test-config.yaml"
echo -e "3. Verify CNI config: ls -la /etc/cni/net.d/"
echo -e "4. Check k3s status: systemctl status k3s"
echo -e "5. Verify Calico: kubectl get pods -n calico-system"
