package notebook

import (
	"encoding/json"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mlv1 "github.com/oneblock-ai/oneblock/pkg/apis/ml.oneblock.ai/v1"
	"github.com/oneblock-ai/oneblock/pkg/utils/constant"
)

// createPVCFromAnnotation helps to create PVCs from the resource annotation
func (h *Handler) createPVCFromAnnotation(_ string, notebook *mlv1.Notebook) (*mlv1.Notebook, error) {
	if notebook == nil || notebook.DeletionTimestamp != nil {
		return nil, nil
	}

	pvcTemplates, ok := notebook.Annotations[constant.AnnotationVolumeClaimTemplates]
	if !ok || pvcTemplates == "" {
		return nil, nil
	}

	var pvcs []*corev1.PersistentVolumeClaim
	if err := json.Unmarshal([]byte(pvcTemplates), &pvcs); err != nil {
		return nil, err
	}

	var (
		pvc *corev1.PersistentVolumeClaim
		err error
	)

	for _, annoPVC := range pvcs {
		if pvc, err = h.pvcCache.Get(notebook.Namespace, annoPVC.Name); err != nil {
			if apierrors.IsNotFound(err) {
				annoPVC.Namespace = notebook.Namespace
				annoPVC.OwnerReferences = []metav1.OwnerReference{
					{
						APIVersion: notebook.APIVersion,
						Kind:       notebook.Kind,
						Name:       notebook.Name,
						UID:        notebook.UID,
					},
				}
				if _, err = h.pvcs.Create(annoPVC); err != nil {
					return nil, err
				}
				continue
			}
			return nil, err
		}

		// users may also resize the volumes outside the annotation. In that case, we can't track the update.
		// If the storage request in the notebook annotation is less or equal to the actual PVC size, just skip it.
		if annoPVC.Spec.Resources.Requests.Storage().Cmp(*pvc.Spec.Resources.Requests.Storage()) <= 0 {
			continue
		}

		toUpdate := pvc.DeepCopy()
		toUpdate.Spec.Resources.Requests = annoPVC.Spec.Resources.Requests
		if !reflect.DeepEqual(toUpdate, pvc) {
			if _, err = h.pvcs.Update(toUpdate); err != nil {
				return nil, err
			}
		}
	}
	return nil, nil
}
