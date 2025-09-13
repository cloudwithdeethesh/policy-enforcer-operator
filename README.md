# Policy Enforcer Operator

Kubernetes Operator (built with Kubebuilder) that enforces platform policies by automatically detecting and fixing workloads missing resource limits, liveness/readiness probes, or security settings.

## Description

The Policy Enforcer Operator provides:

- **PolicyRule CRD**: Define organization-wide policies for resource limits, probes, and security defaults
- **PolicyEnforcement CRD**: Track enforcement actions and violations
- **Admission Webhooks**: Real-time policy enforcement during pod creation/updates
- **Controller**: Continuous policy monitoring and violation detection

## Features

### Resource Policies
- Enforce resource limits and requests on containers
- Set default resource values for non-compliant workloads
- Define maximum allowed resource limits
- Support for CPU and memory resources

### Probe Policies
- Require liveness and readiness probes
- Automatically inject default probes when missing
- Configurable probe settings (HTTP, TCP, exec)

### Security Policies
- Enforce non-root user requirements
- Require read-only root filesystems
- Set default security contexts
- Control Linux capabilities (required/forbidden)

### Enforcement Modes
- **enforce**: Block non-compliant workloads
- **warn**: Allow but warn about violations
- **audit**: Log violations for monitoring

## Getting Started

### Prerequisites
- Kubernetes 1.16+
- kubectl
- Go 1.21+ (for development)

### Installation

1. Install the CRDs:
```bash
make install
```

2. Deploy the operator:
```bash
make deploy
```

### Quick Start

1. Create a basic policy rule:
```yaml
apiVersion: policy.cloudwithdeethesh.com/v1alpha1
kind: PolicyRule
metadata:
  name: basic-policy
spec:
  enforcementAction: enforce
  resourcePolicies:
    requireResourceLimits: true
    defaultResourceLimits:
      cpu: "500m"
      memory: "512Mi"
  probePolicies:
    requireLivenessProbe: true
  securityPolicies:
    requireNonRootUser: true
```

2. Apply the policy:
```bash
kubectl apply -f policy.yaml
```

3. Test with a non-compliant pod:
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec:
  containers:
  - name: app
    image: nginx
    # This pod lacks resource limits, probes, and security context
```

The admission webhook will either:
- **Enforce mode**: Inject defaults or reject the pod
- **Warn mode**: Allow the pod but issue warnings
- **Audit mode**: Allow the pod and log the violation

## Examples

See the [examples/](examples/) directory for:
- Sample PolicyRule configurations
- Test pods (compliant and non-compliant)
- Different enforcement scenarios

## Development

### Building
```bash
make build
```

### Testing
```bash
make test
```

### Running locally
```bash
make install
make run
```

## API Reference

### PolicyRule

Defines policies to be enforced on workloads.

**Spec:**
- `selector`: Label selector for target resources
- `namespaces`: List of namespaces to enforce (empty = all)
- `enforcementAction`: "enforce", "warn", or "audit"
- `resourcePolicies`: Resource limit/request policies
- `probePolicies`: Liveness/readiness probe policies
- `securityPolicies`: Security context policies

**Status:**
- `lastApplied`: When policy was last processed
- `violationCount`: Number of detected violations
- `conditions`: Current policy status

### PolicyEnforcement

Tracks enforcement actions on specific resources.

**Spec:**
- `policyRuleRef`: Reference to the PolicyRule
- `targetRef`: Reference to the target resource
- `enforcementTime`: When enforcement was applied

**Status:**
- `phase`: Current enforcement state
- `violationsDetected`: List of violations found
- `actionsApplied`: List of enforcement actions taken

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

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
