package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"
	"github.com/davecgh/go-spew/spew"
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

	go func() {
		t := time.Now()
		genesisBlock := Block{}
		genesisBlock = Block{0, t.String(), 0, "", calculateHash(genesisBlock), difficulty, ""}
		fmt.Println("debug above");
		spew.Dump(genesisBlock)
		fmt.Println("debug below");

		mutex.Lock()
		BlockChain = append(BlockChain, genesisBlock)
		mutex.Unlock()
	}()
	 
	
}

func calculateHash(anyBlock Block) string {
	blockMetadata := strconv.Itoa(anyBlock.Index) + anyBlock.TimeStamp + strconv.Itoa(anyBlock.Data) + anyBlock.PrevHash + anyBlock.Nonce
	hashed := sha256.Sum256([]byte(blockMetadata))
	return hex.EncodeToString(hashed[:])
}
