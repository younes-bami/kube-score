package security

import (
	ks "github.com/younes-bami/kube-score/domain"
	"github.com/younes-bami/kube-score/score/checks"
	"github.com/younes-bami/kube-score/scorecard"
	corev1 "k8s.io/api/core/v1"
)

func Register(allChecks *checks.Checks) {
	allChecks.RegisterPodCheck("Container Security Context User Group ID", `Makes sure that all pods have a security context with valid UID and GID set `, containerSecurityContextUserGroupID)
	allChecks.RegisterPodCheck("Container Security Context Privileged", "Makes sure that all pods have a unprivileged security context set", containerSecurityContextPrivileged)
	allChecks.RegisterPodCheck("Container Security Context ReadOnlyRootFilesystem", "Makes sure that all pods have a security context with read only filesystem set", containerSecurityContextReadOnlyRootFilesystem)

	allChecks.RegisterOptionalPodCheck("Container Seccomp Profile", `Makes sure that all pods have at a seccomp policy configured.`, podSeccompProfile)
}

// containerSecurityContextReadOnlyRootFilesystem checks for pods using writeable root filesystems
func containerSecurityContextReadOnlyRootFilesystem(ps ks.PodSpecer) (score scorecard.TestScore, err error) {
	allContainers := ps.GetPodTemplateSpec().Spec.InitContainers
	allContainers = append(allContainers, ps.GetPodTemplateSpec().Spec.Containers...)

	noContextSet := false
	hasWritableRootFS := false

	for _, container := range allContainers {
		if container.SecurityContext == nil {
			noContextSet = true
			score.AddComment(container.Name, "Container has no configured security context", "Set securityContext to run the container in a more secure context.")
			continue
		}
		sec := container.SecurityContext
		if sec.ReadOnlyRootFilesystem == nil || !*sec.ReadOnlyRootFilesystem {
			hasWritableRootFS = true
			score.AddComment(container.Name, "The pod has a container with a writable root filesystem", "Set securityContext.readOnlyRootFilesystem to true")
		}
	}

	if noContextSet || hasWritableRootFS {
		score.Grade = scorecard.GradeCritical
	} else {
		score.Grade = scorecard.GradeAllOK
	}

	return
}

// containerSecurityContextPrivileged checks for privileged containers
func containerSecurityContextPrivileged(ps ks.PodSpecer) (score scorecard.TestScore, err error) {
	allContainers := ps.GetPodTemplateSpec().Spec.InitContainers
	allContainers = append(allContainers, ps.GetPodTemplateSpec().Spec.Containers...)
	hasPrivileged := false
	for _, container := range allContainers {
		if container.SecurityContext != nil && container.SecurityContext.Privileged != nil && *container.SecurityContext.Privileged {
			hasPrivileged = true
			score.AddComment(container.Name, "The container is privileged", "Set securityContext.privileged to false. Privileged containers can access all devices on the host, and grants almost the same access as non-containerized processes on the host.")
		}
	}
	if hasPrivileged {
		score.Grade = scorecard.GradeCritical
	} else {
		score.Grade = scorecard.GradeAllOK
	}
	return
}

// containerSecurityContextUserGroupID checks that the user and group are valid ( > 10000) in the security context
func containerSecurityContextUserGroupID(ps ks.PodSpecer) (score scorecard.TestScore, err error) {
	allContainers := ps.GetPodTemplateSpec().Spec.InitContainers
	allContainers = append(allContainers, ps.GetPodTemplateSpec().Spec.Containers...)
	podSecurityContext := ps.GetPodTemplateSpec().Spec.SecurityContext
	noContextSet := false
	hasLowUserID := false
	hasLowGroupID := false
	for _, container := range allContainers {
		if container.SecurityContext == nil && podSecurityContext == nil {
			noContextSet = true
			score.AddComment(container.Name, "Container has no configured security context", "Set securityContext to run the container in a more secure context.")
			continue
		}
		sec := container.SecurityContext
		if sec == nil {
			sec = &corev1.SecurityContext{}
		}
		// Forward values from PodSecurityContext to the (container level) SecurityContext if not set
		if podSecurityContext != nil {
			if sec.RunAsGroup == nil {
				sec.RunAsGroup = podSecurityContext.RunAsGroup
			}
			if sec.RunAsUser == nil {
				sec.RunAsUser = podSecurityContext.RunAsUser
			}
		}
		if sec.RunAsUser == nil || *sec.RunAsUser < 10000 {
			hasLowUserID = true
			score.AddComment(container.Name, "The container is running with a low user ID", "A userid above 10 000 is recommended to avoid conflicts with the host. Set securityContext.runAsUser to a value > 10000")
		}

		if sec.RunAsGroup == nil || *sec.RunAsGroup < 10000 {
			hasLowGroupID = true
			score.AddComment(container.Name, "The container running with a low group ID", "A groupid above 10 000 is recommended to avoid conflicts with the host. Set securityContext.runAsGroup to a value > 10000")
		}
	}
	if noContextSet || hasLowUserID || hasLowGroupID {
		score.Grade = scorecard.GradeCritical
	} else {
		score.Grade = scorecard.GradeAllOK
	}
	return
}

// podSeccompProfile checks that a Seccommp profile is configured for the pod
func podSeccompProfile(ps ks.PodSpecer) (score scorecard.TestScore, err error) {
	metadata := ps.GetPodTemplateSpec().ObjectMeta

	seccompAnnotated := false
	if metadata.Annotations != nil {
		if _, ok := metadata.Annotations["seccomp.security.alpha.kubernetes.io/defaultProfileName"]; ok {
			seccompAnnotated = true
		}
	}

	if !seccompAnnotated {
		score.Grade = scorecard.GradeWarning
		score.AddComment(metadata.Name, "The pod has not configured Seccomp for its containers", "Running containers with Seccomp is recommended to reduce the kernel attack surface")
	} else {
		score.Grade = scorecard.GradeAllOK
	}

	return
}
