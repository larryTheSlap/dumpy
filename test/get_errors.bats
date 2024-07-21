setup_file() {  
    PROJECT_ROOT="$( cd "$( dirname "$BATS_TEST_FILENAME" )/.." >/dev/null 2>&1 && pwd )"
    PATH="$PROJECT_ROOT/test:$PATH"

    export MANIFEST_PATH=$PROJECT_ROOT/test/manifest
    export CAP_NAME="test-get"

}

setup () {
    load 'test_helper/bats-support/load'
    load 'test_helper/bats-assert/load'
}

# GET POD CAPTURE WRONG ARG NUMBER
@test "get pod capture with wrong arg number ==> >1 arg" {
    run kubectl dumpy get ${CAP_NAME} someshit
    assert_output --partial 'Error: too many arguments, get command require capture name or nothing. use -h for help'
}  


# GET NON-EXISTENT CAPTURE 
@test "get non-existent capture" {
    run kubectl dumpy get ${CAP_NAME}-somerndshit 
    assert_output --partial 'Error: '"${CAP_NAME}"'-somerndshit sniffers not found'
}  