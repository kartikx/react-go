#!/bin/bash

# Setup script for testing multi-agent system with Kind
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_step() {
    echo -e "${BLUE}🔄 $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check prerequisites
check_prereqs() {
    print_step "Checking prerequisites..."
    
    if ! command -v kind &> /dev/null; then
        print_error "kind is not installed. Install with: go install sigs.k8s.io/kind@latest"
        exit 1
    fi
    
    if ! command -v kubectl &> /dev/null; then
        print_error "kubectl is not installed. Please install kubectl first."
        exit 1
    fi
    
    if ! command -v docker &> /dev/null; then
        print_error "docker is not installed. Please install docker first."
        exit 1
    fi
    
    if [ -z "$ANTHROPIC_API_KEY" ]; then
        print_error "ANTHROPIC_API_KEY environment variable is not set"
        print_warning "Export your API key: export ANTHROPIC_API_KEY=your-key-here"
        exit 1
    fi
    
    print_success "All prerequisites met"
}

# Create kind cluster
create_cluster() {
    print_step "Creating kind cluster..."
    
    # Delete existing cluster if it exists
    if kind get clusters | grep -q "agent-cluster"; then
        print_warning "Deleting existing agent-cluster..."
        kind delete cluster --name agent-cluster
    fi
    
    # Create new cluster
    kind create cluster --config kind-config.yaml --wait 300s
    
    # Verify cluster is ready
    kubectl cluster-info --context kind-agent-cluster
    
    print_success "Kind cluster created successfully"
}

# Build Docker images and load into kind
build_and_load_images() {
    print_step "Building Docker images..."
    
    # Build images
    docker build -f Dockerfile.coder -t coder-agent:latest .
    docker build -f Dockerfile.doc -t doc-agent:latest .
    
    print_step "Loading images into kind cluster..."
    
    # Load images into kind
    kind load docker-image coder-agent:latest --name agent-cluster
    kind load docker-image doc-agent:latest --name agent-cluster
    
    print_success "Images built and loaded"
}

# Create Kubernetes secret
create_secret() {
    print_step "Creating Anthropic API secret..."
    
    # Create namespace
    kubectl create namespace agents --dry-run=client -o yaml | kubectl apply -f -
    
    # Create secret
    kubectl create secret generic anthropic-secret \
        --from-literal=api-key="$ANTHROPIC_API_KEY" \
        --namespace agents \
        --dry-run=client -o yaml | kubectl apply -f -
    
    print_success "Secret created"
}

# Deploy agents
deploy_agents() {
    print_step "Deploying agents to Kubernetes..."
    
    # Deploy documentation agent (4 replicas)
    kubectl apply -f k8s/doc-agent-deployment.yaml
    
    # Deploy coder agent (2 replicas)
    kubectl apply -f k8s/coder-agent-deployment.yaml
    
    # Deploy NodePort services for external access
    kubectl apply -f k8s/nodeport-services.yaml
    
    print_step "Waiting for deployments to be ready..."
    kubectl rollout status deployment/doc-agent -n agents --timeout=300s
    kubectl rollout status deployment/coder-agent -n agents --timeout=300s
    
    print_success "All deployments ready"
}

# Show status and testing info
show_status() {
    echo ""
    print_success "🎉 Multi-Agent System deployed successfully!"
    echo ""
    
    echo -e "${BLUE}📊 Cluster Status:${NC}"
    kubectl get nodes
    echo ""
    kubectl get pods -n agents -o wide
    echo ""
    kubectl get services -n agents
    echo ""
    
    echo -e "${BLUE}🔍 Testing Commands:${NC}"
    echo -e "  # Test coder agent (accessible on localhost:9080):"
    echo -e "  ${YELLOW}curl -X POST http://localhost:9080/coder -d 'search for encoding/json'${NC}"
    echo ""
    echo -e "  # Test doc agent directly (accessible on localhost:9081):"
    echo -e "  ${YELLOW}curl -X POST http://localhost:9081/doc -d 'encoding/json'${NC}"
    echo ""
    echo -e "  # Watch logs:"
    echo -e "  ${YELLOW}kubectl logs -f deployment/doc-agent -n agents${NC}"
    echo -e "  ${YELLOW}kubectl logs -f deployment/coder-agent -n agents${NC}"
    echo ""
    echo -e "  # Scale agents:"
    echo -e "  ${YELLOW}kubectl scale deployment doc-agent --replicas=6 -n agents${NC}"
    echo ""
    echo -e "  # Load test (requires multiple terminal windows):"
    echo -e "  ${YELLOW}for i in {1..10}; do curl -X POST http://localhost:9080/coder -d 'test request'; done${NC}"
    echo ""
    
    echo -e "${BLUE}🧹 Cleanup:${NC}"
    echo -e "  ${YELLOW}kind delete cluster --name agent-cluster${NC}"
    echo ""
}

# Main execution
main() {
    echo -e "${BLUE}🚀 Setting up Multi-Agent System with Kind${NC}"
    echo ""
    
    check_prereqs
    create_cluster
    build_and_load_images
    create_secret
    deploy_agents
    show_status
}

# Run main function
main "$@"