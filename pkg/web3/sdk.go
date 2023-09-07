package web3

import (
	"crypto/ecdsa"
	"math/big"
	"strconv"
	"strings"

	"github.com/bacalhau-project/lilypad/pkg/web3/bindings/controller"
	"github.com/bacalhau-project/lilypad/pkg/web3/bindings/payments"
	"github.com/bacalhau-project/lilypad/pkg/web3/bindings/storage"
	"github.com/bacalhau-project/lilypad/pkg/web3/bindings/token"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
)

// these are the go-binding wrappers for the various deployed contracts
type Contracts struct {
	Token      *token.Token
	Payments   *payments.Payments
	Storage    *storage.Storage
	Controller *controller.Controller
}

type ContractSDK struct {
	Options    Web3Options
	PrivateKey *ecdsa.PrivateKey
	Client     *ethclient.Client
	Auth       *bind.TransactOpts
	Contracts  *Contracts
}

func NewContracts(options Web3Options, client *ethclient.Client) (*Contracts, error) {
	token, err := token.NewToken(common.HexToAddress(options.TokenAddress), client)
	if err != nil {
		return nil, err
	}
	payments, err := payments.NewPayments(common.HexToAddress(options.PaymentsAddress), client)
	if err != nil {
		return nil, err
	}
	storage, err := storage.NewStorage(common.HexToAddress(options.StorageAddress), client)
	if err != nil {
		return nil, err
	}
	controller, err := controller.NewController(common.HexToAddress(options.ControllerAddress), client)
	if err != nil {
		return nil, err
	}
	return &Contracts{
		Token:      token,
		Payments:   payments,
		Storage:    storage,
		Controller: controller,
	}, nil
}

func NewContractSDK(options Web3Options) (*ContractSDK, error) {
	optionsErr := checkOptions(options)
	if optionsErr != nil {
		return nil, optionsErr
	}
	client, err := ethclient.Dial(options.RpcURL)
	if err != nil {
		return nil, err
	}
	privateKey, err := crypto.HexToECDSA(strings.Replace(options.PrivateKey, "0x", "", 1))
	if err != nil {
		return nil, err
	}
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(int64(options.ChainID)))
	if err != nil {
		return nil, err
	}
	contracts, err := NewContracts(options, client)
	if err != nil {
		return nil, err
	}
	return &ContractSDK{
		PrivateKey: privateKey,
		Options:    options,
		Client:     client,
		Auth:       auth,
		Contracts:  contracts,
	}, nil
}

func (sdk *ContractSDK) getBlockNumber() (uint64, error) {
	var blockNumberHex string
	err := sdk.Client.Client().Call(&blockNumberHex, "eth_blockNumber")
	if err != nil {
		log.Error().Msgf("error for getBlockNumber: %s", err.Error())
		return 0, err
	}
	blockNumberHex = strings.TrimPrefix(blockNumberHex, "0x")
	return strconv.ParseUint(blockNumberHex, 16, 64)
}

// func NewContracts(options ContractOptions) (Contract, error) {
// 	if options.Address == "" {
// 		return nil, fmt.Errorf("contract address option must be set")
// 	}

// 	if options.PrivateKey == "" {
// 		return nil, fmt.Errorf("contract private key option must be set")
// 	}

// 	if options.RPCEndpoint == "" {
// 		return nil, fmt.Errorf("contract rpc endpoint option must be set")
// 	}

// 	if options.ChainID == "" {
// 		return nil, fmt.Errorf("contract chain id option must be set")
// 	}

// 	address := common.HexToAddress(options.Address)
// 	privateKey, err := crypto.HexToECDSA(strings.Replace(options.PrivateKey, "0x", "", 1))
// 	if err != nil {
// 		return nil, err
// 	}

// 	log.Debug().
// 		Str("endpoint", options.RPCEndpoint).
// 		Str("chainID", options.ChainID).
// 		Str("address", options.Address).
// 		Msg("Dial")
// 	client, err := ethclient.Dial(options.RPCEndpoint)
// 	if err != nil {
// 		return nil, err
// 	}

// 	chainId, err := strconv.ParseInt(options.ChainID, 10, 32)
// 	if err != nil {
// 		return nil, err
// 	}

