package pod

import (
	policyv1 "k8s.io/api/policy/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ks "github.com/younes-bami/kube-score/domain"
)

type PodDisruptionBudgetV1beta1 struct {
	Obj      policyv1beta1.PodDisruptionBudget
	Location ks.FileLocation
}

func (p PodDisruptionBudgetV1beta1) GetObjectMeta() metav1.ObjectMeta {
	return p.Obj.ObjectMeta
}

func (p PodDisruptionBudgetV1beta1) GetTypeMeta() metav1.TypeMeta {
	return p.Obj.TypeMeta
}

func (p PodDisruptionBudgetV1beta1) PodDisruptionBudgetSelector() *metav1.LabelSelector {
	return p.Obj.Spec.Selector
}

func (p PodDisruptionBudgetV1beta1) Namespace() string {
	return p.Obj.Namespace
}

func (p PodDisruptionBudgetV1beta1) FileLocation() ks.FileLocation {
	return p.Location
}

func (p PodDisruptionBudgetV1beta1) Spec() policyv1.PodDisruptionBudgetSpec {
	var polType *policyv1.UnhealthyPodEvictionPolicyType
	if p.Obj.Spec.UnhealthyPodEvictionPolicy != nil {
		t := policyv1.UnhealthyPodEvictionPolicyType(*p.Obj.Spec.UnhealthyPodEvictionPolicy)
		polType = &t
	}

	return policyv1.PodDisruptionBudgetSpec{
		MinAvailable:               p.Obj.Spec.MinAvailable,
		Selector:                   p.Obj.Spec.Selector,
		MaxUnavailable:             p.Obj.Spec.MaxUnavailable,
		UnhealthyPodEvictionPolicy: polType,
	}
}

type PodDisruptionBudgetV1 struct {
	Obj      policyv1.PodDisruptionBudget
	Location ks.FileLocation
}

func (p PodDisruptionBudgetV1) GetObjectMeta() metav1.ObjectMeta {
	return p.Obj.ObjectMeta
}

func (p PodDisruptionBudgetV1) GetTypeMeta() metav1.TypeMeta {
	return p.Obj.TypeMeta
}

func (p PodDisruptionBudgetV1) PodDisruptionBudgetSelector() *metav1.LabelSelector {
	return p.Obj.Spec.Selector
}

func (p PodDisruptionBudgetV1) FileLocation() ks.FileLocation {
	return p.Location
}

func (p PodDisruptionBudgetV1) Namespace() string {
	return p.Obj.Namespace
}

func (p PodDisruptionBudgetV1) Spec() policyv1.PodDisruptionBudgetSpec {
	return p.Obj.Spec
}
