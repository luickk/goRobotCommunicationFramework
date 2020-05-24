go run ../examples/node.go &
P1=$!


sleep 1
go run ../examples/publisher_glob.go &
P2=$!

sleep 1
go run ../examples/worker_glob.go &
P3=$!

sleep 1
go run ../examples/publisher_string.go &
P4=$!

sleep 1
go run ../examples/worker_string.go &
P5=$!

sleep 1
go run ../examples/multiple_topic_publisher_glob.go &
P6=$!

sleep 1
go run ../examples/multiple_topic_worker_glob.go &
P7=$!

wait $P1 $P2 $P3 $P4 $P5 $P6 $P7