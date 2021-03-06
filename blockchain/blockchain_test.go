// Copyright (c) 2018 IoTeX
// This is an alpha (internal) release and is not suitable for production. This source code is provided 'as is' and no
// warranties are given as to title or non-infringement, merchantability or fitness for purpose and, to the extent
// permitted by law, all liability for your use of the code is disclaimed. This source code is governed by Apache
// License 2.0 that can be found in the LICENSE file.

package blockchain

import (
	"fmt"
	"math/big"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iotexproject/iotex-core/blockchain/action"
	"github.com/iotexproject/iotex-core/common"
	"github.com/iotexproject/iotex-core/config"
	"github.com/iotexproject/iotex-core/state"
	ta "github.com/iotexproject/iotex-core/test/testaddress"
	"github.com/iotexproject/iotex-core/test/util"
	"github.com/iotexproject/iotex-core/trie"
)

const (
	testingConfigPath = "../config.yaml"
	testDBPath        = "db.test"
	testCoinbaseData  = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"
	testTriePath      = "trie.test"
)

func addTestingTsfBlocks(bc Blockchain) error {
	// Add block 1
	// test --> A, B, C, D, E, F
	tsf1 := action.NewTransfer(1, big.NewInt(20), ta.Addrinfo["miner"].RawAddress, ta.Addrinfo["alfa"].RawAddress)
	tsf1, err := tsf1.Sign(ta.Addrinfo["miner"])
	tsf2 := action.NewTransfer(1, big.NewInt(30), ta.Addrinfo["miner"].RawAddress, ta.Addrinfo["bravo"].RawAddress)
	tsf2, err = tsf2.Sign(ta.Addrinfo["miner"])
	tsf3 := action.NewTransfer(1, big.NewInt(50), ta.Addrinfo["miner"].RawAddress, ta.Addrinfo["charlie"].RawAddress)
	tsf3, err = tsf3.Sign(ta.Addrinfo["miner"])
	tsf4 := action.NewTransfer(1, big.NewInt(70), ta.Addrinfo["miner"].RawAddress, ta.Addrinfo["delta"].RawAddress)
	tsf4, err = tsf4.Sign(ta.Addrinfo["miner"])
	tsf5 := action.NewTransfer(1, big.NewInt(110), ta.Addrinfo["miner"].RawAddress, ta.Addrinfo["echo"].RawAddress)
	tsf5, err = tsf5.Sign(ta.Addrinfo["miner"])
	tsf6 := action.NewTransfer(1, big.NewInt(50<<20), ta.Addrinfo["miner"].RawAddress, ta.Addrinfo["foxtrot"].RawAddress)
	tsf6, err = tsf6.Sign(ta.Addrinfo["miner"])

	blk, err := bc.MintNewBlock([]*action.Transfer{tsf1, tsf2, tsf3, tsf4, tsf5, tsf6}, nil, ta.Addrinfo["miner"], "")
	if err != nil {
		return err
	}
	if err := bc.CommitBlock(blk); err != nil {
		return err
	}

	// Add block 2
	// Charlie --> A, B, D, E, test
	tsf1 = action.NewTransfer(1, big.NewInt(1), ta.Addrinfo["charlie"].RawAddress, ta.Addrinfo["alfa"].RawAddress)
	tsf1, err = tsf1.Sign(ta.Addrinfo["charlie"])
	tsf2 = action.NewTransfer(1, big.NewInt(1), ta.Addrinfo["charlie"].RawAddress, ta.Addrinfo["bravo"].RawAddress)
	tsf2, err = tsf2.Sign(ta.Addrinfo["charlie"])
	tsf3 = action.NewTransfer(1, big.NewInt(1), ta.Addrinfo["charlie"].RawAddress, ta.Addrinfo["delta"].RawAddress)
	tsf3, err = tsf3.Sign(ta.Addrinfo["charlie"])
	tsf4 = action.NewTransfer(1, big.NewInt(1), ta.Addrinfo["charlie"].RawAddress, ta.Addrinfo["echo"].RawAddress)
	tsf4, err = tsf4.Sign(ta.Addrinfo["charlie"])
	tsf5 = action.NewTransfer(1, big.NewInt(1), ta.Addrinfo["charlie"].RawAddress, ta.Addrinfo["miner"].RawAddress)
	tsf5, err = tsf5.Sign(ta.Addrinfo["charlie"])
	blk, err = bc.MintNewBlock([]*action.Transfer{tsf1, tsf2, tsf3, tsf4, tsf5}, nil, ta.Addrinfo["miner"], "")
	if err != nil {
		return err
	}
	if err := bc.CommitBlock(blk); err != nil {
		return err
	}

	// Add block 3
	// Delta --> B, E, F, test
	tsf1 = action.NewTransfer(1, big.NewInt(1), ta.Addrinfo["delta"].RawAddress, ta.Addrinfo["bravo"].RawAddress)
	tsf1, err = tsf1.Sign(ta.Addrinfo["delta"])
	tsf2 = action.NewTransfer(1, big.NewInt(1), ta.Addrinfo["delta"].RawAddress, ta.Addrinfo["echo"].RawAddress)
	tsf2, err = tsf2.Sign(ta.Addrinfo["delta"])
	tsf3 = action.NewTransfer(1, big.NewInt(1), ta.Addrinfo["delta"].RawAddress, ta.Addrinfo["foxtrot"].RawAddress)
	tsf3, err = tsf3.Sign(ta.Addrinfo["delta"])
	tsf4 = action.NewTransfer(1, big.NewInt(1), ta.Addrinfo["delta"].RawAddress, ta.Addrinfo["miner"].RawAddress)
	tsf4, err = tsf4.Sign(ta.Addrinfo["delta"])
	blk, err = bc.MintNewBlock([]*action.Transfer{tsf1, tsf2, tsf3, tsf4}, nil, ta.Addrinfo["miner"], "")
	if err != nil {
		return err
	}
	if err := bc.CommitBlock(blk); err != nil {
		return err
	}

	// Add block 4
	// Delta --> A, B, C, D, F, test
	tsf1 = action.NewTransfer(1, big.NewInt(2), ta.Addrinfo["echo"].RawAddress, ta.Addrinfo["alfa"].RawAddress)
	tsf1, err = tsf1.Sign(ta.Addrinfo["echo"])
	tsf2 = action.NewTransfer(1, big.NewInt(2), ta.Addrinfo["echo"].RawAddress, ta.Addrinfo["bravo"].RawAddress)
	tsf2, err = tsf2.Sign(ta.Addrinfo["echo"])
	tsf3 = action.NewTransfer(1, big.NewInt(2), ta.Addrinfo["echo"].RawAddress, ta.Addrinfo["charlie"].RawAddress)
	tsf3, err = tsf3.Sign(ta.Addrinfo["echo"])
	tsf4 = action.NewTransfer(1, big.NewInt(2), ta.Addrinfo["echo"].RawAddress, ta.Addrinfo["delta"].RawAddress)
	tsf4, err = tsf4.Sign(ta.Addrinfo["echo"])
	tsf5 = action.NewTransfer(1, big.NewInt(2), ta.Addrinfo["echo"].RawAddress, ta.Addrinfo["foxtrot"].RawAddress)
	tsf5, err = tsf5.Sign(ta.Addrinfo["echo"])
	tsf6 = action.NewTransfer(1, big.NewInt(2), ta.Addrinfo["echo"].RawAddress, ta.Addrinfo["miner"].RawAddress)
	tsf6, err = tsf6.Sign(ta.Addrinfo["echo"])
	vote1 := action.NewVote(0, ta.Addrinfo["charlie"].PublicKey, ta.Addrinfo["alfa"].PublicKey)
	vote2 := action.NewVote(1, ta.Addrinfo["alfa"].PublicKey, ta.Addrinfo["charlie"].PublicKey)
	vote1, err = vote1.Sign(ta.Addrinfo["charlie"])
	if err != nil {
		return err
	}

	vote2, err = vote2.Sign(ta.Addrinfo["alfa"])
	if err != nil {
		return err
	}

	blk, err = bc.MintNewBlock([]*action.Transfer{tsf1, tsf2, tsf3, tsf4, tsf5, tsf6}, []*action.Vote{vote1, vote2}, ta.Addrinfo["miner"], "")
	if err != nil {
		return err
	}
	if err := bc.CommitBlock(blk); err != nil {
		return err
	}

	return nil
}

