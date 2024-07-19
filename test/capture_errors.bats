setup_file() {  
    PROJECT_ROOT="$( cd "$( dirname "$BATS_TEST_FILENAME" )/.." >/dev/null 2>&1 && pwd )"
    PATH="$PROJECT_ROOT/test/scripts:$PATH"

    export CAP_NAME="test-capture"
}

setup () {
    load 'test_helper/bats-support/load'
    load 'test_helper/bats-assert/load'
}


# CAPTURE NON-EXISTENT Resource 
@test "capture non-existent pod ==> pod name somerndshit" {
    run kubectl dumpy capture --name ${CAP_NAME} pod somerndshit
    assert_output --partial 'Error: pods "somerndshit" not found'
}  

@test "capture non-existent deployment ==> deploy name somerndshit" {
    run kubectl dumpy capture --name ${CAP_NAME} deploy somerndshit
    assert_output --partial 'Error: target resource pods not found'
}  

@test "capture non-existent type ==> type name somerndtype" {
    run kubectl dumpy capture --name ${CAP_NAME} somerndstype somerndshit
    assert_output --partial 'Error: unkown resource type, use -h for help'
}  

# CAPTURE MUTLIPLE ARG NUMBERS
@test "capture with wrong number of args ==> 1 arg" {
    run kubectl dumpy capture --name ${CAP_NAME} pod
    assert_output --partial 'Error: not enough arguments, use -h for help'
}  

@test "capture with wrong number of args ==> >2 args" {
    run kubectl dumpy capture --name ${CAP_NAME} pod test-pod hmidalaloudi
    assert_output --partial 'Error: too many arguments, use -h for help'
}  