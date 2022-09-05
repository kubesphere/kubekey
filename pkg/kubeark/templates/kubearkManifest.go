/*
 Copyright 2022 The KubeSphere Authors.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package templates

import (
	"text/template"

	"github.com/lithammer/dedent"
)

var KubearkManifest = template.Must(template.New("kubeark.yaml").Parse(
	dedent.Dedent(`
kind: Secret
apiVersion: v1
metadata:
  name: kubeark-docker-hub
  namespace: default
data:
  .dockerconfigjson: >-
    eyJhdXRocyI6eyJkb2NrZXIuaW8iOnsidXNlcm5hbWUiOiJrdWJlYXJrIiwicGFzc3dvcmQiOiI5MThhZjRmNy0zOWFiLTQ1YjUtYTkyNy01ZDUzNjhmMTU4NmEifX19
type: kubernetes.io/dockerconfigjson
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: kubeark-webapp-ingress
  annotations:
    kubernetes.io/ingress.allow-http: "true"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "false"
    cert-manager.io/cluster-issuer: dsf-clusterissuer
    cert-manager.io/acme-challenge-type: http01
    kubernetes.io/ingress.class: nginx
    nginx.ingress.kubernetes.io/ssl-redirect: "false"
    nginx.ingress.kubernetes.io/use-regex: "true"
spec:
  rules:
  - http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: kubeark-web-app-service
            port:
              number: 80
---
kind: Service
apiVersion: v1
metadata:
  name: kubeark-web-app-service
  namespace: default
  labels:
    app: kubeark-web-app
spec:
  selector:
    app: kubeark-web-app
  ports:
    - name: http
      port: 80
      targetPort: 8000
---
kind: Deployment
apiVersion: apps/v1
metadata:
  name: kubeark-web-app
  namespace: default
  generation: 24
  labels:
    app: kubeark-web-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: kubeark-web-app
  template:
    metadata:
      namespace: default
      labels:
        app: kubeark-web-app
    spec:
      containers:
        - name: kubeark
          image: 'kubeark/kubeark:latest'
          env:
          - name: FLASK_ENV
            value: "dev"
          - name: FLASK_APP
            value: run
          - name: CONTAINER_ROLE
            value: app
          - name: KUBE_TOKEN
            value: "eyJhbGciOiJSUzI1NiIsImtpZCI6ImM2Z1dQVUxvTXJaWHpBNTIyNGFtTWlEMmlqMHRCdUNRd2diOEd5dFhybXcifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZWZhdWx0Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6Imt1YmUtdG9rZW4tZzY0cTIiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC5uYW1lIjoia3ViZSIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50LnVpZCI6ImM4MTZkMjU1LWM2ODItNGZjMi05NTAzLWIxMzI5ZGEyY2Q0ZCIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZWZhdWx0Omt1YmUifQ.oOTsuvZwYTIkvro-S1G2j0S0u-_RsXxuqeDeSznjrMyJLDsga7eIh2hGdGfmTYoy-ZaF5CfPNilA_dp4qkHzA6l9gHVnbLi8f7j_e38S18ViD1HEc-ib9ZKJ4vhouJDcOFzfQ4Z5jWKFmDUmvWzj18efjSydhhRi80aa2NrzFgkeel7AowRFfQOf9tkwA2J6wv3IEnpWrngdXYlk6gfxT0Uj4elWRWcTH5qRfFGCB5XfnoDd_Nc8YOJDUhCEAzgHxiQzQgxZVom30sW0UZ4Aa-W4oupxUmqb0yLJWHZDxhHuKS-2r4wzunYget7O_rLtWP11qhCI9skFq6S8J1X2Dm5jNYeiPDiOt4dU1_KmyfydaXr3L4_Xhl3Ide8fElHTA5ZYtiG_PNaBY_xrjc9FSt0-xDVWajLoWnZ1828RWbUN3Mt6GGMEC0ccjjhfqzRPC3ZqOTIemLkfQpKk4MQ5B1zsz6BtV6pnQpCYKyPdS0fjc2rJZinTI-KTix9_sTT5q_flNACCUrDnm8iWTNL3suEkfGoAdEU2Q7BAdyKf0X8AkkBScxMkua3g6YblYxu3kh0awi3CU2UFufMgcvQUQNjQE23XpihpSYbQCXISHXddXcYDAgfbx6Ua-QZz6cszlETqwjawcEFkgTGQ2WaGnMTpDoPdaR17TrseuB7eJDI"
          - name: KUBE_APISERVER
            value: https://ka-dsf-dev-65dcf27a.hcp.westeurope.azmk8s.io
          - name: REMOTE_CLUSTER
            value: "True"
          - name: DATABASE_URL
            value: postgresql://kube-db:0%3AOUMKz5Iq%40RDs=Db6oM7-K8@postgresscluster-primary.postgres-operator.svc:5432/kubeark
          - name: REDIS_JWT_BLOCKLIST
            value: redis://kubeark-redis-service.default.svc:6379/5
          - name: CELERY_BROKER_URL
            value: redis://kubeark-redis-service.default.svc:6379/0
          - name: CELERY_RESULT_BACKEND
            value: redis://kubeark-redis-service.default.svc:6379/1
          ports:
            - containerPort: 8000
              protocol: TCP
          resources: {}
          imagePullPolicy: Always
      restartPolicy: Always
      terminationGracePeriodSeconds: 0
      serviceAccountName: kube
      serviceAccount: kube
      securityContext: {}
      imagePullSecrets:
        - name: kubeark-docker-hub
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
      maxSurge: 1
  minReadySeconds: 5
  revisionHistoryLimit: 10
  progressDeadlineSeconds: 600
---
kind: Deployment
apiVersion: apps/v1
metadata:
  name: kubeark-web-worker
  namespace: default
  generation: 24
  labels:
    app: kubeark-web-worker
spec:
  replicas: 1
  selector:
    matchLabels:
      app: kubeark-web-worker
  template:
    metadata:
      namespace: default
      labels:
        app: kubeark-web-worker
    spec:
      containers:
        - name: kubeark
          image: 'kubeark/kubeark:latest'
          cmd: ["python"]
          args: ["worker.py"]
          env:
          - name: FLASK_ENV
            value: "dev"
          - name: FLASK_APP
            value: run
          - name: CONTAINER_ROLE
            value: worker
          - name: KUBE_TOKEN
            value: "eyJhbGciOiJSUzI1NiIsImtpZCI6ImM2Z1dQVUxvTXJaWHpBNTIyNGFtTWlEMmlqMHRCdUNRd2diOEd5dFhybXcifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZWZhdWx0Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6Imt1YmUtdG9rZW4tZzY0cTIiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC5uYW1lIjoia3ViZSIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50LnVpZCI6ImM4MTZkMjU1LWM2ODItNGZjMi05NTAzLWIxMzI5ZGEyY2Q0ZCIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZWZhdWx0Omt1YmUifQ.oOTsuvZwYTIkvro-S1G2j0S0u-_RsXxuqeDeSznjrMyJLDsga7eIh2hGdGfmTYoy-ZaF5CfPNilA_dp4qkHzA6l9gHVnbLi8f7j_e38S18ViD1HEc-ib9ZKJ4vhouJDcOFzfQ4Z5jWKFmDUmvWzj18efjSydhhRi80aa2NrzFgkeel7AowRFfQOf9tkwA2J6wv3IEnpWrngdXYlk6gfxT0Uj4elWRWcTH5qRfFGCB5XfnoDd_Nc8YOJDUhCEAzgHxiQzQgxZVom30sW0UZ4Aa-W4oupxUmqb0yLJWHZDxhHuKS-2r4wzunYget7O_rLtWP11qhCI9skFq6S8J1X2Dm5jNYeiPDiOt4dU1_KmyfydaXr3L4_Xhl3Ide8fElHTA5ZYtiG_PNaBY_xrjc9FSt0-xDVWajLoWnZ1828RWbUN3Mt6GGMEC0ccjjhfqzRPC3ZqOTIemLkfQpKk4MQ5B1zsz6BtV6pnQpCYKyPdS0fjc2rJZinTI-KTix9_sTT5q_flNACCUrDnm8iWTNL3suEkfGoAdEU2Q7BAdyKf0X8AkkBScxMkua3g6YblYxu3kh0awi3CU2UFufMgcvQUQNjQE23XpihpSYbQCXISHXddXcYDAgfbx6Ua-QZz6cszlETqwjawcEFkgTGQ2WaGnMTpDoPdaR17TrseuB7eJDI"
          - name: KUBE_APISERVER
            value: https://ka-dsf-dev-65dcf27a.hcp.westeurope.azmk8s.io
          - name: REMOTE_CLUSTER
            value: "True"
          - name: DATABASE_URL
            value: postgresql://kube-db:0%3AOUMKz5Iq%40RDs=Db6oM7-K8@postgresscluster-primary.postgres-operator.svc:5432/kubeark
          - name: REDIS_JWT_BLOCKLIST
            value: redis://kubeark-redis-service.default.svc:6379/5
          - name: CELERY_BROKER_URL
            value: redis://kubeark-redis-service.default.svc:6379/0
          - name: CELERY_RESULT_BACKEND
            value: redis://kubeark-redis-service.default.svc:6379/1
          ports:
            - containerPort: 8000
              protocol: TCP
          resources: {}
          imagePullPolicy: Always
      restartPolicy: Always
      terminationGracePeriodSeconds: 0
      serviceAccountName: kube
      serviceAccount: kube
      securityContext: {}
      imagePullSecrets:
        - name: kubeark-docker-hub
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
      maxSurge: 1
  minReadySeconds: 5
  revisionHistoryLimit: 10
  progressDeadlineSeconds: 600
---
kind: Deployment
apiVersion: apps/v1
metadata:
  name: kubeark-web-scheduler
  namespace: default
  generation: 24
  labels:
    app: kubeark-web-scheduler
spec:
  replicas: 1
  selector:
    matchLabels:
      app: kubeark-web-scheduler
  template:
    metadata:
      namespace: default
      labels:
        app: kubeark-web-scheduler
    spec:
      containers:
        - name: kubeark
          image: 'kubeark/kubeark:latest'
          cmd: ["python"]
          args: ["scheduler.py"]
          env:
          - name: FLASK_ENV
            value: "dev"
          - name: FLASK_APP
            value: run
          - name: CONTAINER_ROLE
            value: scheduler
          - name: KUBE_TOKEN
            value: "eyJhbGciOiJSUzI1NiIsImtpZCI6ImM2Z1dQVUxvTXJaWHpBNTIyNGFtTWlEMmlqMHRCdUNRd2diOEd5dFhybXcifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZWZhdWx0Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6Imt1YmUtdG9rZW4tZzY0cTIiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC5uYW1lIjoia3ViZSIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50LnVpZCI6ImM4MTZkMjU1LWM2ODItNGZjMi05NTAzLWIxMzI5ZGEyY2Q0ZCIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZWZhdWx0Omt1YmUifQ.oOTsuvZwYTIkvro-S1G2j0S0u-_RsXxuqeDeSznjrMyJLDsga7eIh2hGdGfmTYoy-ZaF5CfPNilA_dp4qkHzA6l9gHVnbLi8f7j_e38S18ViD1HEc-ib9ZKJ4vhouJDcOFzfQ4Z5jWKFmDUmvWzj18efjSydhhRi80aa2NrzFgkeel7AowRFfQOf9tkwA2J6wv3IEnpWrngdXYlk6gfxT0Uj4elWRWcTH5qRfFGCB5XfnoDd_Nc8YOJDUhCEAzgHxiQzQgxZVom30sW0UZ4Aa-W4oupxUmqb0yLJWHZDxhHuKS-2r4wzunYget7O_rLtWP11qhCI9skFq6S8J1X2Dm5jNYeiPDiOt4dU1_KmyfydaXr3L4_Xhl3Ide8fElHTA5ZYtiG_PNaBY_xrjc9FSt0-xDVWajLoWnZ1828RWbUN3Mt6GGMEC0ccjjhfqzRPC3ZqOTIemLkfQpKk4MQ5B1zsz6BtV6pnQpCYKyPdS0fjc2rJZinTI-KTix9_sTT5q_flNACCUrDnm8iWTNL3suEkfGoAdEU2Q7BAdyKf0X8AkkBScxMkua3g6YblYxu3kh0awi3CU2UFufMgcvQUQNjQE23XpihpSYbQCXISHXddXcYDAgfbx6Ua-QZz6cszlETqwjawcEFkgTGQ2WaGnMTpDoPdaR17TrseuB7eJDI"
          - name: KUBE_APISERVER
            value: https://ka-dsf-dev-65dcf27a.hcp.westeurope.azmk8s.io
          - name: REMOTE_CLUSTER
            value: "True"
          - name: DATABASE_URL
            value: postgresql://kube-db:0%3AOUMKz5Iq%40RDs=Db6oM7-K8@postgresscluster-primary.postgres-operator.svc:5432/kubeark
          - name: REDIS_JWT_BLOCKLIST
            value: redis://kubeark-redis-service.default.svc:6379/5
          - name: CELERY_BROKER_URL
            value: redis://kubeark-redis-service.default.svc:6379/0
          - name: CELERY_RESULT_BACKEND
            value: redis://kubeark-redis-service.default.svc:6379/1
          ports:
            - containerPort: 8000
              protocol: TCP
          resources: {}
          imagePullPolicy: Always
      restartPolicy: Always
      terminationGracePeriodSeconds: 0
      serviceAccountName: kube
      serviceAccount: kube
      securityContext: {}
      imagePullSecrets:
        - name: kubeark-docker-hub
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
      maxSurge: 1
  minReadySeconds: 5
  revisionHistoryLimit: 10
  progressDeadlineSeconds: 600
---
apiVersion: v1
kind: Service
metadata:
  name: kubeark-redis-service
  namespace: default
  labels:
    app: redis
spec:
  ports:
    - name: http
      protocol: TCP
      port: 6379
      targetPort: 6379
  selector:
    app: redis
  type: ClusterIP
  sessionAffinity: None
  ipFamilies:
    - IPv4
  ipFamilyPolicy: SingleStack
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
  labels:
    app: redis
    tier: kubeark-backend
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
        tier: kubeark-backend
    spec:
      containers:
      - name: redis
        image: "docker.io/redis:6.0.5"
        resources:
          requests:
            cpu: 100m
            memory: 100Mi
        ports:
        - containerPort: 6379
`)))
