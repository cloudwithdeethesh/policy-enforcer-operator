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

package v1

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	policyv1alpha1 "github.com/cloudwithdeethesh/policy-enforcer-operator/api/v1alpha1"
)

// nolint:unused
// log is for logging in this package.
var podlog = logf.Log.WithName("pod-resource")

// SetupPodWebhookWithManager registers the webhook for Pod in the manager.
func SetupPodWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&corev1.Pod{}).
		WithValidator(&PodCustomValidator{Client: mgr.GetClient()}).
		WithDefaulter(&PodCustomDefaulter{Client: mgr.GetClient()}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate--v1-pod,mutating=true,failurePolicy=fail,sideEffects=None,groups="",resources=pods,verbs=create;update,versions=v1,name=mpod-v1.kb.io,admissionReviewVersions=v1

// PodCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind Pod when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type PodCustomDefaulter struct {
	Client client.Client
}

var _ webhook.CustomDefaulter = &PodCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind Pod.
func (d *PodCustomDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return fmt.Errorf("expected an Pod object but got %T", obj)
	}
	podlog.Info("Defaulting for Pod", "name", pod.GetName())

	// Get applicable policy rules
	policyRules, err := d.getApplicablePolicyRules(ctx, pod)
	if err != nil {
		podlog.Error(err, "Failed to get applicable policy rules")
		return err
	}

	// Apply defaults from policy rules
	for _, rule := range policyRules {
		if rule.Spec.EnforcementAction != "enforce" {
			continue // Only apply defaults when enforcing
		}

		d.applyResourceDefaults(pod, &rule)
		d.applyProbeDefaults(pod, &rule)
		d.applySecurityDefaults(pod, &rule)
	}

	return nil
}

func (d *PodCustomDefaulter) getApplicablePolicyRules(ctx context.Context, pod *corev1.Pod) ([]policyv1alpha1.PolicyRule, error) {
	var policyRules policyv1alpha1.PolicyRuleList
	if err := d.Client.List(ctx, &policyRules); err != nil {
		return nil, err
	}

	var applicableRules []policyv1alpha1.PolicyRule
	for _, rule := range policyRules.Items {
		if d.isPolicyApplicable(&rule, pod) {
			applicableRules = append(applicableRules, rule)
		}
	}

	return applicableRules, nil
}

