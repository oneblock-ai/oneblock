package constant

const (
	DefaultSystemNamespace = "oneblock-system"

	RedisSecretKeyName = "redis-password" // #nosec G101
	RedisSecretName    = "kuberay-redis"  // #nosec G101

	EnabledExposeSvcAnnotation = "ml.oneblock.ai/expose-svc"
)
