package wallet

import (
	"log"
	"os"
	"pactus-faucet/config"

	"github.com/k0kubun/pp"
	"github.com/pactus-project/pactus/crypto"
	"github.com/pactus-project/pactus/crypto/bls"
	"github.com/pactus-project/pactus/genesis"
	"github.com/pactus-project/pactus/util"
	pwallet "github.com/pactus-project/pactus/wallet"
)

const entropy = 128
const faucetAddressLabel = "faucet"

type Balance struct {
	Available float64
	Staked    float64
}

type Wallet struct {
	address  string
	wallet   *pwallet.Wallet
	password string
}

func Create(cfg *config.Config, mnemonic string) *Wallet {
	network := genesis.Testnet
	if mnemonic == "" {
		mnemonic = pwallet.GenerateMnemonic(entropy)
	}
	mywallet, err := pwallet.Create(cfg.WalletPath, mnemonic, cfg.WalletPassword, network)

	if err != nil {
		log.Printf("error creating wallet: %v", err)
		return nil
	}
	address, err := mywallet.DeriveNewAddress(faucetAddressLabel)
	if err != nil {
		log.Printf("error deriving wallet faucet address: %v", err)
		return nil
	}
	cfg.FaucetAddress = address
	err = cfg.Save()
	if err != nil {
		log.Printf("error updating configuration faucet address: %v", err)
		return nil
	}
	err = mywallet.Save()
	if err != nil {
		log.Printf("error saving wallet: %v", err)
		return nil
	}
	pp.Printf("Wallet created successfully at: %s\n", mywallet.Path())
	pp.Printf("Seed: \"%v\"\n", mnemonic)
	pp.Printf("Please keep your seed in a safe place;\nif you lose it, you will not be able to restore your wallet.\n")
	return &Wallet{wallet: mywallet, address: cfg.FaucetAddress, password: cfg.WalletPassword}
}

func Open(cfg *config.Config) *Wallet {
	if doesWalletExist(cfg.WalletPath) {
		wt, err := pwallet.Open(cfg.WalletPath, true)
		if err != nil {
			log.Printf("error opening exising wallet: %v", err)
			return nil
		}
		err = wt.Connect(cfg.Server)
		if err != nil {
			log.Printf("error establishing connection: %v", err)
			return nil
		}
		return &Wallet{wallet: wt, address: cfg.FaucetAddress, password: cfg.WalletPassword}
	}
	//if the wallet does not exist, create one
	return Create(cfg, "")
}

func (w *Wallet) BondTransaction(pubKey, toAddress string, amount float64) string {
	opts := []pwallet.TxOption{
		pwallet.OptionStamp(""),
		pwallet.OptionFee(util.CoinToChange(0)),
		pwallet.OptionSequence(int32(0)),
		pwallet.OptionMemo(""),
	}
	tx, err := w.wallet.MakeBondTx(w.address, toAddress, pubKey,
		util.CoinToChange(amount), opts...)
	if err != nil {
		log.Printf("error creating bond transaction: %v", err)
		return ""
	}
	//sign transaction
	err = w.wallet.SignTransaction(w.password, tx)
	if err != nil {
		log.Printf("error signing bond transaction: %v", err)
		return ""
	}

	//broadcast transaction
	res, err := w.wallet.BroadcastTransaction(tx)
	if err != nil {
		log.Printf("error broadcasting bond transaction: %v", err)
		return ""
	}

	err = w.wallet.Save()
	if err != nil {
		log.Printf("error saving wallet transaction history: %v", err)
	}
	return res //return transaction hash
}

func (w *Wallet) GetBalance() *Balance {
	balance := &Balance{Available: 0, Staked: 0}
	b, err := w.wallet.Balance(w.address)
	if err != nil {
		log.Printf("error getting balance: %v", err)
		return balance
	}
	balance.Available = util.ChangeToCoin(b)
	stake, err := w.wallet.Stake(w.address)
	if err != nil {
		log.Printf("error getting staking amount: %v", err)
		return balance
	}
	balance.Staked = util.ChangeToCoin(stake)
	return balance
}

func IsValidData(address, pubKey string) bool {
	addr, err := crypto.AddressFromString(address)
	if err != nil {
		return false
	}
	pub, err := bls.PublicKeyFromString(pubKey)
	if err != nil {
		return false
	}
	err = pub.VerifyAddress(addr)
	return err == nil
}

// function to check if file exists
func doesWalletExist(fileName string) bool {
	_, error := os.Stat(fileName)
	if os.IsNotExist(error) {
		return false
	} else {
		return true
	}
}