func TestCreateBlockchain(t *testing.T) {
	assert := assert.New(t)

	config, err := config.LoadConfigWithPathWithoutValidation(testingConfigPath)
	assert.Nil(err)
	// disable account-based testing
	config.Chain.TrieDBPath = ""
	config.Chain.InMemTest = true
	// Disable block reward to make bookkeeping easier
	Gen.BlockReward = uint64(0)

	// create chain
	bc := CreateBlockchain(config, nil)
	assert.NotNil(bc)
	height, err := bc.TipHeight()
	assert.Nil(err)
	assert.Equal(0, int(height))
	fmt.Printf("Create blockchain pass, height = %d\n", height)
	defer bc.Stop()

	// verify Genesis block
	genesis, _ := bc.GetBlockByHeight(0)
	assert.NotNil(genesis)
	// serialize
	data, err := genesis.Serialize()
	assert.Nil(err)

	assert.Equal(10, len(genesis.Transfers))
	assert.Equal(21, len(genesis.Votes))

	fmt.Printf("Block size match pass\n")
	fmt.Printf("Marshaling Block pass\n")

	// deserialize
	deserialize := Block{}
	err = deserialize.Deserialize(data)
	assert.Nil(err)
	fmt.Printf("Unmarshaling Block pass\n")

	hash := genesis.HashBlock()
	assert.Equal(hash, deserialize.HashBlock())
	fmt.Printf("Serialize/Deserialize Block hash = %x match\n", hash)

	hash = genesis.TxRoot()
	assert.Equal(hash, deserialize.TxRoot())
	fmt.Printf("Serialize/Deserialize Block merkle = %x match\n", hash)

	// add 4 sample blocks
	assert.Nil(addTestingTsfBlocks(bc))
	height, err = bc.TipHeight()
	assert.Nil(err)
	assert.Equal(4, int(height))
}

