package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

const difficulty = 1

type Block struct {
	Index      int
	TimeStamp  string
	Data       int
	PrevHash   string
	Hash       string
	Difficulty int
	Nonce      string
}

var BlockChain []Block

var mutex = sync.Mutex{}

func main() {
	fmt.Println("Hello world!")
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		t := time.Now()
		genesisBlock := Block{}
		genesisBlock = Block{0, t.String(), 0, "", calculateHash(genesisBlock), difficulty, ""}
		fmt.Println("debug above")
		spew.Dump(genesisBlock)
		fmt.Println("debug below")

		mutex.Lock()
		BlockChain = append(BlockChain, genesisBlock)
		mutex.Unlock()
	}()

	wg.Wait()
	log.Fatal(run())
}

func run() error {
	mux := makeMuxRouter()
	http_port := os.Getenv("PORT")
	log.Println("HTTP server listening on port: ", http_port)
	s := &http.Server{
		Addr: ":" + http_port,
		Handler: mux,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	err := s.ListenAndServe();
	if err != nil {
		return err
	}
	
	return nil
}

func makeMuxRouter() http.Handler {
	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/", handleGetBlockchain).Methods("GET")
	muxRouter.HandleFunc("/", handleWriteBlock).Methods("POST")
	return muxRouter
}

func handleGetBlockchain(w http.ResponseWriter, r *http.Request) {}
func handleWriteBlock(w http.ResponseWriter, r *http.Request)    {
	w.Header().Set("Content-Type", "application/json")
	
}

func calculateHash(anyBlock Block) string {
	blockMetadata := strconv.Itoa(anyBlock.Index) + anyBlock.TimeStamp + strconv.Itoa(anyBlock.Data) + anyBlock.PrevHash + anyBlock.Nonce
	hashed := sha256.Sum256([]byte(blockMetadata))
	return hex.EncodeToString(hashed[:])
}
