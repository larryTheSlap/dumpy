setup_file() {  
    PROJECT_ROOT="$( cd "$( dirname "$BATS_TEST_FILENAME" )/.." >/dev/null 2>&1 && pwd )"
    PATH="$PROJECT_ROOT/test/scripts:$PATH"

    export MANIFEST_PATH=$PROJECT_ROOT/test/scripts/manifest
    export CAP_NAME="test-restart"

}

setup () {
    load 'test_helper/bats-support/load'
    load 'test_helper/bats-assert/load'
}

# RESTART POD CAPTURE WRONG ARG NUMBER
@test "restart pod capture with wrong arg number ==> >1 arg" {
    run kubectl dumpy restart ${CAP_NAME} someshit
    assert_output --partial 'Error: unkown arguments, restart command require capture name. use -h for help'
}  


# RESTART NON-EXISTENT CAPTURE 
@test "restart non-existent capture" {
    run kubectl dumpy restart ${CAP_NAME}-somerndshit 
    assert_output --partial 'Error: '"${CAP_NAME}"'-somerndshit sniffers not found'
}  