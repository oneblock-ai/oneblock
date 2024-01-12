package constant

const (
	DefaultSystemNamespace    = "oneblock-system"
	ResourceStoppedAnnotation = "core.oneblock.ai/resource-stopped"

	RedisSecretKeyName = "redis-password" // #nosec G101
	RedisSecretName    = "kuberay-redis"  // #nosec G101

	EnabledExposeSvcAnnotation      = "ml.oneblock.ai/expose-svc"
	ClusterPolicyProviderAnnotation = "ml.oneblock.ai/k8s-provider"
)
