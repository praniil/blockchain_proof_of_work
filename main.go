package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
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

type Message struct {
	Data int
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
		Addr:           ":" + http_port,
		Handler:        mux,
		ReadTimeout:    100 * time.Second,
		WriteTimeout:   100 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	err := s.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}

func makeMuxRouter() http.Handler {
	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/getBlock", handleGetBlockchain).Methods("GET")
	muxRouter.HandleFunc("/block", handleWriteBlock).Methods("POST")
	return muxRouter
}

func handleGetBlockchain(w http.ResponseWriter, r *http.Request) {
	bytes, err := json.MarshalIndent(BlockChain, "", "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.WriteString(w, string(bytes))
}
func handleWriteBlock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	var m Message

	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		log.Fatal("Unable to decode the request body: ", err)
	}
	defer r.Body.Close()

	mutex.Lock()
	newBlock := generateBlock(BlockChain[len(BlockChain)-1], m.Data)
	defer mutex.Unlock()

	if isBlockValid(newBlock, BlockChain[len(BlockChain)-1]) {
		BlockChain = append(BlockChain, newBlock)
		spew.Dump(BlockChain)
	}

	respondWithJson(w, http.StatusCreated, newBlock)
	fmt.Println("content of m: ", m)
}

func respondWithJson(w http.ResponseWriter, code int, payload Block) {
	w.Header().Set("Content-Type", "application/json")
	response, err := json.MarshalIndent(payload, "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("HTTP 500: Internal Server Error"))
		return
	}
	w.WriteHeader(code)
	w.Write(response)
}

func isBlockValid(newBlock Block, oldBlock Block) bool {
	if newBlock.PrevHash != oldBlock.Hash {
		return false
	}
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}
	if calculateHash(newBlock) != newBlock.Hash {
		return false
	}
	return true
}

func generateBlock(oldBlock Block, Data int) Block {
	var newBlock Block

	newBlock.Data = Data
	newBlock.Difficulty = difficulty
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Index = oldBlock.Index + 1
	newBlock.TimeStamp = time.Now().String()

	for i := 0; ; i++ {
		hex := fmt.Sprintf("%x", i)
		newBlock.Nonce = hex
		if !isHashValid(calculateHash(newBlock), newBlock.Difficulty) {
			fmt.Println(calculateHash(newBlock), "do more work")
			time.Sleep(time.Second)
			continue
		} else {
			fmt.Println(calculateHash(newBlock), "work done!")
			newBlock.Hash = calculateHash(newBlock)
			break
		}
	}
	return newBlock
}

func isHashValid(hash string, difficulty int) bool {
	prefix := strings.Repeat("0", difficulty)
	return strings.HasPrefix(hash, prefix)
}

func calculateHash(anyBlock Block) string {
	blockMetadata := strconv.Itoa(anyBlock.Index) + anyBlock.TimeStamp + strconv.Itoa(anyBlock.Data) + anyBlock.PrevHash + anyBlock.Nonce
	hashed := sha256.Sum256([]byte(blockMetadata))
	return hex.EncodeToString(hashed[:])
}
