setup_file() {  
    PROJECT_ROOT="$( cd "$( dirname "$BATS_TEST_FILENAME" )/.." >/dev/null 2>&1 && pwd )"
    PATH="$PROJECT_ROOT/test:$PATH"

    export MANIFEST_PATH=$PROJECT_ROOT/test/manifest
    export CAP_NAME="test-restart"

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
    sleep 2
}

teardown_file() {
    kubectl delete -f $MANIFEST_PATH/pod_currNS.yml
    kubectl delete -f $MANIFEST_PATH/deploy_currNS.yml
}

# RESTART POD CAPTURE
@test "restart pod capture ==> pod name test-pod" {
    kubectl dumpy capture --name ${CAP_NAME} pod test-pod
    run kubectl dumpy restart ${CAP_NAME}
    assert_output --partial ''"${CAP_NAME}"' sniffers have been successfully redeployed'
}  

# RESTART DEPLOY CAPTURE
@test "restart deployment capture ==> deploy name test-deploy" {
    kubectl dumpy capture --name ${CAP_NAME} deploy test-deploy
    run kubectl dumpy restart ${CAP_NAME}
    assert_output --partial ''"${CAP_NAME}"' sniffers have been successfully redeployed'
}  

# RESTART NODE CAPTURE
@test "restart node capture" {
    RND_NODE=$(kubectl get node --no-headers | awk '{print $1}' | tac | head -n 1)
    kubectl dumpy capture --name ${CAP_NAME} node $RND_NODE
    run kubectl dumpy restart ${CAP_NAME}
    assert_output --partial ''"${CAP_NAME}"' sniffers have been successfully redeployed'
}  