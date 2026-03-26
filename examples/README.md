# Pod-Lens Examples

## Kubernetes Deployments

### k8s-deployment-with-pod-labels-file.yaml

A production-ready Kubernetes deployment example for pod-lens with security best practices applied.

#### Security Features

✅ **Pod Security Standards (PSS) Compliant:**
- `readOnlyRootFilesystem: true` - Immutable root filesystem prevents tampering
- `runAsNonRoot: true` - Runs as unprivileged user (UID 1000)
- `allowPrivilegeEscalation: false` - Prevents privilege escalation
- `drop: ["ALL"]` - Removes all Linux capabilities
- `seccompProfile: RuntimeDefault` - Uses default seccomp profile

✅ **Resource Management:**
- CPU requests: 50m, limits: 200m
- Memory requests: 64Mi, limits: 128Mi
- Temp volume with 64Mi limit for `/tmp`

✅ **Health Checks:**
- Liveness probe checks `/healthz` every 30s
- Readiness probe checks `/healthz` every 10s

✅ **Configuration:**
- ReadOnly ConfigMap mounting for pod labels
- Service Account with disabled token mounting
- Named ports for better visibility

#### Deployment

```bash
# Create resources
kubectl apply -f examples/k8s-deployment-with-pod-labels-file.yaml

# Check deployment status
kubectl rollout status deployment/pod-lens

# Port forward to access locally
kubectl port-forward svc/pod-lens 8080:80

# Clean up
kubectl delete -f examples/k8s-deployment-with-pod-labels-file.yaml
```

#### Customization

Edit the ConfigMap `pod-labels` to add or modify pod labels:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: pod-labels
data:
  labels.txt: |
    app.kubernetes.io/name="my-app"
    custom-label="my-value"
```

Or use environment variable instead:

```yaml
env:
- name: POD_LABELS
  value: "app=my-app,env=prod,version=1.0"
```

#### Trivy/Security Scan Results

This deployment passes Kubernetes security checks:
- ✅ KSV-0014: Read-only root filesystem configured
- ✅ KSV-0118: Non-default security context applied
- ✅ No privilege escalation allowed
- ✅ Non-root user enforcement
