package yamls

var IMConfigYamls = []string{
	IM_CONFIG_JOB,
}

var IM_CONFIG_JOB = `
apiVersion: batch/v1
kind: Job
metadata:
  name: mcsp-im-config-job
  labels:
    app: mcsp-im-config-job
spec:
  template:
    metadata:
      labels:
        app: mcsp-im-config-job
    spec:
      containers:
      - name: mcsp-im-config-job
        image: MCSP_IM_CONFIG_JOB_IMAGE
        command: ["./mcsp-im-config-job"]
        imagePullPolicy: Always
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            drop:
              - ALL
          runAsNonRoot: true
          seccompProfile:
            type: RuntimeDefault
        env:
          - name: LOG_LEVEL
            value: debug
          - name: NAMESPACE
            value: {{ .AccountIAMNamespace }}
          - name: IM_HOST_URL
            value: {{ .IAMHOSTURL }}
          - name: ACCOUNT_IAM_URL
            value: {{ .AccountIAMURL }}
      serviceAccountName: user-mgmt-operand-serviceaccount
      restartPolicy: OnFailure
`
