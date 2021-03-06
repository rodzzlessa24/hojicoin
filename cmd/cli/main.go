package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"gitlab.com/rodzzlessa24/hoji"
)

func main() {
	cli := CLI{}
	cli.Run()
}

// CLI responsible for processing command line arguments
type CLI struct{}

func (cli *CLI) createBlockchain(address string) {
	if !hoji.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	if err := hoji.CreateBlockchain([]byte(address)); err != nil {
		log.Panic(err)
	}
	fmt.Println("Done!")
}

func (cli *CLI) getBalance(address string) {
	if !hoji.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc, _ := hoji.NewBlockchain()
	defer bc.DB.Close()

	utxoSet := hoji.UTXOSet{Bc: bc}
	balance := 0
	UTXOs, err := utxoSet.FindUTXO([]byte(address))
	if err != nil {
		log.Panic(err)
	}

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

func (cli *CLI) reindexUTXO() {
	bc, _ := hoji.NewBlockchain()
	UTXOSet := hoji.UTXOSet{
		Bc: bc,
	}
	UTXOSet.Reindex()

	count, err := UTXOSet.CountTransactions()
	if err != nil {
		log.Panic(err)
	}
	fmt.Printf("Done! There are %d transactions in the UTXO set.\n", count)
}

func (cli *CLI) send(from, to string, amount int) {
	if !hoji.ValidateAddress(from) {
		log.Panic("ERROR: from address is not valid")
	}
	if !hoji.ValidateAddress(to) {
		log.Panic("ERROR: to address is not valid")
	}
	bc, _ := hoji.NewBlockchain()
	defer bc.DB.Close()

	tx, err := bc.NewTx([]byte(from), []byte(to), amount)
	if err != nil {
		log.Panic(err)
	}

	coinbaseTx, err := hoji.NewCoinbaseTx([]byte(from), []byte("reward tx"))
	if err != nil {
		log.Panic(err)
	}

	txs := []*hoji.Transaction{tx, coinbaseTx}

	b, err := bc.MineBlock(txs)
	if err != nil {
		log.Panic(err)
	}

	utxoSet := hoji.UTXOSet{Bc: bc}
	if err := utxoSet.Update(b); err != nil {
		log.Panic(err)
	}

	fmt.Println("money sent!")
}

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
	fmt.Println("  createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  reindexutxo - Rebuilds the UTXO set")
	fmt.Println("  createwallet - Generates a new key-pair and saves it into the wallet file")
	fmt.Println("  printchain - Print all the blocks of the blockchain")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT - Send AMOUNT of coins from FROM address to TO")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) createWallet() {
	wallets, err := hoji.NewWallets()
	if err != nil {
		log.Panic("err creating new wallets", err)
	}
	address, err := wallets.AddWallet()
	if err != nil {
		log.Panic(err)
	}
	wallets.SaveToFile()

	fmt.Printf("Your new address: %s\n", address)
}

func (cli *CLI) printChain() {
	// TODO: Fix this
	bc, _ := hoji.NewBlockchain()
	defer bc.DB.Close()

	bci := bc.Iterator()

	for {
		block := bci.Next()

		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := hoji.NewPOW(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CLI) listAddresses() {
	wallets, err := hoji.NewWallets()
	if err != nil {
		log.Panic(err)
	}
	addresses := wallets.GetAddresses()

	for _, address := range addresses {
		fmt.Println(address)
	}
}

// Run parses command line arguments and processes commands
func (cli *CLI) Run() {
	cli.validateArgs()

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	reindexUTXOCmd := flag.NewFlagSet("reindexutxo", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	switch os.Args[1] {
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "reindexutxo":
		err := reindexUTXOCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "listaddresses":
		err := listAddressesCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getBalanceAddress)
	}

	if reindexUTXOCmd.Parsed() {
		cli.reindexUTXO()
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockchain(*createBlockchainAddress)
	}

	if createWalletCmd.Parsed() {
		cli.createWallet()
	}

	if listAddressesCmd.Parsed() {
		cli.listAddresses()
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}

		cli.send(*sendFrom, *sendTo, *sendAmount)
	}
}
