package fakeclients

import (
	"context"

	"github.com/rancher/wrangler/v2/pkg/generic"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
	scheduling "volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	ctlschedulv1 "github.com/oneblock-ai/oneblock/pkg/generated/clientset/versioned/typed/scheduling.volcano.sh/v1beta1"
)

type QueueClient func() ctlschedulv1.QueueInterface

func (q QueueClient) Create(queue *scheduling.Queue) (*scheduling.Queue, error) {
	return q().Create(context.TODO(), queue, metav1.CreateOptions{})
}

func (q QueueClient) Update(queue *scheduling.Queue) (*scheduling.Queue, error) {
	return q().Update(context.TODO(), queue, metav1.UpdateOptions{})
}

func (q QueueClient) UpdateStatus(queue *scheduling.Queue) (*scheduling.Queue, error) {
	return q().UpdateStatus(context.TODO(), queue, metav1.UpdateOptions{})
}

func (q QueueClient) Delete(name string, options *metav1.DeleteOptions) error {
	return q().Delete(context.TODO(), name, *options)
}

func (q QueueClient) Get(name string, options metav1.GetOptions) (*scheduling.Queue, error) {
	return q().Get(context.TODO(), name, options)
}

func (q QueueClient) List(opts metav1.ListOptions) (*scheduling.QueueList, error) {
	return q().List(context.TODO(), opts)
}

func (q QueueClient) Watch(_ metav1.ListOptions) (watch.Interface, error) {
	//TODO implement me
	panic("implement me")
}

func (q QueueClient) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *scheduling.Queue, err error) {
	return q().Patch(context.TODO(), name, pt, data, metav1.PatchOptions{}, subresources...)
}

func (q QueueClient) WithImpersonation(_ rest.ImpersonationConfig) (generic.NonNamespacedClientInterface[*scheduling.Queue, *scheduling.QueueList], error) {
	//TODO implement me
	panic("implement me")
}

type QueueCache func() ctlschedulv1.QueueInterface

func (q QueueCache) Get(name string) (*scheduling.Queue, error) {
	return q().Get(context.TODO(), name, metav1.GetOptions{})
}

func (q QueueCache) List(selector labels.Selector) ([]*scheduling.Queue, error) {
	list, err := q().List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		return nil, err
	}
	result := make([]*scheduling.Queue, 0, len(list.Items))
	for i := range list.Items {
		result = append(result, &list.Items[i])
	}
	return result, err
}

func (q QueueCache) AddIndexer(_ string, _ generic.Indexer[*scheduling.Queue]) {
	//TODO implement me
	panic("implement me")
}

func (q QueueCache) GetByIndex(_, _ string) ([]*scheduling.Queue, error) {
	//TODO implement me
	panic("implement me")
}
