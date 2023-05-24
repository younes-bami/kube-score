package cronjob

import (
	ks "github.com/younes-bami/kube-score/domain"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CronJobV1 struct {
	Obj      v1.CronJob
	Location ks.FileLocation
}

func (c CronJobV1) StartingDeadlineSeconds() *int64 {
	return c.Obj.Spec.StartingDeadlineSeconds
}

func (c CronJobV1) BackoffLimit() *int32 {
	return c.Obj.Spec.JobTemplate.Spec.BackoffLimit
}

func (c CronJobV1) FileLocation() ks.FileLocation {
	return c.Location
}

func (c CronJobV1) GetTypeMeta() metav1.TypeMeta {
	return c.Obj.TypeMeta
}

func (c CronJobV1) GetObjectMeta() metav1.ObjectMeta {
	return c.Obj.ObjectMeta
}

func (c CronJobV1) GetPodTemplateSpec() corev1.PodTemplateSpec {
	t := c.Obj.Spec.JobTemplate.Spec.Template
	t.ObjectMeta.Namespace = c.Obj.ObjectMeta.Namespace
	return t
}
