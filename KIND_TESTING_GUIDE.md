# Kind Testing Guide for Multi-Agent Load Balancing

## Quick Start

### Prerequisites
```bash
# Install kind (if not installed)
go install sigs.k8s.io/kind@latest

# Install kubectl (if not installed)
# macOS: brew install kubectl
# Linux: https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/

# Set your Anthropic API key
export ANTHROPIC_API_KEY=your-api-key-here
```

### One-Command Setup
```bash
# This will create cluster, build images, and deploy everything
./setup-kind.sh
```

## What Gets Created

### Kind Cluster
- **1 Control Plane + 2 Worker Nodes** for realistic multi-node testing
- **Port forwards**: `localhost:9080` → coder agents, `localhost:9081` → doc agents
- **Multi-zone simulation** for demonstrating pod distribution

### Deployed Architecture
```
┌─────────────────┐    ┌─────────────────────────┐
│   Coder Agent   │    │   Documentation Agent  │
│   (2 replicas)  │────│     (4 replicas)       │
│                 │    │                         │
│  Port 30080     │    │      Port 30081        │
└─────────────────┘    └─────────────────────────┘
        │                           │
        │                           │
   ClusterIP Service           ClusterIP Service
   + NodePort Service          + NodePort Service
```

### Services Created
1. **Internal Load Balancing**: `doc-agent-service` (ClusterIP)
2. **External Access**: NodePort services for testing
3. **Health Checks**: `/health` endpoints with probes
4. **Service Discovery**: DNS-based communication

## Testing Commands

### Basic Health Check
```bash
# Check if services are up
curl http://localhost:9080/health  # Coder agent health
curl http://localhost:9081/health  # Doc agent health
```

### Test Load Balancing
```bash
# Run comprehensive load balancing tests
./test-load-balancing.sh
```

### Manual Testing
```bash
# Test coder → doc agent communication (load balanced)
curl -X POST http://localhost:9080/coder -d "search for encoding/json"

# Test doc agent directly
curl -X POST http://localhost:9081/doc -d "encoding/json"

# Send multiple requests to see load distribution
for i in {1..10}; do 
  echo "Request $i:"
  curl -X POST http://localhost:9080/coder -d "test request $i"
  echo ""
done
```

### Scaling Demo
```bash
# Check current pod count
kubectl get pods -n agents

# Scale up doc agents
kubectl scale deployment doc-agent --replicas=8 -n agents

# Wait for scaling
kubectl rollout status deployment/doc-agent -n agents

# Scale down coder agents
kubectl scale deployment coder-agent --replicas=1 -n agents

# Watch pod distribution across nodes
kubectl get pods -n agents -o wide
```

### Monitoring & Debugging
```bash
# Watch logs (see which pods handle requests)
kubectl logs -f deployment/doc-agent -n agents
kubectl logs -f deployment/coder-agent -n agents

# Check service endpoints
kubectl get endpoints -n agents

# Describe services to see load balancing
kubectl describe service doc-agent-service -n agents

# Check resource usage
kubectl top pods -n agents
```

## Load Balancing Verification

### Expected Behavior
1. **Coder Agent Requests**: Round-robin between 2 coder pods
2. **Doc Agent Requests**: Round-robin between 4 doc pods  
3. **Health Checks**: Only healthy pods receive traffic
4. **Scaling**: New pods automatically join load balancer
5. **Node Distribution**: Pods spread across worker nodes

### Verification Commands
```bash
# See which pods are handling requests
kubectl logs deployment/doc-agent -n agents | grep "Handling request"

# Check load balancer endpoints
kubectl get endpoints doc-agent-service -n agents -o yaml

# Verify service discovery
kubectl exec -it deployment/coder-agent -n agents -- nslookup doc-agent-service.agents.svc.cluster.local
```

## Performance Testing

### Concurrent Load Test
```bash
# Generate concurrent load
./test-load-balancing.sh

# Or manual concurrent test
for i in {1..50}; do
  (curl -X POST http://localhost:9080/coder -d "concurrent test $i" &)
done
wait
```

### Expected Results
- **Response Distribution**: Roughly equal across all doc agent pods
- **No Failed Requests**: All requests should succeed
- **Sub-second Latency**: Even with LLM calls (depending on model)
- **Linear Scaling**: More pods = higher throughput

## Presentation Demo Sequence

### 1. Setup Demo (30 seconds)
```bash
./setup-kind.sh
# Show cluster starting, images building, pods deploying
```

### 2. Architecture Overview (30 seconds)
```bash
kubectl get pods -n agents -o wide
kubectl get services -n agents
# Show 2 coder + 4 doc agents across multiple nodes
```

### 3. Load Balancing Demo (60 seconds)
```bash
./test-load-balancing.sh
# Show requests distributed across multiple pods
```

### 4. Live Scaling Demo (30 seconds)
```bash
kubectl scale deployment doc-agent --replicas=8 -n agents
kubectl rollout status deployment/doc-agent -n agents
# Show new pods joining load balancer in real-time
```

## Troubleshooting

### Common Issues

**Pods Stuck in Pending**
```bash
kubectl describe pods -n agents
# Check resource constraints or node capacity
```

**Service Not Accessible**
```bash
kubectl get services -n agents
kubectl describe service doc-agent-service -n agents
# Verify port mappings and selectors
```

**Load Balancing Not Working**
```bash
kubectl get endpoints -n agents
# Ensure all pods are in Ready state
```

**API Key Issues**
```bash
kubectl get secrets -n agents
kubectl describe secret anthropic-secret -n agents
# Verify secret is created correctly
```

### Reset Everything
```bash
# Complete cleanup
kind delete cluster --name agent-cluster

# Start fresh
./setup-kind.sh
```

## Cleanup

### Stop Cluster (Keeps Images)
```bash
kind delete cluster --name agent-cluster
```

### Complete Cleanup
```bash
kind delete cluster --name agent-cluster
docker rmi coder-agent:latest doc-agent:latest
```

## Key Benefits Demonstrated

1. **Zero-Config Load Balancing**: Kubernetes Services handle everything
2. **Health-Aware Routing**: Unhealthy pods automatically removed
3. **Horizontal Scaling**: Add/remove pods without code changes
4. **Multi-Node Distribution**: Pods spread across cluster nodes
5. **Service Discovery**: Stable DNS names for inter-service communication
6. **Go's Container Advantage**: Tiny images, fast startup, low resource usage

This setup perfectly demonstrates how Go's cloud-native features enable production-ready LLM agent systems!