package apps

import (
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ks "github.com/younes-bami/kube-score/domain"
	"github.com/younes-bami/kube-score/score/checks"
	"github.com/younes-bami/kube-score/score/internal"
	"github.com/younes-bami/kube-score/scorecard"
)

func Register(allChecks *checks.Checks, allHPAs []ks.HpaTargeter, allServices []ks.Service) {
	allChecks.RegisterDeploymentCheck("Deployment has host PodAntiAffinity", "Makes sure that a podAntiAffinity has been set that prevents multiple pods from being scheduled on the same node. https://kubernetes.io/docs/concepts/configuration/assign-pod-node/", deploymentHasAntiAffinity)
	allChecks.RegisterStatefulSetCheck("StatefulSet has host PodAntiAffinity", "Makes sure that a podAntiAffinity has been set that prevents multiple pods from being scheduled on the same node. https://kubernetes.io/docs/concepts/configuration/assign-pod-node/", statefulsetHasAntiAffinity)

	allChecks.RegisterDeploymentCheck("Deployment targeted by HPA does not have replicas configured", "Makes sure that Deployments using a HorizontalPodAutoscaler doesn't have a statically configured replica count set", hpaDeploymentNoReplicas(allHPAs))
	allChecks.RegisterStatefulSetCheck("StatefulSet has ServiceName", "Makes sure that StatefulSets have an existing headless serviceName.", statefulsetHasServiceName(allServices))

	allChecks.RegisterDeploymentCheck("Deployment Pod Selector labels match template metadata labels", "Ensure the StatefulSet selector labels match the template metadata labels.", deploymentSelectorLabelsMatching)
	allChecks.RegisterStatefulSetCheck("StatefulSet Pod Selector labels match template metadata labels", "Ensure the StatefulSet selector labels match the template metadata labels.", statefulSetSelectorLabelsMatching)
}

func hpaDeploymentNoReplicas(allHPAs []ks.HpaTargeter) func(deployment appsv1.Deployment) (scorecard.TestScore, error) {
	return func(deployment appsv1.Deployment) (score scorecard.TestScore, err error) {
		// If is targeted by a HPA
		for _, hpa := range allHPAs {
			target := hpa.HpaTarget()

			if hpa.GetObjectMeta().Namespace == deployment.Namespace &&
				strings.EqualFold(target.Kind, deployment.Kind) &&
				target.Name == deployment.Name {

				if deployment.Spec.Replicas == nil {
					score.Grade = scorecard.GradeAllOK
					return
				}

				score.Grade = scorecard.GradeCritical
				score.AddComment("", "The deployment is targeted by a HPA, but a static replica count is configured in the DeploymentSpec", "When replicas are both statically set and managed by the HPA, the replicas will be changed to the statically configured count when the spec is applied, even if the HPA wants the replica count to be higher.")
				return
			}
		}

		score.Grade = scorecard.GradeAllOK
		score.Skipped = true
		score.AddComment("", "Skipped because the deployment is not targeted by a HorizontalPodAutoscaler", "")
		return
	}
}

func deploymentHasAntiAffinity(deployment appsv1.Deployment) (score scorecard.TestScore, err error) {
	// Ignore if the deployment only has a single replica
	// If replicas is not explicitly set, we'll still warn if the anti affinity is missing
	// as that might indicate use of a Horizontal Pod Autoscaler
	if deployment.Spec.Replicas != nil && *deployment.Spec.Replicas < 2 {
		score.Skipped = true
		score.AddComment("", "Skipped because the deployment has less than 2 replicas", "")
		return
	}

	warn := func() {
		score.Grade = scorecard.GradeWarning
		score.AddComment("", "Deployment does not have a host podAntiAffinity set", "It's recommended to set a podAntiAffinity that stops multiple pods from a deployment from being scheduled on the same node. This increases availability in case the node becomes unavailable.")
	}

	affinity := deployment.Spec.Template.Spec.Affinity
	if affinity == nil || affinity.PodAntiAffinity == nil {
		warn()
		return
	}

	labels := internal.MapLabels(deployment.Spec.Template.GetObjectMeta().GetLabels())

	if hasPodAntiAffinity(labels, affinity) {
		score.Grade = scorecard.GradeAllOK
		return
	}

	warn()
	return
}

