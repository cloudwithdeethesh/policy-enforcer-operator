/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	policyv1alpha1 "github.com/cloudwithdeethesh/policy-enforcer-operator/api/v1alpha1"
)

// PolicyRuleReconciler reconciles a PolicyRule object
type PolicyRuleReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=policy.cloudwithdeethesh.com,resources=policyrules,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=policy.cloudwithdeethesh.com,resources=policyrules/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=policy.cloudwithdeethesh.com,resources=policyrules/finalizers,verbs=update
// +kubebuilder:rbac:groups=policy.cloudwithdeethesh.com,resources=policyenforcements,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *PolicyRuleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the PolicyRule instance
	var policyRule policyv1alpha1.PolicyRule
	if err := r.Get(ctx, req.NamespacedName, &policyRule); err != nil {
		if errors.IsNotFound(err) {
			// PolicyRule was deleted, cleanup any enforcement records
			logger.Info("PolicyRule not found, cleaning up", "name", req.Name)
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get PolicyRule")
		return ctrl.Result{}, err
	}

	logger.Info("Reconciling PolicyRule", "name", policyRule.Name, "namespace", policyRule.Namespace)

	// Get all pods that match this policy rule
	matchingPods, err := r.getMatchingPods(ctx, &policyRule)
	if err != nil {
		logger.Error(err, "Failed to get matching pods")
		return ctrl.Result{}, err
	}

	logger.Info("Found matching pods", "count", len(matchingPods))

	// Update status with enforcement information
	now := metav1.Now()
	policyRule.Status.LastApplied = &now
	policyRule.Status.ViolationCount = 0

	// Count violations for reporting
	for _, pod := range matchingPods {
		violations := r.detectViolations(&pod, &policyRule)
		policyRule.Status.ViolationCount += int32(len(violations))
	}

	// Update the status
	if err := r.Status().Update(ctx, &policyRule); err != nil {
		logger.Error(err, "Failed to update PolicyRule status")
		return ctrl.Result{}, err
	}

	// Requeue after 5 minutes to check for new violations
	return ctrl.Result{RequeueAfter: time.Minute * 5}, nil
}

func (r *PolicyRuleReconciler) getMatchingPods(ctx context.Context, policyRule *policyv1alpha1.PolicyRule) ([]corev1.Pod, error) {
	var allPods corev1.PodList
	var matchingPods []corev1.Pod

	// List pods in specified namespaces or all namespaces if none specified
	if len(policyRule.Spec.Namespaces) > 0 {
		for _, ns := range policyRule.Spec.Namespaces {
			var pods corev1.PodList
			if err := r.List(ctx, &pods, client.InNamespace(ns)); err != nil {
				return nil, err
			}
			allPods.Items = append(allPods.Items, pods.Items...)
		}
	} else {
		if err := r.List(ctx, &allPods); err != nil {
			return nil, err
		}
	}

	// Filter pods based on label selector
	for _, pod := range allPods.Items {
		if r.isPodMatching(&pod, policyRule) {
			matchingPods = append(matchingPods, pod)
		}
	}

	return matchingPods, nil
}

func (r *PolicyRuleReconciler) isPodMatching(pod *corev1.Pod, policyRule *policyv1alpha1.PolicyRule) bool {
	// Check label selector
	if policyRule.Spec.Selector.MatchLabels != nil || len(policyRule.Spec.Selector.MatchExpressions) > 0 {
		selector := labels.Set(policyRule.Spec.Selector.MatchLabels).AsSelector()
		if !selector.Matches(labels.Set(pod.Labels)) {
			return false
		}
	}

	return true
}

func (r *PolicyRuleReconciler) detectViolations(pod *corev1.Pod, policyRule *policyv1alpha1.PolicyRule) []string {
	var violations []string

	// Check resource policies
	if policyRule.Spec.ResourcePolicies != nil {
		for i, container := range pod.Spec.Containers {
			if policyRule.Spec.ResourcePolicies.RequireResourceLimits && container.Resources.Limits == nil {
				violations = append(violations, fmt.Sprintf("container %d (%s) missing resource limits", i, container.Name))
			}
			if policyRule.Spec.ResourcePolicies.RequireResourceRequests && container.Resources.Requests == nil {
				violations = append(violations, fmt.Sprintf("container %d (%s) missing resource requests", i, container.Name))
			}
		}
	}

	// Check probe policies
	if policyRule.Spec.ProbePolicies != nil {
		for i, container := range pod.Spec.Containers {
			if policyRule.Spec.ProbePolicies.RequireLivenessProbe && container.LivenessProbe == nil {
				violations = append(violations, fmt.Sprintf("container %d (%s) missing liveness probe", i, container.Name))
			}
			if policyRule.Spec.ProbePolicies.RequireReadinessProbe && container.ReadinessProbe == nil {
				violations = append(violations, fmt.Sprintf("container %d (%s) missing readiness probe", i, container.Name))
			}
		}
	}

	// Check security policies
	if policyRule.Spec.SecurityPolicies != nil {
		for i, container := range pod.Spec.Containers {
			if container.SecurityContext != nil {
				if policyRule.Spec.SecurityPolicies.RequireNonRootUser && (container.SecurityContext.RunAsNonRoot == nil || !*container.SecurityContext.RunAsNonRoot) {
					violations = append(violations, fmt.Sprintf("container %d (%s) not running as non-root", i, container.Name))
				}
			} else if policyRule.Spec.SecurityPolicies.RequireNonRootUser {
				violations = append(violations, fmt.Sprintf("container %d (%s) missing security context", i, container.Name))
			}
		}
	}

	return violations
}

// SetupWithManager sets up the controller with the Manager.
func (r *PolicyRuleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&policyv1alpha1.PolicyRule{}).
		Named("policyrule").
		Complete(r)
}
