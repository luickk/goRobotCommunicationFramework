sleep 1
go run ../examples/node.go &
P1=$!

sleep 1
go run ../examples/publisher_glob.go altsensglob &
P2=$!

sleep 1
go run ../examples/worker_glob.go altsensglob &
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

sleep 1
go run ../examples/single_topic_string_worker.go &
P8=$!

sleep 1
go run ../examples/single_topic_glob_worker.go &
P8=$!

wait $P1 $P2 $P3 $P4 $P5 $P6 $P7 $P8 $P9