func statefulsetHasAntiAffinity(statefulset appsv1.StatefulSet) (score scorecard.TestScore, err error) {
	// Ignore if the statefulset only has a single replica
	// If replicas is not explicitly set, we'll still warn if the anti affinity is missing
	// as that might indicate use of a Horizontal Pod Autoscaler
	if statefulset.Spec.Replicas != nil && *statefulset.Spec.Replicas < 2 {
		score.Skipped = true
		score.AddComment("", "Skipped because the statefulset has less than 2 replicas", "")
		return
	}

	warn := func() {
		score.Grade = scorecard.GradeWarning
		score.AddComment("", "StatefulSet does not have a host podAntiAffinity set", "It's recommended to set a podAntiAffinity that stops multiple pods from a statefulset from being scheduled on the same node. This increases availability in case the node becomes unavailable.")
	}

	affinity := statefulset.Spec.Template.Spec.Affinity
	if affinity == nil || affinity.PodAntiAffinity == nil {
		warn()
		return
	}

	labels := internal.MapLabels(statefulset.Spec.Template.GetObjectMeta().GetLabels())

	if hasPodAntiAffinity(labels, affinity) {
		score.Grade = scorecard.GradeAllOK
		return
	}

	warn()
	return
}

func hasPodAntiAffinity(selfLabels internal.MapLabels, affinity *corev1.Affinity) bool {
	approvedTopologyKeys := map[string]struct{}{
		"kubernetes.io/hostname":        {},
		"topology.kubernetes.io/region": {},
		"topology.kubernetes.io/zone":   {},

		// Deprecated in Kubernetes v1.17
		"failure-domain.beta.kubernetes.io/region": {},
		"failure-domain.beta.kubernetes.io/zone":   {},
	}

	for _, pref := range affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
		if _, ok := approvedTopologyKeys[pref.PodAffinityTerm.TopologyKey]; ok {
			if selector, err := metav1.LabelSelectorAsSelector(pref.PodAffinityTerm.LabelSelector); err == nil {
				if selector.Matches(selfLabels) {
					return true
				}
			}
		}
	}

	for _, req := range affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution {
		if _, ok := approvedTopologyKeys[req.TopologyKey]; ok {
			if selector, err := metav1.LabelSelectorAsSelector(req.LabelSelector); err == nil {
				if selector.Matches(selfLabels) {
					return true
				}
			}
		}
	}

	return false
}

func statefulsetHasServiceName(allServices []ks.Service) func(statefulset appsv1.StatefulSet) (scorecard.TestScore, error) {
	return func(statefulset appsv1.StatefulSet) (score scorecard.TestScore, err error) {
		for _, service := range allServices {
			if service.Service().Namespace != statefulset.Namespace ||
				service.Service().Name != statefulset.Spec.ServiceName ||
				service.Service().Spec.ClusterIP != "None" {
				continue
			}

			if internal.LabelSelectorMatchesLabels(
				service.Service().Spec.Selector,
				statefulset.Spec.Template.GetObjectMeta().GetLabels(),
			) {
				score.Grade = scorecard.GradeAllOK
				return
			}
		}

		score.Grade = scorecard.GradeCritical
		score.AddComment("", "StatefulSet does not have a valid serviceName", "StatefulSets currently require a Headless Service to be responsible for the network identity of the Pods. You are responsible for creating this Service. https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#limitations")
		return
	}
}

func statefulSetSelectorLabelsMatching(statefulset appsv1.StatefulSet) (score scorecard.TestScore, err error) {
	selector, err := metav1.LabelSelectorAsSelector(statefulset.Spec.Selector)
	if err != nil {
		score.Grade = scorecard.GradeCritical
		score.AddComment("", "StatefulSet selector labels are not matching template metadata labels", fmt.Sprintf("Invalid selector: %s", err))
		return
	}

	if selector.Matches(internal.MapLabels(statefulset.Spec.Template.GetObjectMeta().GetLabels())) {
		score.Grade = scorecard.GradeAllOK
		return
	}

	score.Grade = scorecard.GradeCritical
	score.AddComment("", "StatefulSet selector labels not matching template metadata labels", "StatefulSets require `.spec.selector` to match `.spec.template.metadata.labels`. https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#pod-selector")
	return
}

func deploymentSelectorLabelsMatching(deployment appsv1.Deployment) (score scorecard.TestScore, err error) {
	selector, err := metav1.LabelSelectorAsSelector(deployment.Spec.Selector)
	if err != nil {
		score.Grade = scorecard.GradeCritical
		score.AddComment("", "Deployment selector labels are not matching template metadata labels", fmt.Sprintf("Invalid selector: %s", err))
		return
	}

	if selector.Matches(internal.MapLabels(deployment.Spec.Template.GetObjectMeta().GetLabels())) {
		score.Grade = scorecard.GradeAllOK
		return
	}

	score.Grade = scorecard.GradeCritical
	score.AddComment("", "Deployment selector labels not matching template metadata labels", "Deployment require `.spec.selector` to match `.spec.template.metadata.labels`. https://kubernetes.io/docs/concepts/workloads/controllers/deployment/")
	return
}
