# HealthCheck for capkk

refer https://cluster-api.sigs.k8s.io/tasks/automated-machine-management/healthchecking.html

there is a sample for healthcheck

```yaml
apiVersion: cluster.x-k8s.io/v1beta1
kind: MachineHealthCheck
metadata:
  name: hc-capkk-1
spec:
  clusterName: capkk-1
  maxUnhealthy: 100%
  selector:
    matchLabels:
      cluster.x-k8s.io/cluster-name: capkk-1
  unhealthyConditions:
    - type: Ready
      status: Unknown
      timeout: 300s
    - type: Ready
      status: "False"
      timeout: 300s
```

Capkk currently does not have a remediationTemplate.