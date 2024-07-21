setup_file() {  
    PROJECT_ROOT="$( cd "$( dirname "$BATS_TEST_FILENAME" )/.." >/dev/null 2>&1 && pwd )"
    PATH="$PROJECT_ROOT/test:$PATH"

    export MANIFEST_PATH=$PROJECT_ROOT/test/manifest
    export CAP_NAME="test-export"

}

setup () {
    load 'test_helper/bats-support/load'
    load 'test_helper/bats-assert/load'
}

# EXPORT POD CAPTURE WRONG ARG NUMBER
@test "export pod capture with wrong arg number ==> 1 arg" {
    run kubectl dumpy export ${CAP_NAME}
    assert_output --partial 'Error: export requires capture name and destination directory as arguments, use -h for help'
}  

@test "export pod capture with wrong arg number ==> >2 args" {
    run kubectl dumpy export ${CAP_NAME} /tmp/dumps skusku
    assert_output --partial 'Error: export requires capture name and destination directory as arguments, use -h for help'
}  

# EXPORT NON-EXISTENT CAPTURE 
@test "export non-existent capture" {
    run kubectl dumpy export ${CAP_NAME}-somerndshit /tmp/dumps
    assert_output --partial 'Error: '"${CAP_NAME}"'-somerndshit sniffers not found'
}  