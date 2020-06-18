#!/bin/bash
# by deltablue
OLD=""
while true
    do
        PID="3796"
        if [[ $PID != $OLD ]]; then
            echo ""
            echo "   New process started."
            echo ""
        fi
        count="$(lsof -n | grep $PID | grep -i tcp | wc -l)"
        echo "$count"
        OLD="$PID"
        sleep 2s
done
