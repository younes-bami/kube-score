package networkpolicy

import (
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ks "github.com/younes-bami/kube-score/domain"
	"github.com/younes-bami/kube-score/score/checks"
	"github.com/younes-bami/kube-score/score/internal"
	"github.com/younes-bami/kube-score/scorecard"
)

func Register(allChecks *checks.Checks, netpols ks.NetworkPolicies, pods ks.Pods, podspecers ks.PodSpeccers) {
	allChecks.RegisterPodCheck("Pod NetworkPolicy", `Makes sure that all Pods are targeted by a NetworkPolicy`, podHasNetworkPolicy(netpols.NetworkPolicies()))
	allChecks.RegisterNetworkPolicyCheck("NetworkPolicy targets Pod", `Makes sure that all NetworkPolicies targets at least one Pod`, networkPolicyTargetsPod(pods.Pods(), podspecers.PodSpeccers()))
}

// podHasNetworkPolicy returns a function that tests that all pods have matching NetworkPolicies
// podHasNetworkPolicy takes a list of all defined NetworkPolicies as input
func podHasNetworkPolicy(allNetpols []ks.NetworkPolicy) func(ks.PodSpecer) (scorecard.TestScore, error) {
	return func(ps ks.PodSpecer) (score scorecard.TestScore, err error) {
		hasMatchingEgressNetpol := false
		hasMatchingIngressNetpol := false

		for _, n := range allNetpols {
			netPol := n.NetworkPolicy()

			// Make sure that the pod and networkpolicy is in the same namespace
			if ps.GetPodTemplateSpec().Namespace != netPol.Namespace {
				continue
			}

			if selector, err := metav1.LabelSelectorAsSelector(&netPol.Spec.PodSelector); err == nil {
				if selector.Matches(internal.MapLabels(ps.GetPodTemplateSpec().Labels)) {

					// Documentation of PolicyTypes
					//
					// List of rule types that the NetworkPolicy relates to.
					// Valid options are "Ingress", "Egress", or "Ingress,Egress".
					// If this field is not specified, it will default based on the existence of Ingress or Egress rules;
					// policies that contain an Egress section are assumed to affect Egress, and all policies
					// (whether or not they contain an Ingress section) are assumed to affect Ingress.
					// If you want to write an egress-only policy, you must explicitly specify policyTypes [ "Egress" ].
					// Likewise, if you want to write a policy that specifies that no egress is allowed,
					// you must specify a policyTypes value that include "Egress" (since such a policy would not include
					// an Egress section and would otherwise default to just [ "Ingress" ]).

					if netPol.Spec.PolicyTypes == nil || len(netPol.Spec.PolicyTypes) == 0 {
						hasMatchingIngressNetpol = true
						if len(netPol.Spec.Egress) > 0 {
							hasMatchingEgressNetpol = true
						}
					} else {
						for _, policyType := range netPol.Spec.PolicyTypes {
							if policyType == networkingv1.PolicyTypeIngress {
								hasMatchingIngressNetpol = true
							}
							if policyType == networkingv1.PolicyTypeEgress {
								hasMatchingEgressNetpol = true
							}
						}
					}
				}
			}
		}

		switch {
		case hasMatchingEgressNetpol && hasMatchingIngressNetpol:
			score.Grade = scorecard.GradeAllOK
		case hasMatchingEgressNetpol && !hasMatchingIngressNetpol:
			score.Grade = scorecard.GradeWarning
			score.AddComment("", "The pod does not have a matching ingress NetworkPolicy", "Add a ingress policy to the pods NetworkPolicy")
		case hasMatchingIngressNetpol && !hasMatchingEgressNetpol:
			score.Grade = scorecard.GradeWarning
			score.AddComment("", "The pod does not have a matching egress NetworkPolicy", "Add a egress policy to the pods NetworkPolicy")
		default:
			score.Grade = scorecard.GradeCritical
			score.AddComment("", "The pod does not have a matching NetworkPolicy", "Create a NetworkPolicy that targets this pod to control who/what can communicate with this pod. Note, this feature needs to be supported by the CNI implementation used in the Kubernetes cluster to have an effect.")
		}

		return
	}
}

func networkPolicyTargetsPod(pods []ks.Pod, podspecers []ks.PodSpecer) func(networkingv1.NetworkPolicy) (scorecard.TestScore, error) {
	return func(netpol networkingv1.NetworkPolicy) (score scorecard.TestScore, err error) {
		hasMatch := false

		for _, p := range pods {
			pod := p.Pod()
			if pod.Namespace != netpol.Namespace {
				continue
			}

			if selector, err := metav1.LabelSelectorAsSelector(&netpol.Spec.PodSelector); err == nil {
				if selector.Matches(internal.MapLabels(pod.Labels)) {
					hasMatch = true
					break
				}
			}
		}

		if !hasMatch {
			for _, pod := range podspecers {
				if pod.GetObjectMeta().Namespace != netpol.Namespace {
					continue
				}

				if selector, err := metav1.LabelSelectorAsSelector(&netpol.Spec.PodSelector); err == nil {
					if selector.Matches(internal.MapLabels(pod.GetPodTemplateSpec().Labels)) {
						hasMatch = true
						break
					}
				}
			}
		}

		if hasMatch {
			score.Grade = scorecard.GradeAllOK
		} else {
			score.Grade = scorecard.GradeCritical
			score.AddComment("", "The NetworkPolicys selector doesn't match any pods", "")
		}

		return
	}
}
