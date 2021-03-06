// Copyright (c) 2018 IoTeX
// This is an alpha (internal) release and is not suitable for production. This source code is provided 'as is' and no
// warranties are given as to title or non-infringement, merchantability or fitness for purpose and, to the extent
// permitted by law, all liability for your use of the code is disclaimed. This source code is governed by Apache
// License 2.0 that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"runtime/pprof"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/iotexproject/iotex-core/actpool"
	"github.com/iotexproject/iotex-core/blockchain"
	"github.com/iotexproject/iotex-core/blocksync"
	"github.com/iotexproject/iotex-core/common"
	"github.com/iotexproject/iotex-core/config"
	"github.com/iotexproject/iotex-core/consensus"
	"github.com/iotexproject/iotex-core/delegate"
	"github.com/iotexproject/iotex-core/iotxaddress"
	"github.com/iotexproject/iotex-core/logger"
	"github.com/iotexproject/iotex-core/network"
	pb "github.com/iotexproject/iotex-core/simulator/proto/simulator"
	"github.com/iotexproject/iotex-core/state"
)

const (
	port           = ":50051"
	rolldposConfig = "./config_local_rolldpos_sim.yaml"
	dummyMsgType   = 1999
)

// server is used to implement message.SimulatorServer.
type (
	server struct {
		nodes []consensus.Sim // slice of Consensus objects
	}
	byzVal struct {
		val blockchain.Validator
	}
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

// Validate for the byzantine node uses the actual block validator and returns the opposite
func (v *byzVal) Validate(blk *blockchain.Block, tipHeight uint64, tipHash common.Hash32B) error {
	//err := v.val.Validate(blk, tipHeight, tipHash)
	//if err != nil {
	//	return nil
	//}
	//return errors.New("")
	return nil
}

// Ping implements simulator.SimulatorServer
func (s *server) Init(in *pb.InitRequest, stream pb.Simulator_InitServer) error {
	nPlayers := in.NBF + in.NFS + in.NHonest

	var addrs []string // all delegate addresses
	for i := 0; i < int(nPlayers); i++ {
		addrs = append(addrs, "127.0.0.1:32"+strconv.Itoa(i))
	}

	for i := 0; i < int(nPlayers); i++ {
		cfg, err := config.LoadConfigWithPathWithoutValidation(rolldposConfig)
		if err != nil {
			logger.Error().Msg("Error loading config file")
		}

		//s.nodes = make([]consensus.Sim, in.NPlayers) // allocate all the necessary space now because otherwise nodes will get copied and create pointer issues

		// handle node address, delegate addresses, etc.
		cfg.Delegate.Addrs = addrs
		cfg.Network.Addr = addrs[i]

		// create public/private key pair and address
		chainID := make([]byte, 4)
		binary.LittleEndian.PutUint32(chainID, uint32(i))

		addr, err := iotxaddress.NewAddress(true, chainID)
		if err != nil {
			logger.Error().Err(err).Msg("failed to create public/private key pair together with the address derived.")
		}

		cfg.Chain.ProducerAddr.PublicKey = addr.PublicKey
		cfg.Chain.ProducerAddr.PrivateKey = addr.PrivateKey
		cfg.Chain.ProducerAddr.RawAddress = addr.RawAddress

		// set chain database path
		cfg.Chain.ChainDBPath = "./chain" + strconv.Itoa(i) + ".db"

		sf, _ := state.NewFactoryFromTrieDBPath(cfg.Chain.TrieDBPath, false)
		bc := blockchain.CreateBlockchain(cfg, sf)

		if i >= int(in.NFS+in.NHonest) { // is byzantine node
			val := bc.Validator()
			byzVal := &byzVal{val: val}
			bc.SetValidator(byzVal)
		}

		overlay := network.NewOverlay(&cfg.Network)
		ap := actpool.NewActPool(sf)
		dlg := delegate.NewConfigBasedPool(&cfg.Delegate)
		bs, _ := blocksync.NewBlockSyncer(cfg, bc, ap, overlay, dlg)
		bs.Start()

		var node consensus.Sim
		if i < int(in.NHonest) {
			node = consensus.NewSim(cfg, bc, bs, dlg, sf)
		} else if i < int(in.NHonest+in.NFS) {
			s.nodes = append(s.nodes, nil)
			continue
		} else {
			node = consensus.NewSimByzantine(cfg, bc, bs, dlg, sf)
		}

		s.nodes = append(s.nodes, node)

		done := make(chan bool)
		node.SetDoneStream(done)

		node.Start()

		fmt.Printf("Node %d initialized and consensus engine started\n", i)
		time.Sleep(2 * time.Millisecond)
		<-done

		fmt.Printf("Node %d initialization ended\n", i)

		//s.nodes = append(s.nodes, node)
	}

	for i := 0; i < int(in.NFS); i++ {
		s.nodes = append(s.nodes, nil)
	}

	fmt.Printf("Simulator initialized with %d players\n", nPlayers)

	return nil
}

// Ping implements simulator.SimulatorServer
func (s *server) Ping(in *pb.Request, stream pb.Simulator_PingServer) error {
	fmt.Println()

	fmt.Printf("Node %d pinged; opened message stream\n", in.PlayerID)
	msgValue, err := hex.DecodeString(in.Value)
	if err != nil {
		logger.Error().Msg("Could not decode message value into byte array")
	}

	done := make(chan bool)

	s.nodes[in.PlayerID].SetStream(&stream)

	s.nodes[in.PlayerID].SendUnsent()

	// message type of 1999 means that it's a dummy message to allow the engine to pass back proposed blocks
	if in.InternalMsgType != dummyMsgType {
		msg := consensus.CombineMsg(in.InternalMsgType, msgValue)
		s.nodes[in.PlayerID].HandleViewChange(msg, done)
		time.Sleep(2 * time.Millisecond)
		<-done // wait until done
	}

	fmt.Println("closed message stream")
	return nil
}

func (s *server) Exit(context context.Context, in *pb.Empty) (*pb.Empty, error) {
	defer os.Exit(0)
	defer pprof.StopCPUProfile()
	return &pb.Empty{}, nil
}

func main() {
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
	}

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterSimulatorServer(s, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
