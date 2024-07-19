setup_file() {  
    PROJECT_ROOT="$( cd "$( dirname "$BATS_TEST_FILENAME" )/.." >/dev/null 2>&1 && pwd )"
    PATH="$PROJECT_ROOT/test/scripts:$PATH"

    export MANIFEST_PATH=$PROJECT_ROOT/test/scripts/manifest
    export CAP_NAME="test-stop"

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

# STOP POD CAPTURE
@test "stop pod capture ==> pod name test-pod" {
    kubectl dumpy capture --name ${CAP_NAME} pod test-pod
    run kubectl dumpy stop ${CAP_NAME}
    assert_output --partial ''"${CAP_NAME}"' sniffers have been successfully stopped'
}  

# STOP DEPLOY CAPTURE
@test "stop deployment capture ==> deploy name test-deploy" {
    kubectl dumpy capture --name ${CAP_NAME} deploy test-deploy
    run kubectl dumpy stop ${CAP_NAME}
    assert_output --partial ''"${CAP_NAME}"' sniffers have been successfully stopped'
}  

# STOP NODE CAPTURE
@test "stop node capture ==> node name kind-worker" {
    kubectl dumpy capture --name ${CAP_NAME} node kind-worker
    run kubectl dumpy stop ${CAP_NAME}
    assert_output --partial ''"${CAP_NAME}"' sniffers have been successfully stopped'
}  