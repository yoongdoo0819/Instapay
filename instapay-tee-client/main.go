package main

/*
#cgo CPPFLAGS: -I/home/yoongdoo0819/sgxsdk/include -I/home/yoongdoo0819/instapay3.0/instapay/src/github.com/sslab-instapay/instapay-tee-client
#cgo LDFLAGS: -L/home/yoongdoo0819/instapay3.0/instapay/src/github.com/sslab-instapay/instapay-tee-client -ltee

#include "app.h"
*/
import "C"

import (
	"context"
	"crypto/ecdsa"
	"flag"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"strconv"
	"unsafe"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/sslab-instapay/instapay-tee-client/config"
	instapayGrpc "github.com/sslab-instapay/instapay-tee-client/grpc"
	clientPb "github.com/sslab-instapay/instapay-tee-client/proto/client"
	"github.com/sslab-instapay/instapay-tee-client/router"
	"github.com/sslab-instapay/instapay-tee-client/service"
	"github.com/sslab-instapay/instapay-tee-client/util"
	"google.golang.org/grpc"
	clientGrpc "github.com/sslab-instapay/instapay-tee-client/grpc"

)

var numOfChannels = 3000

func main() {
	C.initialize_enclave()

	portNum := flag.String("port", "3001", "port number")
	grpcPortNum := flag.String("grpc_port", "50001", "grpc_port number")
	peerFileDirectory := flag.String("peer_file_directory", "data/peer/peer.json", "dir")
	keyFile := flag.String("key_file", "./data/key/k0", "key file")
	channelFile := flag.String("channel_file", "./data/channel/c0", "channel file")

	previous_sender := flag.String("previous_sender", "", "previous_sender")

	sender := flag.String("sender", "f55ba9376db959fab2af86d565325829b08ea3c4", "sender")
	receiver := flag.String("receiver", "60f640c4505d15b972e6fc2a2a7cba09d05d9f7", "receiver")
	fmt.Println(sender, receiver, previous_sender)

	flag.Parse()

	os.Setenv("port", *portNum)
	os.Setenv("grpc_port", *grpcPortNum)
	os.Setenv("peer_file_directory", *peerFileDirectory)
	os.Setenv("key_file", *keyFile)
	os.Setenv("channel_file", *channelFile)


	fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~~~");
	LoadPeerInformation(os.Getenv("peer_file_directory"))
	// CreateAccount(os.Getenv("peer_file_directory"))

	fmt.Println("======= test func w start ==========\n");
	C.ecall_test_func_w()
	fmt.Println("======= test func w end   ==========\n");
	
	if fileExists(*keyFile) {
		LoadAccount(os.Getenv("key_file"))
	} else {
		CreateAccount(os.Getenv("key_file"))
	}
	
	//CreateAccount(os.Getenv("key_file"))

	if fileExists(*channelFile) {
		LoadChannelData(*channelFile)
	}

	LoadDataToTEE(os.Getenv("key_file"), os.Getenv("channel_file"))

	go service.ListenContractEvent()
	go startGrpcServer()
/*	if *portNum == "3001" {
		createPaymentChannelsForAlice(*sender, *receiver)
	} else if *portNum == "3002" {
		createPaymentChannelsPrev(*previous_sender, *sender)
		createPaymentChannelsForBob(*sender, *receiver)
	} else if *portNum == "3003" {
		createPaymentChannelsForCarol(*sender, *receiver)
	}
*/
	startClientWebServer()
}

func startGrpcServer() {

	clientGrpc.ChanCreate()
	log.Println("---Start Grpc Server---")
	grpcPort, err := strconv.Atoi(os.Getenv("grpc_port"))
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	clientPb.RegisterClientServer(grpcServer, &instapayGrpc.ClientGrpc{})
	grpcServer.Serve(lis)
}