func TestLoadBlockchainfromDB(t *testing.T) {
	require := require.New(t)

	config, err := config.LoadConfigWithPathWithoutValidation(testingConfigPath)
	require.Nil(err)
	util.CleanupPath(t, testTriePath)
	defer util.CleanupPath(t, testTriePath)
	util.CleanupPath(t, testDBPath)
	defer util.CleanupPath(t, testDBPath)

	config.Chain.TrieDBPath = testTriePath
	config.Chain.InMemTest = false
	config.Chain.ChainDBPath = testDBPath

	tr, _ := trie.NewTrie(testTriePath, false)
	sf := state.NewFactory(tr)
	sf.CreateState(ta.Addrinfo["miner"].RawAddress, Gen.TotalSupply)
	// Disable block reward to make bookkeeping easier
	Gen.BlockReward = uint64(0)
	// Create a blockchain from scratch
	bc := CreateBlockchain(config, sf)
	require.NotNil(bc)
	height, err := bc.TipHeight()
	require.Nil(err)
	fmt.Printf("Open blockchain pass, height = %d\n", height)
	require.Nil(addTestingTsfBlocks(bc))
	bc.Stop()

	// Load a blockchain from DB
	bc = CreateBlockchain(config, sf)
	defer bc.Stop()
	require.NotNil(bc)

	// check hash<-->height mapping
	hash, err := bc.GetHashByHeight(0)
	require.Nil(err)
	height, err = bc.GetHeightByHash(hash)
	require.Nil(err)
	require.Equal(uint64(0), height)
	blk, err := bc.GetBlockByHash(hash)
	require.Nil(err)
	require.Equal(hash, blk.HashBlock())
	fmt.Printf("Genesis hash = %x\n", hash)

	hash1, err := bc.GetHashByHeight(1)
	require.Nil(err)
	height, err = bc.GetHeightByHash(hash1)
	require.Nil(err)
	require.Equal(uint64(1), height)
	blk, err = bc.GetBlockByHash(hash1)
	require.Nil(err)
	require.Equal(hash1, blk.HashBlock())
	fmt.Printf("block 1 hash = %x\n", hash1)

	hash2, err := bc.GetHashByHeight(2)
	require.Nil(err)
	height, err = bc.GetHeightByHash(hash2)
	require.Nil(err)
	require.Equal(uint64(2), height)
	blk, err = bc.GetBlockByHash(hash2)
	require.Nil(err)
	require.Equal(hash2, blk.HashBlock())
	fmt.Printf("block 2 hash = %x\n", hash2)

	hash3, err := bc.GetHashByHeight(3)
	require.Nil(err)
	height, err = bc.GetHeightByHash(hash3)
	require.Nil(err)
	require.Equal(uint64(3), height)
	blk, err = bc.GetBlockByHash(hash3)
	require.Nil(err)
	require.Equal(hash3, blk.HashBlock())
	fmt.Printf("block 3 hash = %x\n", hash3)

	hash4, err := bc.GetHashByHeight(4)
	require.Nil(err)
	height, err = bc.GetHeightByHash(hash4)
	require.Nil(err)
	require.Equal(uint64(4), height)
	blk, err = bc.GetBlockByHash(hash4)
	require.Nil(err)
	require.Equal(hash4, blk.HashBlock())
	fmt.Printf("block 4 hash = %x\n", hash4)

	empblk, err := bc.GetBlockByHash(common.ZeroHash32B)
	require.Nil(empblk)
	require.NotNil(err.Error())

	blk, err = bc.GetBlockByHeight(60000)
	require.Nil(blk)
	require.NotNil(err)

	// add wrong blocks
	h, err := bc.TipHeight()
	require.Nil(err)
	hash, err = bc.TipHash()
	require.Nil(err)
	blk, err = bc.GetBlockByHeight(h)
	require.Nil(err)
	require.Equal(hash, blk.HashBlock())
	fmt.Printf("Current tip = %d hash = %x\n", h, hash)

	// add block with wrong height
	cbTsf := action.NewCoinBaseTransfer(big.NewInt(50), ta.Addrinfo["bravo"].RawAddress)
	require.NotNil(cbTsf)
	blk = NewBlock(0, h+2, hash, []*action.Transfer{cbTsf}, nil)
	err = bc.ValidateBlock(blk)
	require.NotNil(err)
	fmt.Printf("Cannot validate block %d: %v\n", blk.Height(), err)

	// add block with zero prev hash
	cbTsf2 := action.NewCoinBaseTransfer(big.NewInt(50), ta.Addrinfo["bravo"].RawAddress)
	require.NotNil(cbTsf2)
	blk = NewBlock(0, h+1, common.ZeroHash32B, []*action.Transfer{cbTsf2}, nil)
	err = bc.ValidateBlock(blk)
	require.NotNil(err)
	fmt.Printf("Cannot validate block %d: %v\n", blk.Height(), err)

	// cannot add existing block again
	blk, err = bc.GetBlockByHeight(3)
	require.NotNil(blk)
	require.Nil(err)
	err = bc.(*blockchain).commitBlock(blk)
	require.NotNil(err)
	fmt.Printf("Cannot add block 3 again: %v\n", err)

	// check all Tx from block 4
	blk, err = bc.GetBlockByHeight(4)
	require.Nil(err)
	require.Equal(hash4, blk.HashBlock())
	for _, transfer := range blk.Transfers {
		transferHash := transfer.Hash()
		hash, err := bc.GetBlockHashByTransferHash(transferHash)
		require.Nil(err)
		require.Equal(hash, hash4)
		transfer1, err := bc.GetTransferByTransferHash(transferHash)
		require.Nil(err)
		require.Equal(transfer1.Hash(), transferHash)
	}

	for _, vote := range blk.Votes {
		voteHash := vote.Hash()
		hash, err := bc.GetBlockHashByVoteHash(voteHash)
		require.Nil(err)
		require.Equal(hash, hash4)
		vote1, err := bc.GetVoteByVoteHash(voteHash)
		require.Nil(err)
		require.Equal(vote1.Hash(), voteHash)
	}

	fromTransfers, err := bc.GetTransfersFromAddress(ta.Addrinfo["charlie"].RawAddress)
	require.Nil(err)
	require.Equal(len(fromTransfers), 5)

	toTransfers, err := bc.GetTransfersToAddress(ta.Addrinfo["charlie"].RawAddress)
	require.Nil(err)
	require.Equal(len(toTransfers), 2)

	fromVotes, err := bc.GetVotesFromAddress(ta.Addrinfo["charlie"].RawAddress)
	require.Nil(err)
	require.Equal(len(fromVotes), 1)

	fromVotes, err = bc.GetVotesFromAddress(ta.Addrinfo["alfa"].RawAddress)
	require.Nil(err)
	require.Equal(len(fromVotes), 1)

	toVotes, err := bc.GetVotesToAddress(ta.Addrinfo["charlie"].RawAddress)
	require.Nil(err)
	require.Equal(len(toVotes), 1)

	toVotes, err = bc.GetVotesToAddress(ta.Addrinfo["alfa"].RawAddress)
	require.Nil(err)
	require.Equal(len(toVotes), 1)

	totalTransfers, err := bc.GetTotalTransfers()
	require.Nil(err)
	require.Equal(totalTransfers, uint64(35))

	totalVotes, err := bc.GetTotalVotes()
	require.Nil(err)
	require.Equal(totalVotes, uint64(23))
}

