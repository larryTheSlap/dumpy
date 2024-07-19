setup_file() {  
    PROJECT_ROOT="$( cd "$( dirname "$BATS_TEST_FILENAME" )/.." >/dev/null 2>&1 && pwd )"
    PATH="$PROJECT_ROOT/test/scripts:$PATH"

    export MANIFEST_PATH=$PROJECT_ROOT/test/scripts/manifest
    export CAP_NAME="test-get"

    kubectl apply -f $MANIFEST_PATH/pod_currNS.yml
    kubectl apply -f $MANIFEST_PATH/deploy_currNS.yml
    sleep 5
    kubectl dumpy capture --name ${CAP_NAME}-pod pod test-pod
    kubectl dumpy capture --name ${CAP_NAME}-deploy deploy test-deploy
    kubectl dumpy capture --name ${CAP_NAME}-node node kind-worker
    sleep 5
}

setup () {
    load 'test_helper/bats-support/load'
    load 'test_helper/bats-assert/load'
}

teardown_file() {
    kubectl delete -f $MANIFEST_PATH/pod_currNS.yml
    kubectl delete -f $MANIFEST_PATH/deploy_currNS.yml
    kubectl dumpy delete ${CAP_NAME}-pod
    kubectl dumpy delete ${CAP_NAME}-deploy
    kubectl dumpy delete ${CAP_NAME}-node
}

# GET LIST ALL CAPTURES
@test "get all captures in current namespace" {
    actual=$(kubectl dumpy get | tr '\n' ' ')
    expected=$(cat <<EOF | tr '\n' ' '
NAME             NAMESPACE  TARGET                  TARGETNAMESPACE  TCPDUMPFILTERS  SNIFFERS
----             ---------  ------                  ---------------  --------------  --------
test-get-deploy  foo-ns     deployment/test-deploy  foo-ns           -i any          2/2
test-get-node    foo-ns     node/kind-worker                         -i any          1/1
test-get-pod     foo-ns     pod/test-pod            foo-ns           -i any          1/1
EOF
    )
    if [[ $actual == $expected ]]; then
        run bash -c "echo 'get output match'; exit 0"

    else
        run bash -c "echo 'wrong get output'; exit 1"
    fi
    assert_success
}  

# GET POD CAPTURE
@test "get pod capture" {
    sniff_pod=$(kubectl get pod -l dumpy-capture=test-get-pod --no-headers | awk '{print $1}')
    actual=$(kubectl dumpy get test-get-pod)

    expected=$(cat <<EOF
Getting capture details..

name: ${CAP_NAME}-pod
namespace: foo-ns
tcpdumpfilters: -i any
image: larrytheslap/dumpy:0.2.0
targetSpec:
    name: test-pod
    namespace: foo-ns
    type: pod
    container: nginx-container
    items:
        test-pod  <-----  ${sniff_pod} [Running]
pvc: 
pullsecret: 
EOF
    )
    if [[ "$actual" == "$expected" ]]; then
        run bash -c "echo -e 'actual: ${actual}\nexpect: ${expected}' ; exit 0"

    else
        run bash -c "echo -e 'actual: ${actual}\nexpect: ${expected}'; exit 1"
    fi
    assert_success
}  