func startClientWebServer() {
	defaultRouter := gin.Default()
	defaultRouter.LoadHTMLGlob("templates/*")

	defaultRouter.Use(CORSMiddleware())
	router.RegisterRestRouter(defaultRouter)
	router.RegisterViewRouter(defaultRouter)

	defaultRouter.Run(":" + os.Getenv("port"))
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func CreateAccount(directory string) {
	
	fmt.Println("Create Account \n\n");
	var kf *C.char
	kf = C.CString(directory)

	C.ecall_create_account_w()
	C.ecall_store_account_data_w(kf)
	defer C.free(unsafe.Pointer(kf))

	var paddrs unsafe.Pointer

	paddrs = C.ecall_get_public_addrs_w()
	paddrSize := 20
	paddrSlice := (*[1 << 30]C.address)(unsafe.Pointer(paddrs))[:paddrSize:paddrSize]

	var convertedAddress string
	convertedAddress = fmt.Sprintf("%02x", paddrSlice[0].addr)
	convertedAddress = "0x" + convertedAddress
	fmt.Println("---- Public Key Address ---")
	fmt.Println(convertedAddress)
	sendEther(convertedAddress)
	config.SetAccountConfig(convertedAddress)
}

func LoadChannelData(channelFile string) {
	cf := C.CString(channelFile)

	C.ecall_load_channel_data_w(cf)
	defer C.free(unsafe.Pointer(cf))
}

func LoadAccount(keyFile string) {
	
	fmt.Println("Load Account \n\n");
	kf := C.CString(keyFile)
	C.ecall_load_account_data_w(kf)
	defer C.free(unsafe.Pointer(kf))
	var paddrs unsafe.Pointer

	paddrs = C.ecall_get_public_addrs_w()
	paddrSize := 20
	paddrSlice := (*[1 << 30]C.address)(unsafe.Pointer(paddrs))[:paddrSize:paddrSize]

	var convertedAddress string
	convertedAddress = fmt.Sprintf("%02x", paddrSlice[0].addr)
	convertedAddress = "0x" + convertedAddress
	fmt.Println("---- Public Key Address ---")
	fmt.Println(convertedAddress)
	sendEther(convertedAddress)
	config.SetAccountConfig(convertedAddress)
}

func LoadDataToTEE(keyFile string, channelFile string) {

	fmt.Println("Load Data To TEE \n\n");
	kf := C.CString(keyFile)
	cf := C.CString(channelFile)

	C.ecall_load_account_data_w(kf)
	C.ecall_load_channel_data_w(cf)
	defer C.free(unsafe.Pointer(kf))
	defer C.free(unsafe.Pointer(cf))

	var paddrs unsafe.Pointer

	paddrs = C.ecall_get_public_addrs_w()
	paddrSize := 20
	paddrSlice := (*[1 << 30]C.address)(unsafe.Pointer(paddrs))[:paddrSize:paddrSize]

	var convertedAddress string
	convertedAddress = fmt.Sprintf("%02x", paddrSlice[0].addr)
	convertedAddress = "0x" + convertedAddress
	fmt.Println("---- Public Key Address ---")
	fmt.Println(convertedAddress)
	config.SetAccountConfig(convertedAddress)
}

func LoadPeerInformation(directory string) {
	util.SetPeerInformation(directory)
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// 임시로 이더를 전송해줌.
func sendEther(hexAddress string) {
	client, err := ethclient.Dial("ws://" + config.EthereumConfig["wsHost"] + ":" + config.EthereumConfig["wsPort"])
	if err != nil {
		log.Fatal(err)
	}

	privateKey, err := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19")
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	fmt.Println("fromAddress :", fromAddress.Hex())
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}

	value := big.NewInt(1110000000000000000) // in wei (1 eth)
	gasLimit := uint64(21000)                // in units
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	toAddress := common.HexToAddress(hexAddress)
	fmt.Println("toAddress : %s", toAddress.Hex());
	var data []byte
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}

	/*
	msg, err := tx.AsMessage(types.NewEIP155Signer(tx.ChainId()))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Sender : ", msg.From().Hex())
	*/

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("tx sent: %s", signedTx.Hash().Hex())
}

func createPaymentChannelsPrev(sender string, receiver string) {

	_sender := []C.uchar(sender)
	_receiver := []C.uchar(receiver)
	deposit := C.uint(50000)

	fmt.Println("start")
	for i:=1; i<=numOfChannels; i+=2 {
		C.ecall_receive_create_channel_w(C.uint(uint32(i)), &_sender[0], &_receiver[0], deposit)
	}

	fmt.Println(">> ", sender)
	fmt.Println(">> ", receiver)

	fmt.Println("Prev Payment Channels Created !")
}

func createPaymentChannelsForAlice(sender string, receiver string) {

	_sender := []C.uchar(sender)
	_receiver := []C.uchar(receiver)
	deposit := C.uint(50000)

	fmt.Println("start")
	for i:=1; i<=numOfChannels; i+=2 {
		C.ecall_receive_create_channel_w(C.uint(uint32(i)), &_sender[0], &_receiver[0], deposit)
	}

	fmt.Println(">> ", sender)
	fmt.Println(">> ", receiver)

	fmt.Println("Payment Channels Created !")
}


func createPaymentChannelsForBob(sender string, receiver string) {

	_sender := []C.uchar(sender)
	_receiver := []C.uchar(receiver)
	deposit := C.uint(50000)

	fmt.Println("start")
	for i:=2; i<=numOfChannels; i+=2 {
		C.ecall_receive_create_channel_w(C.uint(uint32(i)), &_sender[0], &_receiver[0], deposit)
	}

	fmt.Println(">> ", sender)
	fmt.Println(">> ", receiver)

	fmt.Println("Payment Channels Created !")
}

func createPaymentChannelsForCarol(sender string, receiver string) {

	_sender := []C.uchar(sender)
	_receiver := []C.uchar(receiver)
	deposit := C.uint(50000)

	fmt.Println("start")
	for i:=2; i<=numOfChannels; i+=2 {
		C.ecall_receive_create_channel_w(C.uint(uint32(i)), &_sender[0], &_receiver[0], deposit)
	}

	fmt.Println(">> ", sender)
	fmt.Println(">> ", receiver)

	fmt.Println("Payment Channels Created !")
}
