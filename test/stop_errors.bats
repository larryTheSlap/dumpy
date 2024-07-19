setup_file() {  
    PROJECT_ROOT="$( cd "$( dirname "$BATS_TEST_FILENAME" )/.." >/dev/null 2>&1 && pwd )"
    PATH="$PROJECT_ROOT/test/scripts:$PATH"

    export MANIFEST_PATH=$PROJECT_ROOT/test/scripts/manifest
    export CAP_NAME="test-stop"

}

setup () {
    load 'test_helper/bats-support/load'
    load 'test_helper/bats-assert/load'
}

# STOP POD CAPTURE WRONG ARG NUMBER
@test "stop pod capture with wrong arg number ==> >1 arg" {
    run kubectl dumpy stop ${CAP_NAME} someshit
    assert_output --partial 'Error: unkown arguments, stop command require capture name. use -h for help'
}  


# STOP NON-EXISTENT CAPTURE 
@test "stop non-existent capture" {
    run kubectl dumpy stop ${CAP_NAME}-somerndshit 
    assert_output --partial 'Error: '"${CAP_NAME}"'-somerndshit sniffers not found'
}  