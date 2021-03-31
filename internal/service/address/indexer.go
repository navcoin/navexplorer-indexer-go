package address

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/internal/service/block"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type Indexer struct {
	navcoin *navcoind.Navcoind
	elastic *elastic_cache.Index
	repo    *Repository
}

func NewIndexer(navcoin *navcoind.Navcoind, elastic *elastic_cache.Index, repo *Repository) *Indexer {
	return &Indexer{navcoin, elastic, repo}
}

func (i *Indexer) BulkIndex(height uint64) {
	start := time.Now()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		addressHistorys := i.generateAddressHistory(
			block.BlockData.First().Block.Height,
			block.BlockData.Last().Block.Height,
			block.BlockData.Addresses(),
			block.BlockData.Txs())
		for _, addressHistory := range addressHistorys {
			i.elastic.AddIndexRequest(elastic_cache.AddressHistoryIndex.Get(), addressHistory)

			err := i.updateAddress(addressHistory)
			if err != nil {
				log.WithError(err).Fatalf("Could not update address: %s", addressHistory.Hash)
			}
		}
	}()

	go func() {
		defer wg.Done()
		for _, tx := range block.BlockData.Txs() {
			i.indexMultiSigs([]*explorer.BlockTransaction{tx})
		}
	}()

	wg.Wait()

	elapsed := time.Since(start)
	log.WithFields(log.Fields{
		"time": elapsed,
	}).Infof("Index address:   %d", height)

	block.BlockData.Reset()
}

func (i *Indexer) Index(block *explorer.Block, txs []*explorer.BlockTransaction) {
	if len(txs) == 0 {
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		log.Infof("Index Addresses %d", block.Height)
		i.indexAddresses(block.Height, txs)
	}()

	go func() {
		defer wg.Done()
		log.Infof("Index Multisigs %d", block.Height)
		i.indexMultiSigs(txs)
	}()

	wg.Wait()
}

func (i *Indexer) indexAddresses(height uint64, txs []*explorer.BlockTransaction) {
	addresses := getAddressesForTxs(txs)
	for _, addressHistory := range i.generateAddressHistory(height, height, addresses, txs) {
		i.elastic.AddIndexRequest(elastic_cache.AddressHistoryIndex.Get(), addressHistory)

		err := i.updateAddress(addressHistory)
		if err != nil {
			log.WithError(err).Fatalf("Could not update address: %s", addressHistory.Hash)
		}
	}
}

