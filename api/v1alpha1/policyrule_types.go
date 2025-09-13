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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PolicyRuleSpec defines the desired state of PolicyRule.
type PolicyRuleSpec struct {
	// Target selector for which resources this policy applies to
	Selector metav1.LabelSelector `json:"selector,omitempty"`

	// Namespaces where this policy should be enforced. If empty, applies to all namespaces
	Namespaces []string `json:"namespaces,omitempty"`

	// ResourcePolicies defines resource limits and requests policies
	ResourcePolicies *ResourcePolicies `json:"resourcePolicies,omitempty"`

	// ProbePolicies defines liveness and readiness probe policies
	ProbePolicies *ProbePolicies `json:"probePolicies,omitempty"`

	// SecurityPolicies defines security context policies
	SecurityPolicies *SecurityPolicies `json:"securityPolicies,omitempty"`

	// Enforcement action: enforce, warn, or audit
	// +kubebuilder:validation:Enum=enforce;warn;audit
	// +kubebuilder:default=enforce
	EnforcementAction string `json:"enforcementAction,omitempty"`
}

// ResourcePolicies defines resource limit and request policies
type ResourcePolicies struct {
	// RequireResourceLimits enforces that containers have resource limits
	RequireResourceLimits bool `json:"requireResourceLimits,omitempty"`

	// RequireResourceRequests enforces that containers have resource requests
	RequireResourceRequests bool `json:"requireResourceRequests,omitempty"`

	// DefaultResourceLimits to apply if none are specified
	DefaultResourceLimits corev1.ResourceList `json:"defaultResourceLimits,omitempty"`

	// DefaultResourceRequests to apply if none are specified
	DefaultResourceRequests corev1.ResourceList `json:"defaultResourceRequests,omitempty"`

	// MaxResourceLimits defines maximum allowed resource limits
	MaxResourceLimits corev1.ResourceList `json:"maxResourceLimits,omitempty"`
}

// ProbePolicies defines probe policies
type ProbePolicies struct {
	// RequireLivenessProbe enforces that containers have liveness probes
	RequireLivenessProbe bool `json:"requireLivenessProbe,omitempty"`

	// RequireReadinessProbe enforces that containers have readiness probes
	RequireReadinessProbe bool `json:"requireReadinessProbe,omitempty"`

	// DefaultLivenessProbe to apply if none is specified
	DefaultLivenessProbe *corev1.Probe `json:"defaultLivenessProbe,omitempty"`

	// DefaultReadinessProbe to apply if none is specified
	DefaultReadinessProbe *corev1.Probe `json:"defaultReadinessProbe,omitempty"`
}

// SecurityPolicies defines security context policies
type SecurityPolicies struct {
	// RequireNonRootUser enforces that containers run as non-root
	RequireNonRootUser bool `json:"requireNonRootUser,omitempty"`

	// RequireReadOnlyRootFilesystem enforces read-only root filesystem
	RequireReadOnlyRootFilesystem bool `json:"requireReadOnlyRootFilesystem,omitempty"`

	// DefaultSecurityContext to apply if none is specified
	DefaultSecurityContext *corev1.SecurityContext `json:"defaultSecurityContext,omitempty"`

	// RequiredCapabilities lists capabilities that must be present
	RequiredCapabilities []corev1.Capability `json:"requiredCapabilities,omitempty"`

	// ForbiddenCapabilities lists capabilities that must not be present
	ForbiddenCapabilities []corev1.Capability `json:"forbiddenCapabilities,omitempty"`
}

// PolicyRuleStatus defines the observed state of PolicyRule.
type PolicyRuleStatus struct {
	// LastApplied timestamp when the policy was last applied
	LastApplied *metav1.Time `json:"lastApplied,omitempty"`

	// EnforcedNamespaces lists namespaces where this policy is currently enforced
	EnforcedNamespaces []string `json:"enforcedNamespaces,omitempty"`

	// ViolationCount tracks the number of policy violations detected
	ViolationCount int32 `json:"violationCount,omitempty"`

	// Conditions represent the latest available observations of the policy's current state
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// PolicyRule is the Schema for the policyrules API.
type PolicyRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PolicyRuleSpec   `json:"spec,omitempty"`
	Status PolicyRuleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PolicyRuleList contains a list of PolicyRule.
type PolicyRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PolicyRule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PolicyRule{}, &PolicyRuleList{})
}
