#!/bin/bash

cat dumpy_ASCII.txt ; echo -e ""

TCPDUMPFILTERS=$1
LOG_INFO="[INFO] ####"
LOG_ERR="[ERR] ####"
LOG_OK="[SUCCESS] ####"
LOG_WARN="[WARN] ####"

# RUN TCPDUMP ON TARGET PODS||NODES
if [[ $DUMPY_TARGET_TYPE == "node" ]]
then
    echo "$LOG_INFO starting capture on target node $TARGET_NODE .."
    nsenter -t 1 -n tcpdump $TCPDUMPFILTERS -w - | tee /tmp/dumpy/${CAPTURE_NAME}-${TARGET_NODE}.pcap | tcpdump -r - &
    TCPDUMP_PID=$!
else
    # GET CONTAINER CLI BINARIES PATHS
    commands=("crictl" "nerdctl" "ctr" "docker")
    for binary in "${commands[@]}"
    do
        if [[ $(nsenter -t 1 -m -n ls /usr/local/bin | grep $binary) != "" ]]; then
            BIN_PATH_LIST+=("/usr/local/bin/$binary")
        elif [[ $(nsenter -t 1 -m -n ls /usr/bin | grep $binary) != "" ]]; then
            BIN_PATH_LIST+=("/usr/bin/$binary")
        fi
    done
    if [[ -z "$BIN_PATH_LIST" ]]; then
        echo "container cli not found, tried [crictl, nerdctl, ctr, docker]"
        exit 1
    fi

    # GET TARGET CONTAINER PID FROM BIN INSPECT
    echo "$LOG_INFO target container ID: $TARGET_CONTAINERID"
    for BINPATH in "${BIN_PATH_LIST[@]}"
    do
        BIN=$(echo $BINPATH | awk -F '/' '{print $NF}')
        echo "$LOG_INFO trying container cli : $BIN"
        if [[ $BIN == "crictl" ]]
        then
            TARGET_PID=$(nsenter -t 1 -m -n $BINPATH inspect $TARGET_CONTAINERID | jq .info.pid)
        elif [[ $BIN == "nerdctl" || $BIN == "docker" ]]
        then
            TARGET_PID=$(nsenter -t 1 -m -n $BINPATH  inspect $TARGET_CONTAINERID | jq .[0].State.Pid)
        elif [[ $BIN == "ctr" ]]
        then
            # GET CONTAINER NAMESPACES
            NAMESPACES=$(nsenter -t 1 -m -n /usr/bin/ctr ns ls -q)
            [ -z "$NAMESPACES" ] && NAMESPACES="default"
            # SCAN EVERY NAMESPACE FOR TARGET CONTAINER
            for NAMESPACE in $NAMESPACES ; do
                TARGET_PID=$(nsenter -t 1 -m -n $BINPATH -n "$NAMESPACE" task ls | awk "/$TARGET_CONTAINERID/ { print \$2 }")
                if [[ "$TARGET_PID" =~ ^[0-9]+$ ]] ; then
                    echo "$LOG_INFO target container PID : $TARGET_PID"
                    break 2
                fi
            done
        fi

        if [[ "$TARGET_PID" =~ ^[0-9]+$ ]]
        then
            echo "$LOG_INFO target container PID : $TARGET_PID"
            break
        fi

    done
    if [[ ! "$TARGET_PID" =~ ^[0-9]+$ ]]
    then
        echo "$LOG_ERR No PID found for target container"
        exit 1
    fi

    # RUN TCPDUMP IN PID NAMESPACE
    echo "$LOG_INFO starting capture on target pod $TARGET_POD .."
    nsenter -t $TARGET_PID -n tcpdump $TCPDUMPFILTERS -w - | tee /tmp/dumpy/${CAPTURE_NAME}-${TARGET_POD}.pcap | tcpdump -r - &
    TCPDUMP_PID=$!
fi

# CHECK TCPDUMP PROCESS
if [[ "$TCPDUMP_PID" =~ ^[0-9]+$ ]]
then
    echo "$LOG_INFO sniffer tcpdump PID: $TCPDUMP_PID"
else
    echo "$LOG_ERR Dumpy sniffer failed to run Tcpdump on target, terminating.."
    exit 1
fi

# INIT DUMPY STATUS FILES
TERMINATION_FILE="/tmp/dumpy/termination_flag"
HEALTH_FILE="/tmp/dumpy/healthy"
echo "false" > $TERMINATION_FILE ;

# WAIT FOR TERMINATION FROM DUMPY CMD
RETRY_COUNT=0
while [[ $(cat $TERMINATION_FILE) == "false" ]]
do
    sleep 2
    if ! kill -0 $TCPDUMP_PID > /dev/null 2>&1
    then
        if [[ $RETRY_COUNT == 0 ]]
        then
            echo "$LOG_ERR network capture failed to initiate, terminating sniffer..."
            exit 1
        fi
        echo "$LOG_INFO network capture has been interrupted, waiting for SIGTERM..."
    fi
    if [[ $RETRY_COUNT == 0 ]]
    then
        echo "true" > $HEALTH_FILE
    fi
    RETRY_COUNT=$((RETRY_COUNT + 1))
    sleep 10
done

# STOP NETWORK CAPTURE TCPDUMP
if kill $TCPDUMP_PID > /dev/null 2>&1
then
    sleep 2
    echo "$LOG_OK Dumpy sniffer is done sniffing!"
    exit 0
else
    echo "$LOG_ERR sniff process was terminated, BAAAD Dumpy!"
    exit 1
fi
