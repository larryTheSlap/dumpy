setup_file() {  
    PROJECT_ROOT="$( cd "$( dirname "$BATS_TEST_FILENAME" )/.." >/dev/null 2>&1 && pwd )"
    PATH="$PROJECT_ROOT/test:$PATH"

    export CAP_NAME="test-delete"
}

setup () {
    load 'test_helper/bats-support/load'
    load 'test_helper/bats-assert/load'
}


# DELETE NON EXISTENT CAPTURE  
@test "delete non-existent capture" {
    run kubectl dumpy delete ${CAP_NAME}
    assert_output --partial 'Error: '"${CAP_NAME}"' sniffers not found'
}  

# DELETE WRONG ARG NUMBER
@test "delete with wrong arg number ==> >1 args" {
    run kubectl dumpy delete ${CAP_NAME} skoulouplain
    assert_output --partial 'Error: unkown arguments, stop command require capture name. use -h for help'
}  
