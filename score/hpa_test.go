package score

import (
	"testing"

	"github.com/younes-bami/kube-score/scorecard"
)

func TestHorizontalPodAutoscalerTargetsDeployment(t *testing.T) {
	t.Parallel()
	testExpectedScore(t, "hpa-targets-deployment.yaml", "HorizontalPodAutoscaler has target", scorecard.GradeAllOK)
}

func TestHorizontalPodAutoscalerHasNoTarget(t *testing.T) {
	t.Parallel()
	testExpectedScore(t, "hpa-has-no-target.yaml", "HorizontalPodAutoscaler has target", scorecard.GradeCritical)
}
