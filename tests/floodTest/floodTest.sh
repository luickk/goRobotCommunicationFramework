sleep 1
go run floodClientPublisher.go alt &
P1=$!

sleep 1
go run floodClientPublisher.go radar &
P2=$!

sleep 1
go run floodClientPublisher.go prox &
P3=$!

sleep 1
go run floodClientPublisher.go amb &
P4=$!

sleep 1
go run floodClientPublisher.go pressure &
P5=$!




sleep 1
go run floodClientReceiver.go alt &
P6=$!

sleep 1
go run floodClientReceiver.go radar &
P7=$!

sleep 1
go run floodClientReceiver.go prox &
P8=$!

sleep 1
go run floodClientReceiver.go amb &
P8=$!

sleep 1
go run floodClientReceiver.go pressure &
P9=$!



# sleep 1
# go run floodClientSubscribe.go amb &
# P10=$!
#
# sleep 1
# go run floodClientSubscribe.go pressure &
# P11=$!




wait $P1 $P2 $P3 $P4 $P5 $P6 $P7 $P8 $P9
