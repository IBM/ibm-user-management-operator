package yamls

var APP_SECRETS = []string{
	ClientAuth,
	OKD_Auth,
	DatabaseSecret,
	MpConfig,
}

var APP_STATIC_YAMLS = []string{
	INGRESS,
	EGRESS,
	CONFIG_ENV,
	DB_MIGRATION_MCSPID_SA,
	DB_MIGRATION_MCSPID,
}

var ClientAuth = `
kind: Secret
apiVersion: v1
metadata:
  name: {{ .NamePrefix }}oidc-client-auth
  namespace: {{ .Namespace }}
  labels:
    by-squad: mcsp-user-management
    for-product: all
    bcdr-candidate: t
    component-name: iam-services
data:
  realm: {{ .Realm }}
  client_id: {{ .ClientID }}
  client_secret: {{ .ClientSecret }}
stringData:
  discovery_endpoint: {{ .DiscoveryEndpoint }}
type: Opaque
`
var OKD_Auth = `
kind: Secret
apiVersion: v1
metadata:
  name: {{ .NamePrefix }}okd-auth
  labels:
    by-squad: mcsp-user-management
    for-product: all
    bcdr-candidate: t
    component-name: iam-services
data:
  user_validation_api_v2: {{ .UserValidationAPIV2 }}
type: Opaque
`

var DatabaseSecret = `
kind: Secret
apiVersion: v1
metadata:
  name: {{ .NamePrefix }}database-secret
  labels:
    by-squad: mcsp-user-management
    for-product: all
    bcdr-candidate: t
    component-name: iam-services
stringData:
  pg_jdbc_host: common-service-db-rw
  pg_jdbc_port: "5432"
  pg_db_name: account_iam
  pg_db_schema: accountiam
  pg_db_user: user_accountiam
  pg_jdbc_password_jndi: "jdbc/iamdatasource"
  pg_ssl_mode: prefer
  GLOBAL_ACCOUNT_IDP: {{ .GlobalAccountIDP }}
data:
  pgPassword: {{ .PGPassword }}
  GLOBAL_ACCOUNT_AUD: {{ .GlobalAccountAud }}
  GLOBAL_ACCOUNT_REALM: {{ .GlobalRealmValue }}
  ENCRYPTION_KEYS: {{ .EncryptionKeys }}
  CURRENT_ENCRYPTION_KEY_NUM: {{ .CurrentEncryptionKeyNum }}
type: Opaque
`

var MpConfig = `
kind: Secret
apiVersion: v1
metadata:
  name: {{ .NamePrefix }}mpconfig-secrets
  labels:
    by-squad: mcsp-user-management
    for-product: all
    bcdr-candidate: t
    component-name: iam-services
data:
  DEFAULT_AUD_VALUE: {{ .DefaultAUDValue }}
  DEFAULT_REALM_VALUE: {{ .DefaultRealmValue }}
  SRE_MCSP_GROUPS_TOKEN: {{ .SREMCSPGroupsToken }}
stringData:
  DEFAULT_IDP_VALUE: {{ .DefaultIDPValue }}
type: Opaque
`

const INGRESS = `
kind: NetworkPolicy
apiVersion: networking.k8s.io/v1
metadata:
  name: {{ .NamePrefix }}ingress-allow
  labels:
    by-squad: mcsp-user-management
    for-product: all
    bcdr-candidate: t
    component-name: iam-services
spec:
  podSelector:
    matchLabels:
      name: {{ .NamePrefix }}account-iam
  ingress:
    - ports:
        # calls to the API
        - protocol: TCP
          port: 9445
  policyTypes:
    - Ingress
`

const EGRESS = `
kind: NetworkPolicy
apiVersion: networking.k8s.io/v1
metadata:
  name: {{ .NamePrefix }}egress-allow
  labels:
    bcdr-candidate: t
    component-name: iam-services
    by-squad: mcsp-user-management
    for-product: all
spec:
  podSelector:
    matchLabels:
      name: {{ .NamePrefix }}account-iam
  policyTypes:
    - Egress
  egress:
    # calls to openshift's dns
    - to:
        - namespaceSelector:
            matchLabels:
              kubernetes.io/metadata.name: openshift-dns
    # other calls
    - ports:
        # default https calls
        - protocol: TCP
          port: 443
        # okd external route - temporary
        - protocol: TCP
          port: 6443
        # Instana agent
        - protocol: TCP
          port: 42699
        # Ephemeral port range for Instana JVM agent
        - protocol: TCP
          port: 32768
          endPort: 65535
        # Port for PG DB
        - protocol: TCP
          port: 5432
`

const CONFIG_ENV = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .NamePrefix }}env-configmap-development
  labels:
    by-squad: mcsp-user-management
    for-product: all
    bcdr-candidate: t
    component-name: iam-services
data:
  CLOUD_INSTANCE_ID: dev
  CLOUD_REGION: dev
  NOTIFICATION_SERVICE_ENABLED: "false"
  LOCAL_TOKEN_ISSUER: https://127.0.0.1:9443/oidc/endpoint/OP
  TOKEN_EXCHANGE_VALIDATE_ROLES: "true"
`

const DB_MIGRATION_MCSPID = `
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ .NamePrefix }}db-migration
  labels:
    by-squad: mcsp-user-management
    for-product: all
    bcdr-candidate: t
    component-name: iam-services
  annotations:
    test: mcspid
spec:
  backoffLimit: 3
  activeDeadlineSeconds: 21600
  template:
    metadata:
      labels:
        by-squad: mcsp-user-management
        for-product: all
        name: {{ .NamePrefix }}account-iam
    spec:
      restartPolicy: Never
      containers:
        - name: dbmigrate
          image: RELATED_IMAGE_ACCOUNT_IAM
          envFrom:
            - secretRef:
                name: {{ .NamePrefix }}database-secret
          command:
            - /bin/sh
            - '-c'
          args:
            - '/dbmigration/run.sh'
          volumeMounts:
          imagePullPolicy: Always
          resources:
            requests:
              cpu: 100m
              memory: 300Mi
            limits:
              cpu: 500m
              memory: 600Mi
  serviceAccountName: {{ .NamePrefix }}migration
      volumes:
        - name: account-iam-token
          projected:
            sources:
              - serviceAccountToken:
                  audience: openshift
                  expirationSeconds: 7200
                  path: account-iam-token
            defaultMode: 420
`

const DB_MIGRATION_MCSPID_SA = `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{ .NamePrefix }}migration
  labels:
    by-squad: mcsp-user-management
    for-product: all
`
