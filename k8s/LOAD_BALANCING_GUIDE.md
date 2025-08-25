# Kubernetes Load Balancing for Multi-Agent Systems

## Your Load Balancing Questions - Answered!

### Q: "How does coder agent know which documentation agent instance to talk to?"
**A: It doesn't need to know!** 

The coder agent talks to `doc-agent-service:8080/doc` (a stable DNS name). Kubernetes Service automatically routes each request to a healthy doc agent pod using round-robin load balancing.

```
Coder Agent Instance 1 ──┐
                         ├─→ doc-agent-service ──┐
Coder Agent Instance 2 ──┘    (Load Balancer)    ├─→ Doc Agent Pod 1
                                                  ├─→ Doc Agent Pod 2  
                                                  ├─→ Doc Agent Pod 3
                                                  └─→ Doc Agent Pod 4
```

### Q: "How does the response come back to the same coder agent?"
**A: HTTP request-response magic!**

Since you're using synchronous HTTP calls, the TCP connection remains open during the request. The response automatically returns through the same connection to the originating coder agent pod.

```
Request Flow:
Coder Pod A ─[HTTP POST]─→ Service ─→ Doc Pod 2
Coder Pod A ←─[Response]─── Service ←─ Doc Pod 2
```

## How Kubernetes Services Work

### 1. **Service Discovery**
```yaml
# In coder agent deployment:
env:
- name: DOC_AGENT_URL
  value: "http://doc-agent-service:8080/doc"
```

`doc-agent-service` is a **stable DNS name** that resolves to the Service's ClusterIP.

### 2. **Load Balancing**
```yaml
apiVersion: v1
kind: Service
metadata:
  name: doc-agent-service
spec:
  selector:
    app: doc-agent  # Targets all pods with this label
  ports:
  - port: 8080      # Service port
    targetPort: 8080 # Pod port
```

The Service automatically:
- Discovers all pods matching `app: doc-agent`
- Load balances requests across healthy pods
- Removes unhealthy pods from rotation

### 3. **Health-Based Routing**
```yaml
readinessProbe:
  httpGet:
    path: /health
    port: 8080
```

Only pods passing readiness checks receive traffic. This ensures requests only go to functioning agents.

## Production Architecture

### Current Setup (Updated)
- **2x Coder Agent pods** (replicas: 2)
- **4x Doc Agent pods** (replicas: 4)  
- **Kubernetes Services** for load balancing
- **Health checks** for reliability

### Traffic Flow
1. External request → `coder-agent-service` → Random coder pod
2. Coder pod makes internal request → `doc-agent-service` → Random doc pod
3. Doc pod processes request and responds
4. Response travels back through the same TCP connection
5. Coder pod receives response and processes it
6. Final response sent back to external client

## Advanced Load Balancing Options

### 1. **Session Affinity** (if needed)
```yaml
apiVersion: v1
kind: Service
metadata:
  name: doc-agent-service
spec:
  selector:
    app: doc-agent
  ports:
  - port: 8080
    targetPort: 8080
  sessionAffinity: ClientIP  # Sticky sessions
```

### 2. **Custom Load Balancing Algorithms**
For more sophisticated routing, consider:
- **Istio Service Mesh**: Weighted routing, circuit breakers
- **NGINX Ingress**: Custom load balancing algorithms
- **Linkerd**: Automatic load balancing with latency awareness

### 3. **Horizontal Pod Autoscaling**
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: doc-agent-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: doc-agent
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

## Testing Load Balancing

### 1. **Deploy and Scale**
```bash
# Apply all manifests
kubectl apply -f k8s/

# Verify pods are running
kubectl get pods

# Check service endpoints
kubectl get endpoints doc-agent-service
```

### 2. **Test Load Distribution**
```bash
# Port forward to coder agent
kubectl port-forward service/coder-agent-service 8080:8080

# Make multiple requests and observe which doc agent handles them
for i in {1..10}; do
  curl -X POST http://localhost:8080/coder -d "search for encoding/json"
done
```

### 3. **Monitor Load Balancing**
```bash
# Watch pod logs to see request distribution
kubectl logs -f deployment/doc-agent

# Check service load balancing
kubectl describe service doc-agent-service
```

## Key Benefits for Your Presentation

1. **Simplified Architecture**: Developers don't manage instance discovery
2. **Automatic Scaling**: Add/remove pods without code changes  
3. **Health-Aware Routing**: Traffic only goes to healthy agents
4. **Zero-Downtime Deployments**: Rolling updates with no service interruption
5. **Go's Advantage**: Static binaries make containerization trivial

## Next Steps

1. **Test locally with kind/minikube**
2. **Add observability** (Prometheus metrics, distributed tracing)
3. **Implement circuit breakers** for fault tolerance
4. **Add rate limiting** for production safety

This architecture showcases Go's cloud-native strengths perfectly for your presentation!