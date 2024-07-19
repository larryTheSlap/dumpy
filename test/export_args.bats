setup_file() {  
    PROJECT_ROOT="$( cd "$( dirname "$BATS_TEST_FILENAME" )/.." >/dev/null 2>&1 && pwd )"
    PATH="$PROJECT_ROOT/test/scripts:$PATH"

    export MANIFEST_PATH=$PROJECT_ROOT/test/scripts/manifest
    export CAP_NAME="test-export"

    kubectl apply -f $MANIFEST_PATH/pod_currNS.yml
    kubectl apply -f $MANIFEST_PATH/deploy_currNS.yml
    sleep 5
}

setup () {
    load 'test_helper/bats-support/load'
    load 'test_helper/bats-assert/load'
}

teardown() {
    kubectl dumpy delete ${CAP_NAME}
}

teardown_file() {
    kubectl delete -f $MANIFEST_PATH/pod_currNS.yml
    kubectl delete -f $MANIFEST_PATH/deploy_currNS.yml
}

# EXPORT POD CAPTURE
@test "export pod capture ==> pod name test-pod" {
    kubectl dumpy capture --name ${CAP_NAME} pod test-pod
    run kubectl dumpy export ${CAP_NAME} /tmp/dumps
    assert_output --partial 'test-pod ---> path /tmp/dumps/'"${CAP_NAME}"'-test-pod.pcap'
}  

# EXPORT DEPLOY CAPTURE
@test "export deployment capture ==> deploy name test-deploy" {
    kubectl dumpy capture --name ${CAP_NAME} deploy test-deploy
    read p1 p2 <<< $(kubectl get pod -l app=test-deploy --no-headers | awk '{print $1}' | tr '\n' ' ')
    run kubectl dumpy export ${CAP_NAME} /tmp/dumps
    assert_output --partial ''"${p1}"' ---> path /tmp/dumps/'"${CAP_NAME}"'-'"${p1}"'.pcap'
    assert_output --partial ''"${p2}"' ---> path /tmp/dumps/'"${CAP_NAME}" '-'"${p2}"'.pcap'
}  

# EXPORT NODE CAPTURE
@test "export node capture ==> node name kind-worker" {
    kubectl dumpy capture --name ${CAP_NAME} node kind-worker
    run kubectl dumpy export ${CAP_NAME} /tmp/dumps
    assert_output --partial 'kind-worker ---> path /tmp/dumps/'"${CAP_NAME}"'-kind-worker.pcap'
}  