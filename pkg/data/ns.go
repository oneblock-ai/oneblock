package data

import (
	"github.com/rancher/wrangler/v2/pkg/apply"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultPublicNamespace = "ml-public"
)

func addPublicNamespace(apply apply.Apply) error {
	// add public namespace for all authenticated users
	namespace := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: defaultPublicNamespace},
	}
	return apply.
		WithDynamicLookup().
		WithSetID("oneblock-public").
		ApplyObjects(namespace)
}
