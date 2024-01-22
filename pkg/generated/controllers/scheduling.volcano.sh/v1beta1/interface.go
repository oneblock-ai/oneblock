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

package v1beta1

import (
	"github.com/rancher/lasso/pkg/controller"
	"github.com/rancher/wrangler/v2/pkg/generic"
	"github.com/rancher/wrangler/v2/pkg/schemes"
	"k8s.io/apimachinery/pkg/runtime/schema"
	v1beta1 "volcano.sh/apis/pkg/apis/scheduling/v1beta1"
)

func init() {
	schemes.Register(v1beta1.AddToScheme)
}

type Interface interface {
	PodGroup() PodGroupController
	Queue() QueueController
}

func New(controllerFactory controller.SharedControllerFactory) Interface {
	return &version{
		controllerFactory: controllerFactory,
	}
}

type version struct {
	controllerFactory controller.SharedControllerFactory
}

func (v *version) PodGroup() PodGroupController {
	return generic.NewNonNamespacedController[*v1beta1.PodGroup, *v1beta1.PodGroupList](schema.GroupVersionKind{Group: "scheduling.volcano.sh", Version: "v1beta1", Kind: "PodGroup"}, "podgroups", v.controllerFactory)
}

func (v *version) Queue() QueueController {
	return generic.NewNonNamespacedController[*v1beta1.Queue, *v1beta1.QueueList](schema.GroupVersionKind{Group: "scheduling.volcano.sh", Version: "v1beta1", Kind: "Queue"}, "queues", v.controllerFactory)
}
