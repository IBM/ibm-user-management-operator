package yamls

var OperandRBACs = []string{
	USER_MGMT_OPERAND_RB,
	USER_MGMT_OPERAND_ROLE,
	USER_MGMT_OPERAND_SA,
}

const USER_MGMT_OPERAND_ROLE = `
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
	name: user-mgmt-operand-role
rules:
  - verbs:
      - get
      - list
      - watch
    apiGroups:
      - ''
    resources:
      - secrets
      - configmaps
      - pods
  - verbs:
      - create
      - update
      - delete
    apiGroups:
      - ''
    resources:
      - secrets
`

const USER_MGMT_OPERAND_SA = `
kind: ServiceAccount
apiVersion: v1
metadata:
  name: user-mgmt-operand-serviceaccount
  labels:
	app.kubernetes.io/instance: user-mgmt-operand-serviceaccount
`

const USER_MGMT_OPERAND_RB = `
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: user-mgmt-operand-rolebinding
subjects:
  - kind: ServiceAccount
    name: user-mgmt-operand-serviceaccount
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: user-mgmt-operand-role
`
