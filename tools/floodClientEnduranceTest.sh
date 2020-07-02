#!/bin/bash

sleep 1
go run ../examples/node.go &
P1=$!

sleep 1
go run ../examples/publisher_glob.go a &
P2=$!

sleep 1
go run ../examples/publisher_glob.go a &
P3=$!

sleep 1
go run ../examples/publisher_glob.go a &
P4=$!

sleep 1
go run ../examples/publisher_glob.go a &
P5=$!

sleep 1
go run ../examples/publisher_glob.go a &
P6=$!

sleep 1
go run ../examples/publisher_glob.go b &
P7=$!

sleep 1
go run ../examples/publisher_glob.go b &
P8=$!

sleep 1
go run ../examples/publisher_glob.go b &
P9=$!

sleep 1
go run ../examples/publisher_glob.go b &
P10=$!

sleep 1
go run ../examples/publisher_glob.go b &
P11=$!

sleep 1
go run ../examples/publisher_glob.go c &
P22=$!

sleep 1
go run ../examples/publisher_glob.go c &
P13=$!

sleep 1
go run ../examples/publisher_glob.go c &
P14=$!

sleep 1
go run ../examples/publisher_glob.go c &
P15=$!

sleep 1
go run ../examples/publisher_glob.go c &
P16=$!

sleep 1
go run ../examples/publisher_glob.go d &
P17=$!

sleep 1
go run ../examples/publisher_glob.go d &
P18=$!

sleep 1
go run ../examples/publisher_glob.go d &
P19=$!

sleep 1
go run ../examples/publisher_glob.go d &
P20=$!




sleep 1
go run ../examples/worker_glob.go a &
P21=$!

sleep 1
go run ../examples/worker_glob.go a &
P22=$! 


sleep 1
go run ../examples/worker_glob.go a &
P23=$!


sleep 1
go run ../examples/worker_glob.go a &
P24=$!


sleep 1
go run ../examples/worker_glob.go a &
P25=$!


sleep 1
go run ../examples/worker_glob.go b &
P26=$!


sleep 1
go run ../examples/worker_glob.go b &
P27=$!


sleep 1
go run ../examples/worker_glob.go b &
P28=$!


sleep 1
go run ../examples/worker_glob.go b &
P29=$!


sleep 1
go run ../examples/worker_glob.go b &
P30=$!


sleep 1
go run ../examples/worker_glob.go c &
P31=$!


sleep 1
go run ../examples/worker_glob.go c &
P32=$!


sleep 1
go run ../examples/worker_glob.go c &
P33=$!


sleep 1
go run ../examples/worker_glob.go c &
P34=$!


sleep 1
go run ../examples/worker_glob.go c &
P35=$!


sleep 1
go run ../examples/worker_glob.go d &
P36=$!


sleep 1
go run ../examples/worker_glob.go d &
P37=$!


sleep 1
go run ../examples/worker_glob.go d &
P38=$!


sleep 1
go run ../examples/worker_glob.go d &
P39=$!


sleep 1
go run ../examples/worker_glob.go d &
P40=$!

wait $P1 $P2 $P3 $P4 $P5 $P6 $P7 $P8 $P9 $P10 $P11 $P12 $P13 $P14 $P15 $P16 $P17 $P18 $P19 $P20 $P21 $P22 $P23 $P24 $P25 $26P $P27 $P28 $P29 $P30 $P31 $P32 $P33 $P34 $P35 $P56 $P37 $P38 $P39 $P40