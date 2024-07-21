setup_file() {  
    PROJECT_ROOT="$( cd "$( dirname "$BATS_TEST_FILENAME" )/.." >/dev/null 2>&1 && pwd )"
    PATH="$PROJECT_ROOT/test:$PATH"

    export MANIFEST_PATH=$PROJECT_ROOT/test/manifest
    export CAP_NAME="test-capture"
    export NAMESPACE="test-ns"

    if [[ $(kubectl create ns $NAMESPACE | grep -i "already exist") != "" ]]
    then
        echo "${NAMESPACE} already created, proceeding.."
    fi
    kubectl apply -f $MANIFEST_PATH/pod_currNS.yml
    kubectl apply -f $MANIFEST_PATH/pod_diffNS.yml
    sleep 5
}

setup () {
    load 'test_helper/bats-support/load'
    load 'test_helper/bats-assert/load'
}

teardown_file() {
    kubectl dumpy delete ${CAP_NAME}
    kubectl dumpy delete ${CAP_NAME}-diff -n $NAMESPACE
    kubectl dumpy delete ${CAP_NAME}-diffcurr
    kubectl dumpy delete ${CAP_NAME}-diffdiff -n $NAMESPACE
    kubectl dumpy delete ${CAP_NAME}-container
    kubectl dumpy delete ${CAP_NAME}-image
    kubectl dumpy delete ${CAP_NAME}-pvc
    kubectl dumpy delete ${CAP_NAME}-secret

    kubectl delete -f $MANIFEST_PATH/pod_currNS.yml
    kubectl delete -f $MANIFEST_PATH/pod_diffNS.yml
}


# CAPTURE MUTLIPLE NS CONFIGURATIONS
@test "capture pod multiple ns configurations ==> target | sniffer curr ns" {
    run kubectl dumpy capture --name ${CAP_NAME} pod test-pod
    assert_output --partial 'All dumpy sniffers are Ready.'
}  

@test "capture pod multiple ns configurations ==> target diff ns | sniffer diff ns" {
    run kubectl dumpy capture --name ${CAP_NAME}-diff pod test-pod -n test-ns
    assert_output --partial 'All dumpy sniffers are Ready.'
}  

@test "capture pod multiple ns configurations ==> target diff ns | sniffer curr ns" {
    run kubectl dumpy capture --name ${CAP_NAME}-diffcurr pod test-pod -t test-ns
    assert_output --partial 'All dumpy sniffers are Ready.'
}  

@test "capture pod multiple ns configurations ==> target | sniffer diff ns" {
    run kubectl dumpy capture --name ${CAP_NAME}-diffdiff pod test-pod -n test-ns -t default
    assert_output --partial 'All dumpy sniffers are Ready.'
}  

# CAPTURE SPECIFIC CONTAINER
@test "capture pod specific container ==> target container name nginx-container" {
    kubectl dumpy capture --name ${CAP_NAME}-container pod test-pod -c nginx-container
    run kubectl dumpy get ${CAP_NAME}-container
    assert_output --partial 'container: nginx-container'
}  

# APTURE SPECIFIC IMAGE
@test "capture pod specific dumpy image ==> image larrytheslap/dumpy:0.1.0" {
    kubectl dumpy capture --name ${CAP_NAME}-image pod test-pod -i larrytheslap/dumpy:0.1.0
    run kubectl dumpy get ${CAP_NAME}-image
    assert_output --partial 'image: larrytheslap/dumpy:0.1.0'
} 

# CAPTURE PVC MOUNT
@test "capture pod with pvc mount ==> pvc name test-pvc" {
    kubectl dumpy capture --name ${CAP_NAME}-pvc pod test-pod -v test-pvc
    run kubectl dumpy get ${CAP_NAME}-pvc
    assert_output --partial 'pvc: test-pvc'
} 

# CAPTURE REGISTRY SECRET
@test "capture pod with registry secret ==> secret name test-secret" {
    kubectl dumpy capture --name ${CAP_NAME}-secret pod test-pod -s test-secret
    run kubectl dumpy get ${CAP_NAME}-secret
    assert_output --partial 'secret: test-secret'
} 