// 	contract, err := NewModicum(address, client)
// 	if err != nil {
// 		return nil, err
// 	}

// 	number, err := client.BlockNumber(context.Background())
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &contractWrapper{
// 		client:       client,
// 		contract:     contract,
// 		chainID:      big.NewInt(chainId),
// 		privateKey:   privateKey,
// 		maxSeenBlock: number,
// 	}, nil
// }

// func (r *realContract) GetImageIDs(
// 	ctx context.Context,
// ) ([]int, error) {
// 	ids, err := r.contract.ArtistAttributionCaller.GetImageIDs(&bind.CallOpts{
// 		Context: ctx,
// 	})
// 	if err != nil {
// 		return nil, err
// 	}
// 	ret := []int{}
// 	for _, num := range ids {
// 		ret = append(ret, int(num.Int64()))
// 	}
// 	return ret, nil
// }

// func (r *realContract) GetArtistIDs(
// 	ctx context.Context,
// ) ([]string, error) {
// 	return r.contract.ArtistAttributionCaller.GetArtistIDs(&bind.CallOpts{
// 		Context: ctx,
// 	})
// }

// func (r *realContract) GetImage(
// 	ctx context.Context,
// 	id int,
// ) (ArtistAttributionImage, error) {
// 	return r.contract.ArtistAttributionCaller.GetImage(&bind.CallOpts{
// 		Context: ctx,
// 	}, big.NewInt(int64(id)))
// }

// // Complete implements SmartContract
// func (r *realContract) ArtistComplete(ctx context.Context, id string) error {
// 	opts, err := r.prepareTransaction(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	txn, err := r.contract.ArtistAttributionTransactor.ArtistComplete(
// 		opts,
// 		id,
// 	)
// 	if err != nil {
// 		return err
// 	}

// 	log.Ctx(ctx).Info().Stringer("txn", txn.Hash()).Msgf("ArtistComplete: %d", id)
// 	return nil
// }

// // Refund implements SmartContract
// func (r *realContract) ArtistCancelled(ctx context.Context, id string) error {
// 	opts, err := r.prepareTransaction(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	txn, err := r.contract.ArtistAttributionTransactor.ArtistCancelled(
// 		opts,
// 		id,
// 	)
// 	if err != nil {
// 		return err
// 	}

// 	log.Ctx(ctx).Info().Stringer("txn", txn.Hash()).Msgf("ArtistCancelled: %s", id)
// 	return nil
// }

// func (r *realContract) ImageComplete(ctx context.Context, id int, result string) error {
// 	opts, err := r.prepareTransaction(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	txn, err := r.contract.ArtistAttributionTransactor.ImageComplete(
// 		opts,
// 		big.NewInt(int64(id)),
// 		result,
// 	)
// 	if err != nil {
// 		return err
// 	}

// 	log.Ctx(ctx).Info().Stringer("txn", txn.Hash()).Msgf("ImageComplete: %d", id)
// 	return nil
// }

// // Refund implements SmartContract
// func (r *realContract) ImageCancelled(ctx context.Context, id int, errorString string) error {
// 	opts, err := r.prepareTransaction(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	txn, err := r.contract.ArtistAttributionTransactor.ImageCancelled(
// 		opts,
// 		big.NewInt(int64(id)),
// 		errorString,
// 	)
// 	if err != nil {
// 		return err
// 	}

// 	log.Ctx(ctx).Info().Stringer("txn", txn.Hash()).Msgf("ImageCancelled: %d %s", id, errorString)
// 	return nil
// }

// func (r *realContract) Listen(
// 	ctx context.Context,
// 	imageChan chan<- *types.ImageCreatedEvent,
// 	artistChan chan<- *types.ArtistCreatedEvent,
// ) error {

// 	t := time.NewTicker(r.tickerTime)
// 	defer t.Stop()

// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return nil
// 		case <-t.C:
// 			err := r.ReadEvents(ctx, imageChan, artistChan)
// 			if err != nil {
// 				return err
// 			}
// 		}
// 	}
// }

// func (r *realContract) ReadEvents(
// 	ctx context.Context,
// 	imageChan chan<- *types.ImageCreatedEvent,
// 	artistChan chan<- *types.ArtistCreatedEvent,
// ) error {
// 	log.Ctx(ctx).Debug().Uint64("fromBlock", r.maxSeenBlock+1).Msg("Polling for smart contract image and artist events")

// 	// We deliberately ask for the current block *before* we make the events
// 	// call. It's possible that a block will be written between the two calls:
// 	//
// 	//    FilterNewJobs(block: #1) -> seen block #1
// 	//    block #2 gets written
// 	//    BlockNumber() -> block #3
// 	//    ...
// 	//    FilterNewJobs(block: #3)
// 	//
// 	// In this case we would never see any events in block #2. So we instead
// 	// remember the block number before the events call, and if a block is
// 	// written between them, we will get it again next time we ask for events.
// 	currentBlock, err := r.client.BlockNumber(ctx)
// 	if err != nil {
// 		log.Ctx(ctx).Error().Err(err).Send()
// 		return err
// 	}

// 	opts := bind.FilterOpts{Start: uint64(r.maxSeenBlock + 1), Context: ctx}

// 	imageLogs, err := r.contract.ArtistAttributionFilterer.FilterEventImageCreated(&opts)
// 	if err != nil {
// 		log.Ctx(ctx).Error().Err(err).Send()
// 		return err
// 	}
// 	defer imageLogs.Close()

// 	artistLogs, err := r.contract.ArtistAttributionFilterer.FilterEventArtistCreated(&opts)
// 	if err != nil {
// 		log.Ctx(ctx).Error().Err(err).Send()
// 		return err
// 	}
// 	defer artistLogs.Close()

// 	r.maxSeenBlock = currentBlock

// 	for imageLogs.Next() {
// 		recvEvent := imageLogs.Event
// 		// IMPORTANT: this means the log was reverted, so we should ignore it
// 		if recvEvent.Raw.Removed {
// 			continue
// 		}
// 		log.Ctx(ctx).Info().
// 			Stringer("txn", recvEvent.Raw.TxHash).
// 			Uint64("block#", recvEvent.Raw.BlockNumber).
// 			Uint64("id", recvEvent.Raw.BlockNumber).
// 			Str("artist", recvEvent.Image.Artist).
// 			Str("prompt", recvEvent.Image.Prompt).
// 			Msg("Image")
// 		imageChan <- &types.ImageCreatedEvent{
// 			ContractID: int(recvEvent.Image.Id.Int64()),
// 			ArtistCode: recvEvent.Image.Artist,
// 			Prompt:     recvEvent.Image.Prompt,
// 		}
// 	}

// 	for artistLogs.Next() {
// 		recvEvent := artistLogs.Event
// 		// IMPORTANT: this means the log was reverted, so we should ignore it
// 		if recvEvent.Raw.Removed {
// 			continue
// 		}
// 		log.Ctx(ctx).Info().
// 			Stringer("txn", recvEvent.Raw.TxHash).
// 			Uint64("block#", recvEvent.Raw.BlockNumber).
// 			Str("artist", recvEvent.Artist.Id).
// 			Msg("Artist")
// 		artistChan <- &types.ArtistCreatedEvent{
// 			ArtistCode: recvEvent.Artist.Id,
// 		}
// 	}

// 	return nil
// }

// func (r *contractWrapper) publicKey() *ecdsa.PublicKey {
// 	return r.privateKey.Public().(*ecdsa.PublicKey)
// }

// func (r *contractWrapper) wallet() common.Address {
// 	return crypto.PubkeyToAddress(*r.publicKey())
// }

// func (r *contractWrapper) pendingNonce(ctx context.Context) (uint64, error) {
// 	return r.client.PendingNonceAt(ctx, r.wallet())
// }

// func (r *contractWrapper) prepareTransaction(ctx context.Context) (*bind.TransactOpts, error) {
// 	nonce, err := r.pendingNonce(ctx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	opts, err := bind.NewKeyedTransactorWithChainID(r.privateKey, r.chainID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	opts.Nonce = big.NewInt(int64(nonce))
// 	opts.Value = big.NewInt(0)
// 	opts.Context = ctx

// 	return opts, nil
// }