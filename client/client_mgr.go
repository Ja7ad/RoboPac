package client

import (
	"errors"

	"github.com/kehiy/RoboPac/log"
	pactus "github.com/pactus-project/pactus/www/grpc/gen/go"
)

type Mgr struct {
	clients []IClient
}

func NewClientMgr() *Mgr {
	return &Mgr{
		clients: make([]IClient, 0),
	}
}

func (cm *Mgr) AddClient(c IClient) {
	cm.clients = append(cm.clients, c)
}

// NOTE: local client is always the first client.
func (cm *Mgr) getLocalClient() IClient {
	return cm.clients[0]
}

func (cm *Mgr) GetRandomClient() IClient {
	for _, c := range cm.clients {
		return c
	}

	return nil
}

func (cm *Mgr) GetBlockchainInfo() (*pactus.GetBlockchainInfoResponse, error) {
	localClient := cm.getLocalClient()
	info, err := localClient.GetBlockchainInfo()
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (cm *Mgr) GetBlockchainHeight() (uint32, error) {
	localClient := cm.getLocalClient()
	height, err := localClient.GetBlockchainHeight()
	if err != nil {
		return 0, err
	}
	return height, nil
}

func (cm *Mgr) GetLastBlockTime() (uint32, uint32) {
	localClient := cm.getLocalClient()
	lastBlockTime, lastBlockHeight, err := localClient.LastBlockTime()
	if err != nil {
		return 0, 0
	}

	return lastBlockTime, lastBlockHeight
}

func (cm *Mgr) GetNetworkInfo() (*pactus.GetNetworkInfoResponse, error) {
	for _, c := range cm.clients {
		info, err := c.GetNetworkInfo()
		if err != nil {
			continue
		}
		return info, nil
	}

	return nil, errors.New("unable to get network info")
}

func (cm *Mgr) GetPeerInfoFirstVal(address string) (*pactus.PeerInfo, error) {
	for _, c := range cm.clients {
		networkInfo, err := c.GetNetworkInfo()
		if err != nil {
			continue
		}

		if networkInfo != nil {
			for _, p := range networkInfo.ConnectedPeers {
				for i, addr := range p.ConsensusAddress {
					if addr == address {
						if i != 0 {
							return nil, errors.New("please enter the first validator address")
						}
						return p, nil
					}
				}
			}
		}
	}

	return nil, errors.New("peer does not exist")
}

func (cm *Mgr) GetPeerInfo(address string) (*pactus.PeerInfo, error) {
	for _, c := range cm.clients {
		networkInfo, err := c.GetNetworkInfo()
		if err != nil {
			continue
		}

		if networkInfo != nil {
			for _, p := range networkInfo.ConnectedPeers {
				for _, addr := range p.ConsensusAddress {
					if addr == address {
						return p, nil
					}
				}
			}
		}
	}

	return nil, errors.New("peer does not exist")
}

func (cm *Mgr) GetValidatorInfo(address string) (*pactus.GetValidatorResponse, error) {
	localClient := cm.getLocalClient()
	val, err := localClient.GetValidatorInfo(address)
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (cm *Mgr) GetValidatorInfoByNumber(num int32) (*pactus.GetValidatorResponse, error) {
	localClient := cm.getLocalClient()
	val, err := localClient.GetValidatorInfoByNumber(num)
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (cm *Mgr) GetTransactionData(txID string) (*pactus.GetTransactionResponse, error) {
	localClient := cm.getLocalClient()
	txData, err := localClient.GetTransactionData(txID)
	if err != nil {
		return nil, err
	}
	return txData, nil
}

func (cm *Mgr) Stop() {
	for addr, c := range cm.clients {
		if err := c.Close(); err != nil {
			log.Error("could not close connection to RPC node", "err", err, "RPCAddr", addr)
		}
	}
}
