setup_file() {  
    PROJECT_ROOT="$( cd "$( dirname "$BATS_TEST_FILENAME" )/.." >/dev/null 2>&1 && pwd )"
    PATH="$PROJECT_ROOT/test:$PATH"

    export MANIFEST_PATH=$PROJECT_ROOT/test/manifest
    export CAP_NAME="test-get"
    export IMG_VER="0.2.0"

    if [[ $(kubectl create ns $NAMESPACE | grep -i "already exist") != "" ]]
    then
        echo "${NAMESPACE} already created, proceeding.."
    fi

    kubectl apply -f $MANIFEST_PATH/pod_diffNS.yml
    sleep 5
    kubectl dumpy capture --name ${CAP_NAME}-pod pod test-pod -n test-ns
    sleep 5
}

setup () {
    load 'test_helper/bats-support/load'
    load 'test_helper/bats-assert/load'
}

teardown_file() {
    kubectl delete -f $MANIFEST_PATH/pod_diffNS.yml
    kubectl dumpy delete ${CAP_NAME}-pod -n test-ns
}

# GET POD CAPTURE IN DIFFERENT NS
@test "get pod capture" {
    sniff_pod=$(kubectl get pod -l dumpy-capture=test-get-pod --no-headers -n test-ns | awk '{print $1}')
    actual=$(kubectl dumpy get test-get-pod -n test-ns)

    expected=$(cat <<EOF
Getting capture details..

name: ${CAP_NAME}-pod
namespace: test-ns
tcpdumpfilters: -i any
image: larrytheslap/dumpy:${IMG_VER}
targetSpec:
    name: test-pod
    namespace: test-ns
    type: pod
    container: test-pod
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