/*
Copyright 2022.

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

package v1

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/togettoyou/sidecar-go/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var podlog = logf.Log.WithName("pod-resource")

//+kubebuilder:webhook:path=/mutate-core-v1-pod,mutating=true,failurePolicy=fail,sideEffects=None,groups=core,resources=pods,verbs=create;update,versions=v1,name=mpod.kb.io,admissionReviewVersions=v1

type podMutate struct {
	Client  client.Client
	decoder *admission.Decoder
}

func NewPodMutate(c client.Client) admission.Handler {
	return &podMutate{Client: c}
}

func (pm *podMutate) Handle(ctx context.Context, req admission.Request) admission.Response {
	podlog.Info("pod webhook")
	pod := &corev1.Pod{}

	err := pm.decoder.Decode(req, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	specs := util.PodMatchedSidecarGo(pod)
	initContainers := make([]corev1.Container, 0)
	containers := make([]corev1.Container, 0)
	volumes := make([]corev1.Volume, 0)
	for _, spec := range specs {
		initContainers = append(initContainers, spec.InitContainers...)
		containers = append(containers, spec.Containers...)
		volumes = append(volumes, spec.Volumes...)
	}
	// 1.inject init containers
	pod.Spec.InitContainers = append(pod.Spec.InitContainers, initContainers...)
	// 2.inject containers
	pod.Spec.Containers = util.MergeContainers(pod.Spec.Containers, containers)
	// 3.inject volumes
	pod.Spec.Volumes = util.MergeVolumes(pod.Spec.Volumes, volumes)

	marshaledPod, err := json.Marshal(pod)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}

// InjectDecoder injects the decoder.
func (pm *podMutate) InjectDecoder(d *admission.Decoder) error {
	pm.decoder = d
	return nil
}
