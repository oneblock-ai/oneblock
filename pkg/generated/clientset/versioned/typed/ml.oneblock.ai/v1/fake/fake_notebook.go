/*
Copyright 2024 1block.ai.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by main. DO NOT EDIT.

package fake

import (
	"context"

	v1 "github.com/oneblock-ai/oneblock/pkg/apis/ml.oneblock.ai/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeNotebooks implements NotebookInterface
type FakeNotebooks struct {
	Fake *FakeMlV1
	ns   string
}

var notebooksResource = v1.SchemeGroupVersion.WithResource("notebooks")

var notebooksKind = v1.SchemeGroupVersion.WithKind("Notebook")

// Get takes name of the notebook, and returns the corresponding notebook object, and an error if there is any.
func (c *FakeNotebooks) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.Notebook, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(notebooksResource, c.ns, name), &v1.Notebook{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.Notebook), err
}

// List takes label and field selectors, and returns the list of Notebooks that match those selectors.
func (c *FakeNotebooks) List(ctx context.Context, opts metav1.ListOptions) (result *v1.NotebookList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(notebooksResource, notebooksKind, c.ns, opts), &v1.NotebookList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1.NotebookList{ListMeta: obj.(*v1.NotebookList).ListMeta}
	for _, item := range obj.(*v1.NotebookList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested notebooks.
func (c *FakeNotebooks) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(notebooksResource, c.ns, opts))

}

// Create takes the representation of a notebook and creates it.  Returns the server's representation of the notebook, and an error, if there is any.
func (c *FakeNotebooks) Create(ctx context.Context, notebook *v1.Notebook, opts metav1.CreateOptions) (result *v1.Notebook, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(notebooksResource, c.ns, notebook), &v1.Notebook{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.Notebook), err
}

// Update takes the representation of a notebook and updates it. Returns the server's representation of the notebook, and an error, if there is any.
func (c *FakeNotebooks) Update(ctx context.Context, notebook *v1.Notebook, opts metav1.UpdateOptions) (result *v1.Notebook, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(notebooksResource, c.ns, notebook), &v1.Notebook{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.Notebook), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeNotebooks) UpdateStatus(ctx context.Context, notebook *v1.Notebook, opts metav1.UpdateOptions) (*v1.Notebook, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(notebooksResource, "status", c.ns, notebook), &v1.Notebook{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.Notebook), err
}

// Delete takes name of the notebook and deletes it. Returns an error if one occurs.
func (c *FakeNotebooks) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(notebooksResource, c.ns, name, opts), &v1.Notebook{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeNotebooks) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(notebooksResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1.NotebookList{})
	return err
}

// Patch applies the patch and returns the patched notebook.
func (c *FakeNotebooks) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.Notebook, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(notebooksResource, c.ns, name, pt, data, subresources...), &v1.Notebook{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.Notebook), err
}
