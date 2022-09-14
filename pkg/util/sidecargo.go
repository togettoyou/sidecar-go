package util

import (
	"github.com/togettoyou/sidecar-go/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sync"
)

var (
	sidecarGoSpecM     = make(map[string]*v1alpha1.SidecarGoSpec, 0)
	sidecarGoSelectorM = make(map[string]labels.Selector, 0)
	sidecarGoSpecMu    sync.RWMutex
)

func UpdateSidecarGoSpec(namespacedName string, spec *v1alpha1.SidecarGoSpec) error {
	sidecarGoSpecMu.Lock()
	defer sidecarGoSpecMu.Unlock()

	if namespacedName == "" {
		return nil
	}
	if spec == nil {
		delete(sidecarGoSpecM, namespacedName)
		delete(sidecarGoSelectorM, namespacedName)
		return nil
	}
	sidecarGoSpecM[namespacedName] = spec
	if spec.Selector != nil {
		selector, err := v1.LabelSelectorAsSelector(spec.Selector)
		if err != nil {
			return err
		}
		sidecarGoSelectorM[namespacedName] = selector
	}
	return nil
}

func PodMatchedSidecarGo(pod *corev1.Pod) []*v1alpha1.SidecarGoSpec {
	sidecarGoSpecMu.RLock()
	defer sidecarGoSpecMu.RUnlock()

	specs := make([]*v1alpha1.SidecarGoSpec, 0)

	for namespacedName, spec := range sidecarGoSpecM {
		if spec.Namespace != "" && spec.Namespace != pod.Namespace {
			continue
		}
		if selector, ok := sidecarGoSelectorM[namespacedName]; ok {
			if !selector.Empty() && selector.Matches(labels.Set(pod.Labels)) {
				specs = append(specs, spec)
			}
		} else {
			if spec.Namespace != "" && spec.Namespace == pod.Namespace {
				specs = append(specs, spec)
			}
		}
	}

	return specs
}