func TestBlockchain_Validator(t *testing.T) {
	config, err := config.LoadConfigWithPathWithoutValidation(testingConfigPath)
	assert.Nil(t, err)
	// disable account-based testing
	config.Chain.TrieDBPath = ""
	config.Chain.InMemTest = true

	bc := CreateBlockchain(config, nil)
	defer bc.Stop()
	assert.NotNil(t, bc)

	val := bc.Validator()
	assert.NotNil(t, bc)
	bc.SetValidator(val)
	assert.NotNil(t, bc.Validator())
}

func TestBlockchain_MintNewDummyBlock(t *testing.T) {
	config, err := config.LoadConfigWithPathWithoutValidation(testingConfigPath)
	assert.Nil(t, err)
	// disable account-based testing
	config.Chain.TrieDBPath = ""
	config.Chain.InMemTest = true

	bc := CreateBlockchain(config, nil)
	defer bc.Stop()
	assert.NotNil(t, bc)

	blk, err := bc.MintNewDummyBlock()
	assert.Nil(t, err)
	assert.Equal(t, uint64(1), blk.Height())
}

func TestBlockchainInitialCandidate(t *testing.T) {
	require := require.New(t)

	config, err := config.LoadConfigWithPathWithoutValidation(testingConfigPath)
	require.Nil(err)
	util.CleanupPath(t, testTriePath)
	defer util.CleanupPath(t, testTriePath)
	util.CleanupPath(t, testDBPath)
	defer util.CleanupPath(t, testDBPath)

	config.Chain.TrieDBPath = testTriePath
	config.Chain.InMemTest = false
	config.Chain.ChainDBPath = testDBPath
	// Disable block reward to make bookkeeping easier
	Gen.BlockReward = uint64(0)

	tr, _ := trie.NewTrie(testTriePath, false)
	sf := state.NewFactory(tr)

	height, candidate := sf.Candidates()
	require.True(height == 0)
	require.True(len(candidate) == 0)
	bc := CreateBlockchain(config, sf)
	require.NotNil(t, bc)
	// TODO: change the value when Candidates size is changed
	height, candidate = sf.Candidates()
	require.True(height == 0)
	require.True(len(candidate) == 2)
}

