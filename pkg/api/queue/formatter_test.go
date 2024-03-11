package queue

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	scheduling "volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"github.com/oneblock-ai/oneblock/pkg/generated/clientset/versioned/fake"
	"github.com/oneblock-ai/oneblock/pkg/utils/constant"
	"github.com/oneblock-ai/oneblock/pkg/utils/fakeclients"
)

var (
	defaultQueue = &scheduling.Queue{
		ObjectMeta: metav1.ObjectMeta{
			Name: "default-queue",
		},
	}
	toRemoveDefaultQueues = []*scheduling.Queue{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "non-default-queue1",
				Annotations: map[string]string{
					constant.AnnotationDefaultSchedulingKey: "true",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "non-default-queue2",
				Annotations: map[string]string{
					constant.AnnotationDefaultSchedulingKey: "true",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "non-default-queue3",
				Annotations: map[string]string{
					constant.AnnotationDefaultSchedulingKey: "true",
				},
			},
		},
	}
)

func Test_setDefaultQueueAction(t *testing.T) {
	type input struct {
		name  string
		queue *scheduling.Queue
	}
	type output struct {
		err error
	}
	var testCases = []struct {
		name     string
		given    input
		expected output
	}{
		{
			name: "set default queue",
			given: input{
				name:  defaultQueue.Name,
				queue: defaultQueue,
			},
			expected: output{
				err: nil,
			},
		},
	}

	for _, tc := range testCases {
		assert := require.New(t)
		typedObjects := []runtime.Object{defaultQueue, toRemoveDefaultQueues[0], toRemoveDefaultQueues[1], toRemoveDefaultQueues[2]}
		client := fake.NewSimpleClientset(typedObjects...)

		h := Handler{
			queue:      fakeclients.QueueClient(client.SchedulingV1beta1().Queues),
			queueCache: fakeclients.QueueCache(client.SchedulingV1beta1().Queues),
		}

		fakeHTTP := httptest.NewRecorder()
		err := h.setDefaultQueue(fakeHTTP, tc.given.name)
		assert.NoError(err, "expected no error while set default queue")

		queues, err := h.queueCache.List(labels.Everything())
		assert.NoError(err, "expected no error while listing queues")
		for _, q := range queues {
			if q.Name == tc.given.name {
				assert.Equal("true", q.Annotations[constant.AnnotationDefaultSchedulingKey], "expected to find default queue")
			} else {
				assert.Empty(q.Annotations[constant.AnnotationDefaultSchedulingKey], "expected to find non-default queue")
			}
		}
	}
}