func (i *Indexer) indexMultiSigs(txs []*explorer.BlockTransaction) {
	multiSigs := getMultiSigsForTxs(txs)

	for _, multiSig := range multiSigs {
		address, err := i.getAddress(multiSig.Key())
		if err != nil {
			log.Fatalf("Failed to get or create address %s", multiSig.Key())
			continue
		}

		if address.MultiSig == nil {
			address.MultiSig = multiSig
		}

		for _, tx := range txs {
			addressHistory := &explorer.AddressHistory{
				Height:      tx.Height,
				TxIndex:     tx.Index,
				TxId:        tx.Txid,
				Time:        tx.Time,
				Hash:        multiSig.Key(),
				CfundPayout: false,
				StakePayout: false,
				Changes: explorer.AddressChanges{
					Spendable:    0,
					Stakable:     0,
					VotingWeight: 0,
				},
				Balance: explorer.AddressBalance{
					Spendable:    address.Spendable,
					Stakable:     address.Stakable,
					VotingWeight: address.VotingWeight,
				},
				Stake: tx.IsAnyStaking(),
			}
			for _, vin := range tx.Vin {
				if vin.PreviousOutput.MultiSig != nil && vin.PreviousOutput.MultiSig.Key() == multiSig.Key() {
					addressHistory.Changes = explorer.AddressChanges{
						Spendable:    addressHistory.Changes.Spendable - int64(vin.ValueSat),
						Stakable:     addressHistory.Changes.Stakable - int64(vin.ValueSat),
						VotingWeight: addressHistory.Changes.VotingWeight - int64(vin.ValueSat),
					}
					addressHistory.Balance = explorer.AddressBalance{
						Spendable:    addressHistory.Balance.Spendable - int64(vin.ValueSat),
						Stakable:     addressHistory.Balance.Stakable - int64(vin.ValueSat),
						VotingWeight: addressHistory.Balance.VotingWeight - int64(vin.ValueSat),
					}
				}
			}
			for _, vout := range tx.Vout {
				if vout.MultiSig != nil && vout.MultiSig.Key() == multiSig.Key() {
					addressHistory.Changes = explorer.AddressChanges{
						Spendable:    addressHistory.Changes.Spendable + int64(vout.ValueSat),
						Stakable:     addressHistory.Changes.Stakable + int64(vout.ValueSat),
						VotingWeight: addressHistory.Changes.VotingWeight + int64(vout.ValueSat),
					}
					addressHistory.Balance = explorer.AddressBalance{
						Spendable:    addressHistory.Balance.Spendable + int64(vout.ValueSat),
						Stakable:     addressHistory.Balance.Stakable + int64(vout.ValueSat),
						VotingWeight: addressHistory.Balance.VotingWeight + int64(vout.ValueSat),
					}
				}
			}
			i.elastic.AddIndexRequest(elastic_cache.AddressHistoryIndex.Get(), addressHistory)

			err := i.updateAddress(addressHistory)
			if err != nil {
				log.WithError(err).Fatalf("Could not update address: %s", addressHistory.Hash)
			}
		}
	}
}

func getMultiSigsForTxs(txs []*explorer.BlockTransaction) []*explorer.MultiSig {
	multiSigs := make([]*explorer.MultiSig, 0)
	for _, tx := range txs {
		for _, multiSig := range tx.GetAllMultiSigs() {
			multiSigs = append(multiSigs, multiSig)
		}
	}

	return multiSigs
}

func (i *Indexer) generateAddressHistory(start, end uint64, addresses []string, txs []*explorer.BlockTransaction) []*explorer.AddressHistory {
	addressHistory := make([]*explorer.AddressHistory, 0)

	history, err := i.navcoin.GetAddressHistory(&start, &end, addresses...)
	if err != nil {
		log.WithError(err).Errorf("Could not get address history for height: %d-%d", start, end)
		return addressHistory
	}

	for _, h := range history {
		addressHistory = append(addressHistory, CreateAddressHistory(h, getTxById(h.TxId, txs)))
	}

	return addressHistory
}

func (i *Indexer) getAddress(hash string) (*explorer.Address, error) {
	address := Addresses.GetByHash(hash)
	if address != nil {
		return address, nil
	}

	return i.repo.GetOrCreateAddress(hash)
}

func (i *Indexer) updateAddress(history *explorer.AddressHistory) error {
	address, err := i.getAddress(history.Hash)
	if err != nil {
		return err
	}

	address.Height = history.Height
	address.Spendable = history.Balance.Spendable
	address.Stakable = history.Balance.Stakable
	address.VotingWeight = history.Balance.VotingWeight

	Addresses[address.Hash] = address

	i.elastic.AddUpdateRequest(elastic_cache.AddressIndex.Get(), address)
	return nil
}

func getAddressesForTxs(txs []*explorer.BlockTransaction) []string {
	addresses := make([]string, 0)
	for _, tx := range txs {
		for _, address := range tx.GetAllAddresses() {
			addresses = append(addresses, address)
		}
	}

	return unique(addresses)
}

func unique(stringSlice []string) []string {
	keys := make(map[string]bool)
	var list []string
	for _, entry := range stringSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func getTxById(id string, txs []*explorer.BlockTransaction) *explorer.BlockTransaction {
	for _, tx := range txs {
		if tx.Txid == id {
			return tx
		}
	}
	log.Fatal("Could not match tx: ", id)
	return nil
}
