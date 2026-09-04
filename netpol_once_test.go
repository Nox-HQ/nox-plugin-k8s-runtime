package main

import (
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func netpolPod(name, namespace string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec: corev1.PodSpec{
			AutomountServiceAccountToken: boolPtr(false),
			Containers: []corev1.Container{{
				Name: "app", Image: "nginx:1.25",
				SecurityContext: &corev1.SecurityContext{RunAsNonRoot: boolPtr(true)},
				Resources: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("100m"),
						corev1.ResourceMemory: resource.MustParse("128Mi"),
					},
				},
			}},
		},
	}
}

// "This namespace has no NetworkPolicy" is one fact about the namespace. It was
// evaluated inside the per-pod loop, so it was emitted once per pod: on a real
// cluster that turned 5 facts into 49 findings, 27 of them the identical
// sentence about one namespace.
//
// The existing tests could not see it — each used a single pod in a single
// namespace, where one-per-pod and one-per-namespace are the same number.
func TestKRUNT004IsReportedOncePerNamespaceNotOncePerPod(t *testing.T) {
	client := fake.NewSimpleClientset(
		netpolPod("web-1", "unguarded"),
		netpolPod("web-2", "unguarded"),
		netpolPod("web-3", "unguarded"),
		netpolPod("api-1", "also-unguarded"),
		netpolPod("safe-1", "guarded"),
		&networkingv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{Name: "deny-all", Namespace: "guarded"},
		},
	)

	findings, err := NewScanner(client).ScanCluster(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}

	var got []string
	for _, f := range findings {
		if f.RuleID == "KRUNT-004" {
			got = append(got, f.Namespace)
		}
	}

	if len(got) != 2 {
		t.Errorf("KRUNT-004 findings = %d %v, want 2 (one per unguarded namespace, not one per pod)", len(got), got)
	}
	if len(got) == 2 && (got[0] != "also-unguarded" || got[1] != "unguarded") {
		t.Errorf("namespaces = %v, want [also-unguarded unguarded] in sorted order", got)
	}
}

// The finding used to be anchored to an arbitrary pod in the namespace, so its
// fingerprint changed every time that pod was replaced — breaking any baseline
// entry or waiver written against it. A namespace-level fact belongs at a
// namespace-level path.
func TestKRUNT004PathIsStableAcrossPodChurn(t *testing.T) {
	before := scanNetpolPaths(t, netpolPod("web-abc123", "unguarded"))
	after := scanNetpolPaths(t, netpolPod("web-xyz789", "unguarded"))

	if len(before) != 1 || len(after) != 1 {
		t.Fatalf("expected one KRUNT-004 finding each, got %v and %v", before, after)
	}
	if before[0] != after[0] {
		t.Errorf("path changed when the pod was replaced: %q then %q", before[0], after[0])
	}
	if strings.Contains(before[0], "/Pod/") {
		t.Errorf("namespace-level finding is anchored to a pod: %q", before[0])
	}
}

func scanNetpolPaths(t *testing.T, pod *corev1.Pod) []string {
	t.Helper()
	client := fake.NewSimpleClientset(pod)
	findings, err := NewScanner(client).ScanCluster(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	// Indexed rather than ranged by value: Finding is 128 bytes and gocritic's
	// rangeValCopy is enforced on new code here.
	for i := range findings {
		if findings[i].RuleID == "KRUNT-004" {
			paths = append(paths, findings[i].Path)
		}
	}
	return paths
}
