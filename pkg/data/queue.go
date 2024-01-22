package data

import (
	"github.com/rancher/wrangler/v2/pkg/apply"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	schedulingv1 "volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"github.com/oneblock-ai/oneblock/pkg/utils/constant"
)

const defaultQueueName = "raycluster-default"

// addDefaultQueue adds the default queue to the cluster
func addDefaultQueue(apply apply.Apply) error {

	//qClient := mgmt.SchedulingFactory.Scheduling().V1beta1().Queue()

	//_, err := qClient.Get(defaultQueueName, metav1.GetOptions{})
	//if err != nil && !errors.IsNotFound(err) {
	//	return err
	//}

	resources, err := getResourceList("22", "44Gi") // calculated upon the default ray cluster
	if err != nil {
		return err
	}
	queue := getQueueObject(resources)
	//if _, err := qClient.Create(queue); err != nil {
	//	return err
	//}

	return apply.
		WithDynamicLookup().
		WithSetID("add-raycluster-queue").
		ApplyObjects(queue)
}

func getQueueObject(resources corev1.ResourceList) *schedulingv1.Queue {
	return &schedulingv1.Queue{
		ObjectMeta: metav1.ObjectMeta{
			Name:        defaultQueueName,
			Annotations: GetDefaultQueueAnnotations("all"),
		},
		Spec: schedulingv1.QueueSpec{
			Weight:      1,
			Reclaimable: pointer.Bool(true),
			Capability:  resources,
		},
	}
}

func GetDefaultQueueAnnotations(namespaces string) map[string]string {
	return map[string]string{
		constant.AnnotationDefaultSchedulingKey:             "true",
		constant.AnnotationSchedulingSupportedNamespacesKey: namespaces,
	}
}
