rm test_client;
rm test_node;
go build test_node.go;
go build test_client.go;
./test_node;
