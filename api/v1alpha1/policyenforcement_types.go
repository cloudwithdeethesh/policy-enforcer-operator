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

// PolicyEnforcementSpec defines the desired state of PolicyEnforcement.
type PolicyEnforcementSpec struct {
	// PolicyRuleRef references the PolicyRule that this enforcement tracks
	PolicyRuleRef corev1.LocalObjectReference `json:"policyRuleRef"`

	// TargetRef references the target resource being enforced
	TargetRef corev1.ObjectReference `json:"targetRef"`

	// EnforcementTime when the enforcement was applied
	EnforcementTime *metav1.Time `json:"enforcementTime,omitempty"`
}

// PolicyEnforcementStatus defines the observed state of PolicyEnforcement.
type PolicyEnforcementStatus struct {
	// Phase indicates the current enforcement phase
	// +kubebuilder:validation:Enum=Pending;Enforced;Failed;Skipped
	Phase string `json:"phase,omitempty"`

	// ViolationsDetected lists the policy violations that were found
	ViolationsDetected []PolicyViolation `json:"violationsDetected,omitempty"`

	// ActionsApplied lists the enforcement actions that were taken
	ActionsApplied []EnforcementAction `json:"actionsApplied,omitempty"`

	// LastEnforced timestamp when enforcement was last applied
	LastEnforced *metav1.Time `json:"lastEnforced,omitempty"`

	// Message provides human-readable information about the enforcement status
	Message string `json:"message,omitempty"`

	// Conditions represent the latest available observations of the enforcement state
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// PolicyViolation describes a specific policy violation
type PolicyViolation struct {
	// Type of violation (e.g., "MissingResourceLimits", "MissingProbe")
	Type string `json:"type"`

	// Field path where the violation was found
	Field string `json:"field,omitempty"`

	// Message describing the violation
	Message string `json:"message"`

	// Severity of the violation
	// +kubebuilder:validation:Enum=Low;Medium;High;Critical
	Severity string `json:"severity,omitempty"`
}

// EnforcementAction describes an action taken during enforcement
type EnforcementAction struct {
	// Type of action (e.g., "AddedResourceLimits", "AddedProbe")
	Type string `json:"type"`

	// Field path where the action was applied
	Field string `json:"field,omitempty"`

	// OldValue before the action (if applicable)
	OldValue string `json:"oldValue,omitempty"`

	// NewValue after the action
	NewValue string `json:"newValue"`

	// Timestamp when the action was applied
	Timestamp metav1.Time `json:"timestamp"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// PolicyEnforcement is the Schema for the policyenforcements API.
type PolicyEnforcement struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PolicyEnforcementSpec   `json:"spec,omitempty"`
	Status PolicyEnforcementStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PolicyEnforcementList contains a list of PolicyEnforcement.
type PolicyEnforcementList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PolicyEnforcement `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PolicyEnforcement{}, &PolicyEnforcementList{})
}
