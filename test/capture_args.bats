setup_file() {  
    PROJECT_ROOT="$( cd "$( dirname "$BATS_TEST_FILENAME" )/.." >/dev/null 2>&1 && pwd )"
    PATH="$PROJECT_ROOT/test:$PATH"

    export MANIFEST_PATH=$PROJECT_ROOT/test/manifest
    export CAP_NAME="test-capture"

    kubectl apply -f $MANIFEST_PATH/deploy_currNS.yml
    kubectl apply -f $MANIFEST_PATH/daemonset_currNS.yml
    kubectl apply -f $MANIFEST_PATH/replicaset_currNS.yml
    kubectl apply -f $MANIFEST_PATH/statefulset_currNS.yml
    sleep 20
}

setup () {
    load 'test_helper/bats-support/load'
    load 'test_helper/bats-assert/load'
}

teardown() {
    kubectl dumpy delete ${CAP_NAME}
}

teardown_file() {
    kubectl delete -f $MANIFEST_PATH/deploy_currNS.yml
    kubectl delete -f $MANIFEST_PATH/daemonset_currNS.yml
    kubectl delete -f $MANIFEST_PATH/replicaset_currNS.yml
    kubectl delete -f $MANIFEST_PATH/statefulset_currNS.yml
}

# CAPTURE DEPLOYMENT 
@test "capture deployment ==> deploy name test-deploy" {
    run kubectl dumpy capture --name ${CAP_NAME} deployment test-deploy
    assert_output --partial 'All dumpy sniffers are Ready.'
}  

# CAPTURE DAEMONSET 
@test "capture daemnoset ==> ds name test-ds" {
    run kubectl dumpy capture --name ${CAP_NAME} ds test-ds
    assert_output --partial 'All dumpy sniffers are Ready.'
}  

# CAPTURE REPLICASET
@test "capture replicaset ==> rs name test-rs" {
    run kubectl dumpy capture --name ${CAP_NAME} rs test-rs
    assert_output --partial 'All dumpy sniffers are Ready.'
}  

# CAPTURE STATEFULSET
@test "capture statefulset ==> sts name test-sts" {
    run kubectl dumpy capture --name ${CAP_NAME} sts test-sts
    assert_output --partial 'All dumpy sniffers are Ready.'
}  

# CAPTURE NODE
@test "capture node" {
    RND_NODE=$(kubectl get node --no-headers | awk '{print $1}' | tac | head -n 1)
    run kubectl dumpy capture --name ${CAP_NAME} node $RND_NODE
    assert_output --partial 'All dumpy sniffers are Ready.'
}  

# CAPTURE ALL NODES
@test "capture all nodes" {
    run kubectl dumpy capture --name ${CAP_NAME} node all
    assert_output --partial 'All dumpy sniffers are Ready.'
}  