package constant

const (
	prefix   = "oneblock.ai/"
	mlPrefix = "ml.oneblock.ai/"

	SystemNamespaceName = "oneblock-system"
	PublicNamespaceName = "oneblock-public"
	RedisSecretKeyName  = "redis-password" // #nosec G101

	AnnotationResourceStopped          = prefix + "resourceStopped"
	AnnotationVolumeClaimTemplates     = prefix + "volumeClaimTemplates"
	AnnotationEnabledExposeSvcKey      = prefix + "exposeSvc"
	AnnotationClusterPolicyProviderKey = prefix + "k8sProvider"

	// kubeRay constant
	LabelRaySchedulerName           = "ray.io/scheduler-name"
	AnnotationRayClusterEnableGCS   = mlPrefix + "rayClusterEnableGCS"
	AnnotationRayClusterInitialized = mlPrefix + "rayClusterInitialized"
	AnnotationRayFTEnabledKey       = "ray.io/ft-enabled"

	// Volcano constant
	VolcanoSchedulerName  = "volcano"
	LabelVolcanoQueueName = "volcano.sh/queue-name"

	AnnotationDefaultSchedulingKey             = "scheduling.oneblock.ai/isDefaultQueue"
	AnnotationSchedulingSupportedNamespacesKey = "scheduling.oneblock.ai/supportedNamespaces"
)
