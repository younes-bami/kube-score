package meta

import (
	"regexp"

	"github.com/younes-bami/kube-score/domain"
	"github.com/younes-bami/kube-score/score/checks"
	"github.com/younes-bami/kube-score/scorecard"
)

func Register(allChecks *checks.Checks) {
	allChecks.RegisterMetaCheck("Label values", "Validates label values", validateLabelValues)
}

func validateLabelValues(meta domain.BothMeta) (score scorecard.TestScore, err error) {
	score.Grade = scorecard.GradeAllOK
	r := regexp.MustCompile("^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?$")
	for key, value := range meta.ObjectMeta.Labels {
		if !r.MatchString(value) {
			score.Grade = scorecard.GradeCritical
			score.AddComment(key, "Invalid label value", "The label value is invalid, and will not be accepted by Kubernetes")
		}
	}
	return
}
