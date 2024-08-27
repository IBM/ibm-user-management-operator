package yamls

var ACCOUNT_IAM_RES = []string{
	ACCOUNT_IAM_CA_ISSUER,
	ACCOUNT_IAM_SS_ISSUER,
	ACCOUNT_IAM_CA_CERT,
	ACCOUNT_IAM_SS_CERT,
	ACCOUNT_IAM_SERVICE_ACCOUNT,
	ACCOUNT_IAM_SERVICE,
	ACCOUNT_IAM_DEPLOYMENT,
	ACCOUNT_IAM_ROUTE,
}

var ACCOUNT_IAM_CA_ISSUER = `
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: wlo-ca-issuer
spec:
  ca:
    secretName: wlo-ca-tls
`

var ACCOUNT_IAM_SS_ISSUER = `
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: wlo-self-signed
spec:
  selfSigned: {}
`

var ACCOUNT_IAM_CA_CERT = `
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: account-iam-svc-tls-cm
  annotations:
    argocd.argoproj.io/sync-wave: '3'
spec:
  commonName: account-iam.mcsp1.svc
  dnsNames:
    - account-iam.mcsp1.svc
    - account-iam.mcsp1.svc.cluster.local
  duration: 2160h0m0s
  issuerRef:
    name: wlo-ca-issuer
  secretName: account-iam-svc-tls-cm
`

var ACCOUNT_IAM_SS_CERT = `
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: wlo-ca-cert
spec:
  commonName: User Management Operator
  duration: 8766h0m0s
  isCA: true
  issuerRef:
    name: wlo-self-signed
  secretName: wlo-ca-tls
`

var ACCOUNT_IAM_SERVICE_ACCOUNT = `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: account-iam
  annotations:
    argocd.argoproj.io/sync-wave: "3"
  labels:
    app.kubernetes.io/component: backend
    app.kubernetes.io/instance: account-iam
    app.kubernetes.io/name: account-iam
    app.kubernetes.io/part-of: account-iam
    bcdr-candidate: t
    component-name: iam-services
    for-product: all
    name: account-iam
`

var ACCOUNT_IAM_SERVICE = `
apiVersion: v1
kind: Service
metadata:
  name: account-iam
  annotations:
    argocd.argoproj.io/sync-wave: "3"
    service.kubernetes.io/topology-aware-hints: Auto
    service.kubernetes.io/topology-mode: Auto
  labels:
    app.kubernetes.io/component: backend
    app.kubernetes.io/instance: account-iam
    app.kubernetes.io/name: account-iam
    app.kubernetes.io/part-of: account-iam
    bcdr-candidate: t
    component-name: iam-services
    for-product: all
spec:
  ports:
  - name: 9445-tcp
    port: 9445
    protocol: TCP
    targetPort: 9445
  selector:
    app.kubernetes.io/instance: account-iam
  sessionAffinity: None
  type: ClusterIP
`

var ACCOUNT_IAM_DEPLOYMENT = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: account-iam
  labels:
    app.kubernetes.io/component: backend
    app.kubernetes.io/instance: account-iam
    app.kubernetes.io/name: account-iam
    app.kubernetes.io/part-of: account-iam
    bcdr-candidate: t
    component-name: iam-services
    for-product: all
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/instance: account-iam
  strategy:
    rollingUpdate:
      maxSurge: 25%
      maxUnavailable: 25%
    type: RollingUpdate
  template:
    metadata:
      labels:
        app.kubernetes.io/component: backend
        app.kubernetes.io/instance: account-iam
        app.kubernetes.io/name: account-iam
        app.kubernetes.io/part-of: account-iam
        bcdr-candidate: t
        component-name: iam-services
        for-product: all
    spec:
      containers:
      - name: app
        image: icr.io/automation-saas-platform/access-management/account-iam:20240819135403-development-3f8f7784573adb1cc350ff25fdfc14c9b7a640f1
        imagePullPolicy: Always
        env:
        - name: cert_defaultKeyStore
          value: /var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt
        - name: TLS_DIR
          value: /etc/x509/certs
        - name: SA_RESOURCE_VERSION
          value: "230752380"
        - name: WLP_LOGGING_CONSOLE_LOGLEVEL
          value: info
        - name: WLP_LOGGING_CONSOLE_SOURCE
          value: message,accessLog,ffdc,audit
        - name: WLP_LOGGING_CONSOLE_FORMAT
          value: json
        - name: SEC_IMPORT_K8S_CERTS
          value: "true"
        - name: SERVICE_CERT_SECRET_RESOURCE_VERSION
          value: "230832362"
        envFrom:
        - configMapRef:
            name: account-iam-env-configmap-dev
        ports:
        - containerPort: 9445
          name: 9445-tcp
          protocol: TCP
        livenessProbe:
          httpGet:
            path: /api/2.0/health/liveness
            port: 9445
            scheme: HTTPS
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /api/2.0/health/readiness
            port: 9445
            scheme: HTTPS
          periodSeconds: 10
        startupProbe:
          httpGet:
            path: /api/2.0/health/started
            port: 9445
            scheme: HTTPS
          periodSeconds: 10
        resources:
          limits:
            cpu: 1500m
            memory: 800Mi
          requests:
            cpu: 300m
            memory: 400Mi
        securityContext:
          allowPrivilegeEscalation: false
          runAsNonRoot: true
        volumeMounts:
        - mountPath: /var/run/secrets/tokens
          name: account-iam-token
        - mountPath: /config/variables/oidc
          name: account-iam-oidc
          readOnly: true
        - mountPath: /config/variables/okd
          name: account-iam-okd
          readOnly: true
        - mountPath: /config/variables
          name: account-iam-variables
          readOnly: true
        - mountPath: /etc/x509/certs
          name: svc-certificate
          readOnly: true
      serviceAccountName: account-iam
      volumes:
      - name: account-iam-token
        projected:
          sources:
          - serviceAccountToken:
              path: account-iam-token
      - name: account-iam-oidc
        secret:
          secretName: account-iam-oidc-client-auth
      - name: account-iam-okd
        secret:
          secretName: account-iam-okd-auth
      - name: account-iam-variables
        projected:
          sources:
          - secret:
              name: account-iam-database-secret
          - secret:
              name: account-iam-mpconfig-secrets
      - name: svc-certificate
        secret:
          secretName: account-iam-svc-tls-cm
`

var ACCOUNT_IAM_ROUTE = `
apiVersion: route.openshift.io/v1
kind: Route
metadata:
  name: account-iam
  annotations:
    argocd.argoproj.io/sync-wave: "3"
    openshift.io/host.generated: "true"
  labels:
    app.kubernetes.io/component: backend
    app.kubernetes.io/instance: account-iam
    app.kubernetes.io/name: account-iam
    app.kubernetes.io/part-of: account-iam
    bcdr-candidate: t
    component-name: iam-services
    for-product: all
spec:
  host: account-iam-mcsp1.apps.cutie1.cp.fyre.ibm.com
  port:
    targetPort: 9445-tcp
  tls:
    termination: reencrypt
  to:
    kind: Service
    name: account-iam
    weight: 100
  wildcardPolicy: None
`
