#!/bin/bash

# Load balancing test script for the multi-agent system
set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_test() {
    echo -e "${BLUE}🧪 $1${NC}"
}

print_result() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ️  $1${NC}"
}

# Check if services are accessible
check_services() {
    print_test "Checking service availability..."
    
    # Test coder agent health
    if curl -s http://localhost:9080/health > /dev/null 2>&1; then
        print_result "Coder agent is healthy"
    else
        echo "❌ Coder agent not accessible on localhost:9080"
        exit 1
    fi
    
    # Test doc agent health  
    if curl -s http://localhost:9081/health > /dev/null 2>&1; then
        print_result "Doc agent is healthy"
    else
        echo "❌ Doc agent not accessible on localhost:9081"
        exit 1
    fi
}

# Test direct communication with doc agent
test_doc_agent_direct() {
    print_test "Testing documentation agent directly..."
    
    response=$(curl -s -X POST http://localhost:9081/doc -d "encoding/json")
    if [[ $response == *"encoding/json"* ]]; then
        print_result "Doc agent responded correctly"
    else
        echo "❌ Doc agent didn't respond as expected"
        echo "Response: $response"
    fi
}

# Test coder → doc agent communication (load balancing)
test_load_balancing() {
    print_test "Testing load balancing through coder → doc agent communication..."
    
    print_info "Sending 5 requests through coder agent to test load distribution..."
    
    for i in {1..5}; do
        echo -n "Request $i: "
        response=$(curl -s -X POST http://localhost:9080/coder -d "check documentation for encoding/json" || echo "FAILED")
        
        if [[ $response == *"encoding/json"* ]] || [[ $response == *"documentation"* ]]; then
            echo "✅ Success"
        else
            echo "❌ Failed - Response: $response"
        fi
        
        # Small delay to see different pod handling
        sleep 0.5
    done
    
    print_result "Load balancing test completed"
}

# Show pod distribution
show_pod_status() {
    print_test "Current pod distribution..."
    
    echo ""
    kubectl get pods -n agents -o wide
    echo ""
    kubectl get services -n agents
    echo ""
}

# Test scaling
test_scaling() {
    print_test "Testing horizontal scaling..."
    
    print_info "Current doc agent replicas:"
    kubectl get deployment doc-agent -n agents
    
    print_info "Scaling doc agents to 6 replicas..."
    kubectl scale deployment doc-agent --replicas=6 -n agents
    
    print_info "Waiting for new pods to be ready..."
    kubectl rollout status deployment/doc-agent -n agents --timeout=120s
    
    print_result "Scaling complete"
    kubectl get pods -n agents | grep doc-agent
}

# Performance test
performance_test() {
    print_test "Running performance test..."
    
    print_info "Sending 20 concurrent requests to test load distribution..."
    
    # Create temp directory for results
    temp_dir=$(mktemp -d)
    
    # Send concurrent requests
    for i in {1..20}; do
        (
            start_time=$(date +%s.%3N)
            response=$(curl -s -X POST http://localhost:9080/coder -d "performance test $i")
            end_time=$(date +%s.%3N)
            duration=$(echo "$end_time - $start_time" | bc)
            
            if [[ $response == *"test"* ]] || [[ $? -eq 0 ]]; then
                echo "$i: SUCCESS - ${duration}s" >> "$temp_dir/results.txt"
            else
                echo "$i: FAILED - ${duration}s" >> "$temp_dir/results.txt"
            fi
        ) &
    done
    
    # Wait for all background jobs
    wait
    
    # Show results
    echo ""
    print_info "Performance test results:"
    cat "$temp_dir/results.txt" | sort -n
    
    success_count=$(grep -c "SUCCESS" "$temp_dir/results.txt" || echo "0")
    total_count=$(wc -l < "$temp_dir/results.txt")
    
    print_result "Performance test: $success_count/$total_count requests successful"
    
    # Cleanup
    rm -rf "$temp_dir"
}

# Main test sequence
main() {
    echo -e "${BLUE}🚀 Starting Load Balancing Tests${NC}"
    echo ""
    
    check_services
    echo ""
    
    show_pod_status
    echo ""
    
    test_doc_agent_direct
    echo ""
    
    test_load_balancing
    echo ""
    
    test_scaling
    echo ""
    
    performance_test
    echo ""
    
    print_result "🎉 All tests completed!"
    echo ""
    
    print_info "Check pod logs to see which instances handled requests:"
    echo "kubectl logs -f deployment/doc-agent -n agents"
}

# Check prerequisites
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl not found. Please install kubectl."
    exit 1
fi

if ! command -v curl &> /dev/null; then
    echo "❌ curl not found. Please install curl."
    exit 1
fi

# Run main function
main "$@"