#!/bin/bash

cat dumpy_ASCII.txt
echo -e ""

TCPDUMPFILTERS=$1
LOG_INFO="[INFO] ####"
LOG_ERR="[ERR] ####"
LOG_OK="[SUCCESS] ####"
LOG_WARN="[WARN] ####"

# RUN TCPDUMP on Target
if [[ $DUMPY_TARGET_TYPE == "node" ]]
then
    echo "$LOG_INFO starting capture on target node.." 
    nsenter -t 1 -n tcpdump $TCPDUMPFILTERS -w - | tee /tmp/dumpy/${CAPTURE_NAME}-${TARGET_NODE}.pcap | tcpdump -r - &
    TCPDUMP_PID=$!        
else
    # GET CONTAINERD CLI BINARY PATH FUNCTION
    check_containerd_bin() {
        commands=("crictl" "nerdctl" "ctr")
        for binary in "${commands[@]}"
        do
            if [[ $(nsenter -t 1 -m -n ls /usr/local/bin | grep $binary) != "" ]]; then
                echo "/usr/local/bin/$binary"
                return 0
            elif [[ $(nsenter -t 1 -m -n ls /usr/bin | grep $binary) != "" ]]; then
                echo "/usr/bin/$binary"
                return 0
            fi
        done
        echo "containerd cli not found, tried [crictl, nerdctl, ctr]"
        return 1
    }

    if BINPATH=$(check_containerd_bin) 
    then
        echo "$LOG_INFO using bin $BINPATH"
    else
        exit 1
    fi

    BIN=$(echo $BINPATH | awk -F '/' '{print $NF}')
    # GET TARGET CONTAINER PID FROM BIN INSPECT
    if [[ $BIN == "crictl" ]]
    then
        TARGET_PID=$(nsenter -t 1 -m -n $BINPATH inspect $TARGET_CONTAINERID | jq .info.pid)
    elif [[ $BIN == "nerdctl" ]]
    then
        TARGET_PID=$(nsenter -t 1 -m -n $BINPATH  inspect $TARGET_CONTAINERID | jq .[0].State.Pid)
    elif [[ $BIN == "ctr" ]]
    then
        TARGET_PID=$(nsenter -t 1 -m -n $BINPATH container info $TARGET_CONTAINERID --format=json | jq -r '.info.State.Pid')
    fi

    if [[ $TARGET_PID = "" ]]
    then
        echo "$LOG_WARN No PID found for target container"
        exit 1
    fi

    printf "\n$LOG_INFO target container PID : $TARGET_PID\n"

    # RUN TCPDUMP IN PID NAMESPACE
    echo "$LOG_INFO starting capture on target pod.." 
    nsenter -t $TARGET_PID -n tcpdump $TCPDUMPFILTERS -w - | tee /tmp/dumpy/${CAPTURE_NAME}-${TARGET_POD}.pcap | tcpdump -r - &
    TCPDUMP_PID=$!
fi

if [[ $TCPDUMP_PID != "" ]]
then
    echo "$LOG_INFO Dumpy sniffer PID: $TCPDUMP_PID"
else
    echo $LOG_ERR Dumpy sniffer failed to run Tcpdump on target, terminating..
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
            echo "network Capture failed to initiate, terminating sniffer..."
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