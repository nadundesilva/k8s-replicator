# Troubleshooting Guide 🔧

Don't worry, we've got your back! 🤗 This guide helps you diagnose and resolve common issues with K8s Replicator. We've compiled solutions for the most frequently encountered problems.

## Installation Issues 🚀

### Problem: Operator fails to start

**Symptoms:**

- Pods are in `CrashLoopBackOff` state
- Controller logs show startup errors

**K8s Replicator-specific solutions:**

```bash
# Check K8s Replicator controller logs
kubectl logs -n k8s-replicator-system deployment/k8s-replicator-controller-manager

# Verify K8s Replicator RBAC permissions
kubectl auth can-i create secrets --as=k8s-replicator-system:serviceaccount:k8s-replicator-system:k8s-replicator-controller-manager
```

### Problem: OLM installation fails

**Symptoms:**

- Bundle installation fails
- Operator not appearing in OLM catalog

**K8s Replicator-specific solutions:**

```bash
# Check K8s Replicator OLM status
kubectl get csv -n k8s-replicator-system

# Verify K8s Replicator bundle configuration
operator-sdk bundle validate ./bundle
```

## Replication Problems 🔄

### Problem: Resources not replicating

**Symptoms:**

- Resources with replication label not appearing in target namespaces
- No replication events in logs

**K8s Replicator-specific solutions:**

```bash
# Check replication label
kubectl get secret my-secret -o yaml | grep replicator.nadundesilva.github.io/object-type

# Check namespace filtering
kubectl get namespace my-namespace -o yaml | grep replicator.nadundesilva.github.io/namespace-type

# Check replication events
kubectl logs -n k8s-replicator-system deployment/k8s-replicator-controller-manager | grep replication
```

**K8s Replicator requirements:**

1. **Replication Label**: `replicator.nadundesilva.github.io/object-type=replicated`
2. **Namespace Filtering**: Target namespace not labeled as `ignored`
3. **Resource Type**: Must have a registered replicator implementation

### Problem: Resources replicating to wrong namespaces

**Symptoms:**

- Resources appearing in namespaces where they shouldn't
- System namespaces getting replicated resources

**K8s Replicator-specific solutions:**

```bash
# Check namespace filtering labels
kubectl get namespaces -o yaml | grep replicator.nadundesilva.github.io

# Add ignore label to namespace
kubectl label namespace my-namespace replicator.nadundesilva.github.io/namespace-type=ignored

# Override ignore with managed label
kubectl label namespace my-namespace replicator.nadundesilva.github.io/namespace-type=managed
```

### Problem: Replicated resources not updating

**Symptoms:**

- Changes to source resource not reflected in replicas
- Stale data in target namespaces

**K8s Replicator-specific solutions:**

```bash
# Check replication update events
kubectl logs -n k8s-replicator-system deployment/k8s-replicator-controller-manager | grep update

# Verify replication ownership
kubectl get secret my-secret -o yaml | grep replicator.nadundesilva.github.io

# Force reconciliation
kubectl rollout restart deployment/k8s-replicator-controller-manager -n k8s-replicator-system
```

## Performance Issues ⚡

### Problem: Slow replication with many namespaces

**K8s Replicator-specific solutions:**

```bash
# Adjust controller resource limits for large namespace counts
kubectl patch deployment k8s-replicator-controller-manager -n k8s-replicator-system -p '{"spec":{"template":{"spec":{"containers":[{"name":"manager","resources":{"limits":{"cpu":"500m","memory":"512Mi"}}}]}}}}'
```

## K8s Replicator-Specific Errors 🚨

- **"secrets is forbidden"**: Check RBAC permissions for K8s Replicator service account
- **"resource already exists"**: Delete conflicting resource or check replication labels
- **Replication not working**: Double-check the `replicator.nadundesilva.github.io/object-type=replicated` label
- **Wrong namespace replication**: Double-check namespace labels for `ignored` or `managed` types

## Debugging K8s Replicator 🐛

**Enable K8s Replicator debug logging:**

```bash
kubectl patch deployment k8s-replicator-controller-manager -n k8s-replicator-system -p '{"spec":{"template":{"spec":{"containers":[{"name":"manager","args":["-zap-log-level=4"]}]}}}}'
```

**Check replication events:**

```bash
kubectl logs -n k8s-replicator-system deployment/k8s-replicator-controller-manager | grep replication
```

## Support 💬

See the main [Support section](README.md#support-) for all support options.

Still having issues? Don't hesitate to reach out! We're here to help! 🤝💙