func (d *PodCustomDefaulter) isPolicyApplicable(rule *policyv1alpha1.PolicyRule, pod *corev1.Pod) bool {
	// Check namespace filter
	if len(rule.Spec.Namespaces) > 0 {
		found := false
		for _, ns := range rule.Spec.Namespaces {
			if ns == pod.Namespace {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check label selector
	if rule.Spec.Selector.MatchLabels != nil || len(rule.Spec.Selector.MatchExpressions) > 0 {
		selector := labels.Set(rule.Spec.Selector.MatchLabels).AsSelector()
		if !selector.Matches(labels.Set(pod.Labels)) {
			return false
		}
	}

	return true
}

func (d *PodCustomDefaulter) applyResourceDefaults(pod *corev1.Pod, rule *policyv1alpha1.PolicyRule) {
	if rule.Spec.ResourcePolicies == nil {
		return
	}

	for i := range pod.Spec.Containers {
		container := &pod.Spec.Containers[i]

		// Apply default resource limits
		if rule.Spec.ResourcePolicies.DefaultResourceLimits != nil {
			if container.Resources.Limits == nil {
				container.Resources.Limits = make(corev1.ResourceList)
			}
			for k, v := range rule.Spec.ResourcePolicies.DefaultResourceLimits {
				if _, exists := container.Resources.Limits[k]; !exists {
					container.Resources.Limits[k] = v
				}
			}
		}

		// Apply default resource requests
		if rule.Spec.ResourcePolicies.DefaultResourceRequests != nil {
			if container.Resources.Requests == nil {
				container.Resources.Requests = make(corev1.ResourceList)
			}
			for k, v := range rule.Spec.ResourcePolicies.DefaultResourceRequests {
				if _, exists := container.Resources.Requests[k]; !exists {
					container.Resources.Requests[k] = v
				}
			}
		}
	}
}

func (d *PodCustomDefaulter) applyProbeDefaults(pod *corev1.Pod, rule *policyv1alpha1.PolicyRule) {
	if rule.Spec.ProbePolicies == nil {
		return
	}

	for i := range pod.Spec.Containers {
		container := &pod.Spec.Containers[i]

		// Apply default liveness probe
		if rule.Spec.ProbePolicies.DefaultLivenessProbe != nil && container.LivenessProbe == nil {
			container.LivenessProbe = rule.Spec.ProbePolicies.DefaultLivenessProbe.DeepCopy()
		}

		// Apply default readiness probe
		if rule.Spec.ProbePolicies.DefaultReadinessProbe != nil && container.ReadinessProbe == nil {
			container.ReadinessProbe = rule.Spec.ProbePolicies.DefaultReadinessProbe.DeepCopy()
		}
	}
}

func (d *PodCustomDefaulter) applySecurityDefaults(pod *corev1.Pod, rule *policyv1alpha1.PolicyRule) {
	if rule.Spec.SecurityPolicies == nil {
		return
	}

	for i := range pod.Spec.Containers {
		container := &pod.Spec.Containers[i]

		// Apply default security context
		if rule.Spec.SecurityPolicies.DefaultSecurityContext != nil && container.SecurityContext == nil {
			container.SecurityContext = rule.Spec.SecurityPolicies.DefaultSecurityContext.DeepCopy()
		}
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate--v1-pod,mutating=false,failurePolicy=fail,sideEffects=None,groups="",resources=pods,verbs=create;update,versions=v1,name=vpod-v1.kb.io,admissionReviewVersions=v1

// PodCustomValidator struct is responsible for validating the Pod resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type PodCustomValidator struct {
	Client client.Client
}

var _ webhook.CustomValidator = &PodCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type Pod.
func (v *PodCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return nil, fmt.Errorf("expected a Pod object but got %T", obj)
	}
	podlog.Info("Validation for Pod upon creation", "name", pod.GetName())

	return v.validatePod(ctx, pod)
}

func (v *PodCustomValidator) validatePod(ctx context.Context, pod *corev1.Pod) (admission.Warnings, error) {
	// Get applicable policy rules
	policyRules, err := v.getApplicablePolicyRules(ctx, pod)
	if err != nil {
		return nil, err
	}

	var warnings admission.Warnings
	var errors []string

	// Validate against each applicable policy rule
	for _, rule := range policyRules {
		ruleWarnings, ruleErrors := v.validateAgainstPolicyRule(pod, &rule)
		warnings = append(warnings, ruleWarnings...)

		if rule.Spec.EnforcementAction == "enforce" {
			errors = append(errors, ruleErrors...)
		} else if rule.Spec.EnforcementAction == "warn" {
			for _, err := range ruleErrors {
				warnings = append(warnings, fmt.Sprintf("Policy violation: %s", err))
			}
		}
		// For "audit", we just log but don't warn or block
	}

	if len(errors) > 0 {
		return warnings, fmt.Errorf("policy violations: %v", errors)
	}

	return warnings, nil
}

func (v *PodCustomValidator) getApplicablePolicyRules(ctx context.Context, pod *corev1.Pod) ([]policyv1alpha1.PolicyRule, error) {
	var policyRules policyv1alpha1.PolicyRuleList
	if err := v.Client.List(ctx, &policyRules); err != nil {
		return nil, err
	}

	var applicableRules []policyv1alpha1.PolicyRule
	for _, rule := range policyRules.Items {
		if v.isPolicyApplicable(&rule, pod) {
			applicableRules = append(applicableRules, rule)
		}
	}

	return applicableRules, nil
}

func (v *PodCustomValidator) isPolicyApplicable(rule *policyv1alpha1.PolicyRule, pod *corev1.Pod) bool {
	// Check namespace filter
	if len(rule.Spec.Namespaces) > 0 {
		found := false
		for _, ns := range rule.Spec.Namespaces {
			if ns == pod.Namespace {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check label selector
	if rule.Spec.Selector.MatchLabels != nil || len(rule.Spec.Selector.MatchExpressions) > 0 {
		selector := labels.Set(rule.Spec.Selector.MatchLabels).AsSelector()
		if !selector.Matches(labels.Set(pod.Labels)) {
			return false
		}
	}

	return true
}

func (v *PodCustomValidator) validateAgainstPolicyRule(pod *corev1.Pod, rule *policyv1alpha1.PolicyRule) ([]string, []string) {
	var warnings []string
	var errors []string

	// Validate resource policies
	if rule.Spec.ResourcePolicies != nil {
		resourceWarnings, resourceErrors := v.validateResourcePolicies(pod, rule.Spec.ResourcePolicies)
		warnings = append(warnings, resourceWarnings...)
		errors = append(errors, resourceErrors...)
	}

	// Validate probe policies
	if rule.Spec.ProbePolicies != nil {
		probeWarnings, probeErrors := v.validateProbePolicies(pod, rule.Spec.ProbePolicies)
		warnings = append(warnings, probeWarnings...)
		errors = append(errors, probeErrors...)
	}

	// Validate security policies
	if rule.Spec.SecurityPolicies != nil {
		securityWarnings, securityErrors := v.validateSecurityPolicies(pod, rule.Spec.SecurityPolicies)
		warnings = append(warnings, securityWarnings...)
		errors = append(errors, securityErrors...)
	}

	return warnings, errors
}

func (v *PodCustomValidator) validateResourcePolicies(pod *corev1.Pod, policy *policyv1alpha1.ResourcePolicies) ([]string, []string) {
	var warnings []string
	var errors []string

	for i, container := range pod.Spec.Containers {
		if policy.RequireResourceLimits && container.Resources.Limits == nil {
			errors = append(errors, fmt.Sprintf("container %d (%s) is missing resource limits", i, container.Name))
		}

		if policy.RequireResourceRequests && container.Resources.Requests == nil {
			errors = append(errors, fmt.Sprintf("container %d (%s) is missing resource requests", i, container.Name))
		}

		// Check against max limits
		if policy.MaxResourceLimits != nil && container.Resources.Limits != nil {
			for resource, maxValue := range policy.MaxResourceLimits {
				if value, exists := container.Resources.Limits[resource]; exists {
					if value.Cmp(maxValue) > 0 {
						errors = append(errors, fmt.Sprintf("container %d (%s) %s limit exceeds maximum allowed", i, container.Name, resource))
					}
				}
			}
		}
	}

	return warnings, errors
}

func (v *PodCustomValidator) validateProbePolicies(pod *corev1.Pod, policy *policyv1alpha1.ProbePolicies) ([]string, []string) {
	var warnings []string
	var errors []string

	for i, container := range pod.Spec.Containers {
		if policy.RequireLivenessProbe && container.LivenessProbe == nil {
			errors = append(errors, fmt.Sprintf("container %d (%s) is missing liveness probe", i, container.Name))
		}

		if policy.RequireReadinessProbe && container.ReadinessProbe == nil {
			errors = append(errors, fmt.Sprintf("container %d (%s) is missing readiness probe", i, container.Name))
		}
	}

	return warnings, errors
}

func (v *PodCustomValidator) validateSecurityPolicies(pod *corev1.Pod, policy *policyv1alpha1.SecurityPolicies) ([]string, []string) {
	var warnings []string
	var errors []string

	for i, container := range pod.Spec.Containers {
		if container.SecurityContext != nil {
			if policy.RequireNonRootUser && (container.SecurityContext.RunAsNonRoot == nil || !*container.SecurityContext.RunAsNonRoot) {
				errors = append(errors, fmt.Sprintf("container %d (%s) must run as non-root user", i, container.Name))
			}

			if policy.RequireReadOnlyRootFilesystem && (container.SecurityContext.ReadOnlyRootFilesystem == nil || !*container.SecurityContext.ReadOnlyRootFilesystem) {
				errors = append(errors, fmt.Sprintf("container %d (%s) must have read-only root filesystem", i, container.Name))
			}
		} else if policy.RequireNonRootUser || policy.RequireReadOnlyRootFilesystem {
			errors = append(errors, fmt.Sprintf("container %d (%s) is missing required security context", i, container.Name))
		}
	}

	return warnings, errors
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type Pod.
func (v *PodCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	pod, ok := newObj.(*corev1.Pod)
	if !ok {
		return nil, fmt.Errorf("expected a Pod object for the newObj but got %T", newObj)
	}
	podlog.Info("Validation for Pod upon update", "name", pod.GetName())

	return v.validatePod(ctx, pod)
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type Pod.
func (v *PodCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return nil, fmt.Errorf("expected a Pod object but got %T", obj)
	}
	podlog.Info("Validation for Pod upon deletion", "name", pod.GetName())

	// TODO(user): fill in your validation logic upon object deletion.

	return nil, nil
}
