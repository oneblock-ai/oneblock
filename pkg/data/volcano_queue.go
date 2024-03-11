package data

import (
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	schedulingv1 "volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"github.com/oneblock-ai/oneblock/pkg/server/config"
	"github.com/oneblock-ai/oneblock/pkg/utils/constant"
)

// addDefaultQueue adds the default queue to the cluster
func addDefaultQueue(mgmt *config.Management) error {
	// check if default queue object exist
	queue, err := mgmt.SchedulingFactory.Scheduling().V1beta1().Queue().Get(constant.DefaultQueueName, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return nil
	}

	if queue.Name == constant.DefaultQueueName {
		logrus.Debugf("Default queue %s already exists", queue.Name)
		return nil
	}

	queue = getQueueObject()

	return mgmt.Apply.
		WithDynamicLookup().
		WithSetID("add-volcano-queue").
		ApplyObjects(queue)
}

func getQueueObject() *schedulingv1.Queue {
	return &schedulingv1.Queue{
		ObjectMeta: metav1.ObjectMeta{
			Name:        constant.DefaultQueueName,
			Annotations: GetDefaultQueueAnnotations("all"),
		},
		Spec: schedulingv1.QueueSpec{
			Weight:      1,
			Reclaimable: pointer.Bool(true),
		},
	}
}

func GetDefaultQueueAnnotations(namespaces string) map[string]string {
	return map[string]string{
		constant.AnnotationDefaultSchedulingKey:             "true",
		constant.AnnotationSchedulingSupportedNamespacesKey: namespaces,
	}
}
