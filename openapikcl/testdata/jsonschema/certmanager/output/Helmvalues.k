# No schema imports needed - schemas in same directory

schema Helmvalues:
    nameOverride?: str
    prometheus?: {str:any}
    podLabels?: {str:any} = {}
    http_proxy?: str
    affinity?: {str:any} = {}
    podDisruptionBudget?: {str:any}
    namespace?: str = ""
    dns01RecursiveNameservers?: str = ""
    resources?: {str:any} = {}
    volumes?: [any] = []
    podDnsPolicy?: str
    strategy?: {str:any} = {}
    cainjector?: {str:any}
    ingressShim?: {str:any}
    clusterResourceNamespace?: str = ""
    serviceIPFamilyPolicy?: str
    creator?: str = "helm"
    crds?: {str:any}
    acmesolver?: {str:any}
    maxConcurrentChallenges?: float = 60
    extraEnv?: [any] = []
    webhook?: {str:any}
    tolerations?: [any] = []
    enableServiceLinks?: bool = False
    featureGates?: str = ""
    approveSignerNames?: [any] = ["issuers.cert-manager.io/*", "clusterissuers.cert-manager.io/*"]
    podAnnotations?: {str:any}
    enabled?: bool = True
    fullnameOverride?: str
    no_proxy?: str
    volumeMounts?: [any] = []
    https_proxy?: str
    disableAutoApproval?: bool = False
    automountServiceAccountToken?: bool
    extraArgs?: [any] = []
    hostAliases?: [any] = []
    dns01RecursiveNameserversOnly?: bool = False
    containerSecurityContext?: {str:any} = {allowPrivilegeEscalation: False, capabilities: {drop: ["ALL"]}, readOnlyRootFilesystem: True}
    serviceAccount?: {str:any}
    startupapicheck?: {str:any}
    serviceLabels?: {str:any}
    serviceAnnotations?: {str:any}
    extraObjects?: [any] = []
    installCRDs?: bool = False
    global?: {str:any}
    livenessProbe?: {str:any} = {periodSeconds: 10, successThreshold: 1, timeoutSeconds: 15, enabled: True, failureThreshold: 8, initialDelaySeconds: 10}
    topologySpreadConstraints?: [any] = []
    deploymentAnnotations?: {str:any}
    enableCertificateOwnerRef?: bool = False
    nodeSelector?: {str:any} = {"kubernetes.io/os": "linux"}
    serviceIPFamilies?: [any]
    securityContext?: {str:any} = {runAsNonRoot: True, seccompProfile: {type: "RuntimeDefault"}}
    replicaCount?: float = 1
    podDnsConfig?: {str:any}
    config?: {str:any} = {}
    image?: {str:any}