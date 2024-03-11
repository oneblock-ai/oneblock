package queue

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/oneblock-ai/apiserver/v2/pkg/apierror"
	"github.com/oneblock-ai/apiserver/v2/pkg/types"
	"github.com/rancher/wrangler/v2/pkg/schemas/validation"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/oneblock-ai/oneblock/pkg/utils"
	"github.com/oneblock-ai/oneblock/pkg/utils/constant"
)

func formatter(request *types.APIRequest, resource *types.RawResource) {
	resource.Actions = make(map[string]string, 1)
	resource.AddAction(request, ActionSetDefault)
}

func (h Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if err := h.do(rw, req); err != nil {
		status := http.StatusInternalServerError
		var e *apierror.APIError
		if errors.As(err, &e) {
			status = e.Code.Status
		}
		rw.WriteHeader(status)
		_, _ = rw.Write([]byte(err.Error()))
		return
	}
	rw.WriteHeader(http.StatusNoContent)
}

func (h Handler) do(rw http.ResponseWriter, req *http.Request) error {
	vars := utils.EncodeVars(mux.Vars(req))
	if req.Method == http.MethodPost {
		return h.doPost(vars["action"], rw, req)
	}

	return apierror.NewAPIError(validation.InvalidAction, fmt.Sprintf("Unsupported method %s", req.Method))
}

func (h Handler) doPost(action string, rw http.ResponseWriter, req *http.Request) error {
	vars := utils.EncodeVars(mux.Vars(req))
	name := vars["name"]
	switch action {
	case ActionSetDefault:
		return h.setDefaultQueue(rw, name)
	default:
		return apierror.NewAPIError(validation.InvalidAction, fmt.Sprintf("Unsupported POST action %s", action))
	}
}

func (h Handler) setDefaultQueue(_ http.ResponseWriter, name string) error {
	logrus.Debugf("Set default queue %s", name)
	// check if queue exists
	queue, err := h.queueCache.Get(name)
	if err != nil {
		return err
	}

	// unset all default queues
	if err = h.unsetAllDefaultQueues(); err != nil {
		return fmt.Errorf("failed to unset all default queues: %w", err)
	}

	queueCpy := queue.DeepCopy()
	if queueCpy.Annotations == nil {
		queueCpy.Annotations = make(map[string]string, 1)
	}
	queueCpy.Annotations[constant.AnnotationDefaultSchedulingKey] = "true"
	if _, err = h.queue.Update(queueCpy); err != nil {
		return err
	}

	return nil
}

func (h Handler) unsetAllDefaultQueues() error {
	queues, err := h.queueCache.List(labels.Everything())
	if err != nil {
		return err
	}

	for _, queue := range queues {
		queueCpy := queue.DeepCopy()
		if queueCpy.Annotations == nil {
			continue
		}
		if _, ok := queueCpy.Annotations[constant.AnnotationDefaultSchedulingKey]; ok {
			delete(queueCpy.Annotations, constant.AnnotationDefaultSchedulingKey)
			if _, err = h.queue.Update(queueCpy); err != nil {
				return err
			}
		}
	}

	return nil
}
