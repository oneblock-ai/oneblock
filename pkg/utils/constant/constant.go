package constant

const (
	Prefix   = "oneblock.ai/"
	MLPrefix = "ml.oneblock.ai/"

	SystemNamespaceName = "oneblock-system"
	PublicNamespaceName = "oneblock-public"
	RedisSecretKeyName  = "redis-password" // #nosec G101

	AnnotationResourceStopped          = Prefix + "resourceStopped"
	AnnotationVolumeClaimTemplates     = Prefix + "volumeClaimTemplates"
	AnnotationEnabledExposeSvcKey      = Prefix + "exposeSvc"
	AnnotationClusterPolicyProviderKey = Prefix + "k8sProvider"
	AnnoModelTemplateVersionName       = Prefix + "modelTemplateVersionName"

	// kubeRay constant
	LabelRaySchedulerName           = "ray.io/scheduler-name"
	AnnotationRayClusterEnableGCS   = MLPrefix + "rayClusterEnableGCS"
	AnnotationRayClusterInitialized = MLPrefix + "rayClusterInitialized"
	AnnotationRayFTEnabledKey       = "ray.io/ft-enabled"
	RayRedisCleanUpFinalizer        = "ray.io/gcs-ft-redis-cleanup-finalizer"

	// Volcano constant
	VolcanoSchedulerName  = "volcano"
	LabelVolcanoQueueName = "volcano.sh/queue-name"

	AnnotationDefaultSchedulingKey             = "scheduling.oneblock.ai/isDefaultQueue"
	AnnotationSchedulingSupportedNamespacesKey = "scheduling.oneblock.ai/supportedNamespaces"
)
