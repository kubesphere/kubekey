package templates

import (
	"github.com/lithammer/dedent"
	"text/template"
)

var NodeLocalDNSService = template.Must(template.New("nodelocaldns.yaml").Parse(
	dedent.Dedent(`---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: nodelocaldns
  namespace: kube-system
  labels:
    addonmanager.kubernetes.io/mode: Reconcile

---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: nodelocaldns
  namespace: kube-system
  labels:
    k8s-app: kube-dns
    addonmanager.kubernetes.io/mode: Reconcile
spec:
  selector:
    matchLabels:
      k8s-app: nodelocaldns
  template:
    metadata:
      labels:
        k8s-app: nodelocaldns
      annotations:
        prometheus.io/scrape: 'true'
        prometheus.io/port: '9253'
    spec:
      priorityClassName: system-cluster-critical
      serviceAccountName: nodelocaldns
      hostNetwork: true
      dnsPolicy: Default  # Don't use cluster DNS.
      tolerations:
      - effect: NoSchedule
        operator: "Exists"
      - effect: NoExecute
        operator: "Exists"
      - key: "CriticalAddonsOnly"
        operator: "Exists"
      containers:
      - name: node-cache
        image: {{ .NodelocaldnsImage }}
        resources:
          limits:
            memory: 170Mi
          requests:
            cpu: 100m
            memory: 70Mi
        args: [ "-localip", "169.254.25.10", "-conf", "/etc/coredns/Corefile", "-upstreamsvc", "coredns" ]
        securityContext:
          privileged: true
        ports:
        - containerPort: 53
          name: dns
          protocol: UDP
        - containerPort: 53
          name: dns-tcp
          protocol: TCP
        - containerPort: 9253
          name: metrics
          protocol: TCP
        livenessProbe:
          httpGet:
            host: 169.254.25.10
            path: /health
            port: 9254
            scheme: HTTP
          timeoutSeconds: 5
          successThreshold: 1
          failureThreshold: 10
        readinessProbe:
          httpGet:
            host: 169.254.25.10
            path: /health
            port: 9254
            scheme: HTTP
          timeoutSeconds: 5
          successThreshold: 1
          failureThreshold: 10
        volumeMounts:
        - name: config-volume
          mountPath: /etc/coredns
        - name: xtables-lock
          mountPath: /run/xtables.lock
      volumes:
        - name: config-volume
          configMap:
            name: nodelocaldns
            items:
            - key: Corefile
              path: Corefile
        - name: xtables-lock
          hostPath:
            path: /run/xtables.lock
            type: FileOrCreate
      # Minimize downtime during a rolling upgrade or deletion; tell Kubernetes to do a "force
      # deletion": https://kubernetes.io/docs/concepts/workloads/pods/pod/#termination-of-pods.
      terminationGracePeriodSeconds: 0
  updateStrategy:
    rollingUpdate:
      maxUnavailable: 20%
    type: RollingUpdate

    `)))
