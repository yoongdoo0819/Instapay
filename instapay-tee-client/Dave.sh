source ~/sgxsdk/environment
export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:/home/yoongdoo0819/instapay3.0/instapay/src/github.com/sslab-instapay/instapay-tee-client
go run main.go -port=3005 -grpc_port=50005 -peer_file_directory=./data/peer/dave.json -key_file=./data/key/k3 -channel_file=./data/channel/c3
