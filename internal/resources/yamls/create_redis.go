package yamls

var REDIS_CERTS = []string{
	REDIS_CA_ISSUER,
	REDIS_CA_CERT,
	REDIS_SVC_CERT,
}

var RedisCRTemplate = `
apiVersion: redis.ibm.com/v1
kind: Rediscp
metadata:
  name: {{ .NamePrefix }}ui-redis
spec:
  size: {{.RedisCRSize}}
  license:
     accept: true
  cert_name: {{ .NamePrefix }}ui-redis-svc-tls-cert
  shutdown: false
  scale_config: medium
  version: {{.RedisCRVersion}}
`

var REDIS_CA_ISSUER = `
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: {{ .NamePrefix }}ui-redis-ca-issuer
spec:
  ca:
    secretName: cs-ca-certificate-secret
`

var REDIS_CA_CERT = `
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: {{ .NamePrefix }}ui-redis-ca-cert
spec:
  isCA: true
  commonName: {{ .NamePrefix }}ui-redis-ca-cert
  secretName: {{ .NamePrefix }}ui-redis-ca-cert
  duration: 87660h0m0s
  renewBefore: 85500h0m0s
  issuerRef:
  name: {{ .NamePrefix }}ui-redis-ca-issuer
    kind: Issuer
    group: cert-manager.io
`

var REDIS_SVC_CERT = `
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: {{ .NamePrefix }}ui-redis-svc-tls-cert
spec:
  commonName: {{ .NamePrefix }}ui-redis-svc-tls-cert
  secretName: {{ .NamePrefix }}ui-redis-svc-tls-cert
  issuerRef:
    name: {{ .NamePrefix }}ui-redis-ca-issuer
    kind: Issuer
    group: cert-manager.io
`
