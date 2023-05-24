package stable

import (
	"fmt"

	"github.com/younes-bami/kube-score/config"
	"github.com/younes-bami/kube-score/domain"
	"github.com/younes-bami/kube-score/score/checks"
	"github.com/younes-bami/kube-score/scorecard"
)

func Register(kubernetesVersion config.Semver, allChecks *checks.Checks) {
	allChecks.RegisterMetaCheck("Stable version", `Checks if the object is using a deprecated apiVersion`, metaStableAvailable(kubernetesVersion))
}

// ScoreMetaStableAvailable checks if the supplied TypeMeta is an unstable object type, that has a stable(r) replacement
func metaStableAvailable(kubernetsVersion config.Semver) func(meta domain.BothMeta) (scorecard.TestScore, error) {
	return func(meta domain.BothMeta) (score scorecard.TestScore, err error) {
		type recommendedApi struct {
			newAPI         string
			availableSince config.Semver
		}

		withStable := map[string]map[string]recommendedApi{
			"extensions/v1beta1": {
				"Deployment":   recommendedApi{"apps/v1", config.Semver{Major: 1, Minor: 9}},
				"DaemonSet":    recommendedApi{"apps/v1", config.Semver{Major: 1, Minor: 9}},
				"Ingress":      recommendedApi{"networking.k8s.io/v1", config.Semver{Major: 1, Minor: 19}},
				"IngressClass": recommendedApi{"networking.k8s.io/v1", config.Semver{Major: 1, Minor: 19}},
			},
			"apps/v1beta1": {
				"Deployment":  recommendedApi{"apps/v1", config.Semver{Major: 1, Minor: 9}},
				"StatefulSet": recommendedApi{"apps/v1", config.Semver{Major: 1, Minor: 9}},
			},
			"apps/v1beta2": {
				"Deployment":  recommendedApi{"apps/v1", config.Semver{Major: 1, Minor: 9}},
				"StatefulSet": recommendedApi{"apps/v1", config.Semver{Major: 1, Minor: 9}},
				"DaemonSet":   recommendedApi{"apps/v1", config.Semver{Major: 1, Minor: 9}},
			},
			"batch/v1beta1": {
				"CronJob": recommendedApi{"batch/v1", config.Semver{Major: 1, Minor: 21}},
			},
			"policy/v1beta1": {
				"PodDisruptionBudget": recommendedApi{"policy/v1", config.Semver{Major: 1, Minor: 21}},
			},
			"networking.k8s.io/v1beta1": {
				"Ingress":      recommendedApi{"networking.k8s.io/v1", config.Semver{Major: 1, Minor: 19}},
				"IngressClass": recommendedApi{"networking.k8s.io/v1", config.Semver{Major: 1, Minor: 19}},
			},
		}

		score.Grade = scorecard.GradeAllOK

		if inVersion, ok := withStable[meta.TypeMeta.APIVersion]; ok {
			if recAPI, ok := inVersion[meta.TypeMeta.Kind]; ok {

				// The recommended replacement is not available in the version of Kubernetes
				// that the user is using
				if kubernetsVersion.LessThan(recAPI.availableSince) {
					return
				}

				score.Grade = scorecard.GradeWarning
				score.AddComment("",
					fmt.Sprintf("The apiVersion and kind %s/%s is deprecated", meta.TypeMeta.APIVersion, meta.TypeMeta.Kind),
					fmt.Sprintf("It's recommended to use %s instead which has been available since Kubernetes %s", recAPI.newAPI, recAPI.availableSince.String()),
				)
				return
			}
		}

		return
	}
}
