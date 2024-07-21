setup_file() {  
    PROJECT_ROOT="$( cd "$( dirname "$BATS_TEST_FILENAME" )/.." >/dev/null 2>&1 && pwd )"
    PATH="$PROJECT_ROOT/test:$PATH"

    export MANIFEST_PATH=$PROJECT_ROOT/test/manifest
    export CAP_NAME="test-stop"

    if [[ $(kubectl create ns $NAMESPACE | grep -i "already exist") != "" ]]
    then
        echo "${NAMESPACE} already created, proceeding.."
    fi

    kubectl apply -f $MANIFEST_PATH/pod_diffNS.yml
    sleep 5
    kubectl dumpy capture --name ${CAP_NAME} pod test-pod -n test-ns
}

setup () {
    load 'test_helper/bats-support/load'
    load 'test_helper/bats-assert/load'
}

teardown_file() {
    kubectl delete -f $MANIFEST_PATH/pod_diffNS.yml
    kubectl dumpy delete ${CAP_NAME} -n test-ns
}

# STOP POD CAPTURE SPECIFIC NAMESPACE
@test "stop pod capture on specific namespace ==> ns name test-ns" {
    run kubectl dumpy stop ${CAP_NAME} -n test-ns
    assert_output --partial ''"${CAP_NAME}"' sniffers have been successfully stopped'
}  
