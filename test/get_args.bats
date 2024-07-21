setup_file() {  
    PROJECT_ROOT="$( cd "$( dirname "$BATS_TEST_FILENAME" )/.." >/dev/null 2>&1 && pwd )"
    PATH="$PROJECT_ROOT/test:$PATH"

    export MANIFEST_PATH=$PROJECT_ROOT/test/manifest
    export CAP_NAME="test-get"
    export IMG_VER="0.2.0"

    kubectl apply -f $MANIFEST_PATH/pod_currNS.yml
    kubectl apply -f $MANIFEST_PATH/deploy_currNS.yml
    sleep 10
    kubectl dumpy capture --name ${CAP_NAME}-pod pod test-pod
    kubectl dumpy capture --name ${CAP_NAME}-deploy deploy test-deploy
    export RND_NODE=$(kubectl get node --no-headers | awk '{print $1}' | tac | head -n 1)
    kubectl dumpy capture --name ${CAP_NAME}-node node ${RND_NODE}
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
    run kubectl dumpy get
    assert_output --partial 'NAME             NAMESPACE  TARGET                  TARGETNAMESPACE  TCPDUMPFILTERS  SNIFFERS'
    assert_output --partial '----             ---------  ------                  ---------------  --------------  --------'
    assert_output --partial 'test-get-deploy  default    deployment/test-deploy  default          -i any          2/2'
    assert_output --partial 'test-get-node    default    node/'"${RND_NODE}"'                        -i any          1/1'
    assert_output --partial 'test-get-pod     default    pod/test-pod            default          -i any          1/1'
}  

# GET POD CAPTURE
@test "get pod capture" {
    sniff_pod=$(kubectl get pod -l dumpy-capture=test-get-pod --no-headers | awk '{print $1}')
    actual=$(kubectl dumpy get test-get-pod)

    expected=$(cat <<EOF
Getting capture details..

name: ${CAP_NAME}-pod
namespace: default
tcpdumpfilters: -i any
image: larrytheslap/dumpy:${IMG_VER}
targetSpec:
    name: test-pod
    namespace: default
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

# GET DEPLOY CAPTURE
@test "get deploy capture" {
    read sniff_pod1 sniff_pod2 <<<$(kubectl get pod -l dumpy-capture=${CAP_NAME}-deploy --no-headers | awk '{print $1}' | tr '\n' ' ')
    target_pod1=$(kubectl get pod $sniff_pod1 -o yaml | grep dumpy-target-pod: | awk '{print $2}')
    target_pod2=$(kubectl get pod $sniff_pod2 -o yaml | grep dumpy-target-pod: | awk '{print $2}')
    actual=$(kubectl dumpy get test-get-deploy)

    expected=$(cat <<EOF
Getting capture details..

name: ${CAP_NAME}-deploy
namespace: default
tcpdumpfilters: -i any
image: larrytheslap/dumpy:${IMG_VER}
targetSpec:
    name: test-deploy
    namespace: default
    type: deployment
    container: nginx
    items:
        ${target_pod1}  <-----  ${sniff_pod1} [Running]
        ${target_pod2}  <-----  ${sniff_pod2} [Running]
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

# GET NODE CAPTURE
@test "get node capture" {
    sniff_pod=$(kubectl get pod -l dumpy-capture=${CAP_NAME}-node --no-headers | awk '{print $1}')
    actual=$(kubectl dumpy get test-get-node)

    expected=$(cat <<EOF
Getting capture details..

name: ${CAP_NAME}-node
namespace: default
tcpdumpfilters: -i any
image: larrytheslap/dumpy:${IMG_VER}
targetSpec:
    type: node
    items:
        ${RND_NODE}  <-----  ${sniff_pod} [Running]
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