# Renew Certificates (certs_renew.yaml)

`certs_renew.yaml` is used to automatically renew service/workload certificates for the Kubernetes cluster (**CA certificates are not included**).

## Execution Flow

1. **Global Initialization**
   - Execute the `native/root` role on all nodes.

2. **Load Default Variables**
   - Load the `defaults` role on all nodes.

3. **Certificate Initialization**
   - Execute `certs/init` on `localhost` to prepare the CA and configuration required for certificate renewal.

4. **Execute Renewal**
   - Execute the `certs/renew` role on all nodes to automatically detect and renew certificates that are about to expire or have already expired.

## Notes

- CA root certificates will not be renewed automatically. To replace the CA, please handle it manually or recreate the cluster.
- It is recommended to run this playbook before certificates are close to expiration to avoid service interruption.
- etcd service restarts only when Kubekey detects a systemd-managed `etcd.service` on the target node; otherwise certificate files are updated and the operator should restart the service through the node's own management layer.
