source ~/sgxsdk/environment
export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:/home/yoongdoo0819/instapay3.0/instapay/src/github.com/sslab-instapay/instapay-tee-client/
go run main.go -port=3003 -grpc_port=50003 -peer_file_directory=data/peer/carol.json -key_file=./data/key/k2 -channel_file=./data/channel/c2 -sender="c60f640c4505d15b972e6fc2a2a7cba09d05d9f7" -receiver="70603f1189790fcd0fd753a7fef464bdc2c2ad36"