func TestCoinbaseTransfer(t *testing.T) {
	require := require.New(t)
	config, err := config.LoadConfigWithPathWithoutValidation(testingConfigPath)
	require.Nil(err)
	util.CleanupPath(t, testTriePath)
	defer util.CleanupPath(t, testTriePath)
	util.CleanupPath(t, testDBPath)
	defer util.CleanupPath(t, testDBPath)

	config.Chain.TrieDBPath = testTriePath
	config.Chain.InMemTest = false
	config.Chain.ChainDBPath = testDBPath

	tr, _ := trie.NewTrie(testTriePath, false)
	sf := state.NewFactory(tr)
	sf.CreateState(ta.Addrinfo["miner"].RawAddress, Gen.TotalSupply)

	Gen.BlockReward = uint64(10)

	bc := CreateBlockchain(config, sf)
	require.NotNil(bc)
	height, err := bc.TipHeight()
	require.Nil(err)
	require.Equal(0, int(height))

	transfers := []*action.Transfer{}
	blk, err := bc.MintNewBlock(transfers, nil, ta.Addrinfo["miner"], "")
	require.Nil(err)
	s, err := bc.StateByAddr(ta.Addrinfo["miner"].RawAddress)
	require.Nil(err)
	b := s.Balance
	require.True(b.String() == strconv.Itoa(int(Gen.TotalSupply)))
	err = bc.CommitBlock(blk)
	require.Nil(err)
	height, err = bc.TipHeight()
	require.Nil(err)
	require.True(height == 1)
	require.True(len(blk.Transfers) == 1)
	s, err = bc.StateByAddr(ta.Addrinfo["miner"].RawAddress)
	require.Nil(err)
	b = s.Balance
	require.True(b.String() == strconv.Itoa(int(Gen.TotalSupply)+int(Gen.BlockReward)))
}

func TestBlockchain_StateByAddr(t *testing.T) {
	require := require.New(t)

	config, err := config.LoadConfigWithPathWithoutValidation(testingConfigPath)
	require.Nil(err)
	// disable account-based testing
	config.Chain.InMemTest = true
	// create chain
	bc := CreateBlockchain(config, nil)
	require.NotNil(bc)

	s, _ := bc.StateByAddr(Gen.CreatorAddr)
	require.Equal(uint64(0), s.Nonce)
	require.Equal(big.NewInt(9900000000), s.Balance)
	require.Equal(Gen.CreatorAddr, s.Address)
	require.Equal(false, s.IsCandidate)
	require.Equal(big.NewInt(0), s.VotingWeight)
	require.Equal("", s.Votee)
	require.Equal(map[string]*big.Int(map[string]*big.Int(nil)), s.Voters)
}
