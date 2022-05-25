package e2e

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/suite"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/osmosis-labs/osmosis/v7/tests/e2e/chain"
	dockerconfig "github.com/osmosis-labs/osmosis/v7/tests/e2e/docker"
	net "github.com/osmosis-labs/osmosis/v7/tests/e2e/network"
	"github.com/osmosis-labs/osmosis/v7/tests/e2e/util"
)

type status struct {
	LatestHeight string `json:"latest_block_height"`
}

type syncInfo struct {
	SyncInfo status `json:"SyncInfo"`
}

const (
	// osmosis version being upgraded to (folder must exist here https://github.com/osmosis-labs/osmosis/tree/main/app/upgrades)
	upgradeVersion = "v9"

	// max retries for json unmarshalling
	maxRetries = 60
)

var (
	// whatever number of validator configs get posted here are how many validators that will spawn on chain A and B respectively
	validatorConfigsChainA = []*chain.ValidatorConfig{
		{
			Pruning:            "default",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   20,
			SnapshotKeepRecent: 2,
		},
		{
			Pruning:            "nothing",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   30,
			SnapshotKeepRecent: 1,
		},
		{
			Pruning:            "custom",
			PruningKeepRecent:  "10000",
			PruningInterval:    "13",
			SnapshotInterval:   15,
			SnapshotKeepRecent: 3,
		},
		{
			Pruning:            "everything",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   0,
			SnapshotKeepRecent: 0,
		},
	}
	validatorConfigsChainB = []*chain.ValidatorConfig{
		{
			Pruning:            "default",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   1500,
			SnapshotKeepRecent: 2,
		},
		{
			Pruning:            "nothing",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   1500,
			SnapshotKeepRecent: 2,
		},
		{
			Pruning:            "custom",
			PruningKeepRecent:  "10000",
			PruningInterval:    "13",
			SnapshotInterval:   1500,
			SnapshotKeepRecent: 2,
		},
	}
)

type IntegrationTestSuite struct {
	suite.Suite

	tmpDirs  []string
	networks []*net.Network

	workingDirectory string

	skipIBC       bool
	skipUpgrade   bool
	skipStateSync bool

	dockerImages    *dockerconfig.ImageConfig
	dockerResources *dockerconfig.Resources
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up e2e integration test suite...")

	// The e2e test flow is as follows:
	//
	// 1. Configure two chains - chan A and chain B.
	//   * For each chain, set up two validators
	//   * Initialize configs and genesis for all validators.
	// 2. Start both networks.
	// 3. Run IBC relayer betweeen the two chains.
	// 4. Execute various e2e tests, including IBC.

	var err error

	s.workingDirectory, err = os.Getwd()
	s.Require().NoError(err)

	if str := os.Getenv("OSMOSIS_E2E_SKIP_IBC"); len(str) > 0 {
		s.skipIBC, err = strconv.ParseBool(str)
		s.Require().NoError(err)
		s.T().Log("skipping IBC testing")
	}

	if str := os.Getenv("OSMOSIS_E2E_SKIP_UPGRADE"); len(str) > 0 {
		s.skipUpgrade, err = strconv.ParseBool(str)
		s.Require().NoError(err)
		s.T().Log("skipping upgrade testing")
	}

	if str := os.Getenv("OSMOSIS_E2E_SKIP_STATE_SYNC"); len(str) > 0 {
		s.skipStateSync, err = strconv.ParseBool(str)
		s.Require().NoError(err)
		s.T().Log("skipping state sync testing")
	}

	if s.skipIBC && !s.skipUpgrade {
		s.T().Fatal("if IBC is skipped, upgrade must be as well")
	}

	s.dockerImages = dockerconfig.NewImageConfig(!s.skipUpgrade)

	s.dockerResources, err = dockerconfig.NewResources()
	s.Require().NoError(err)

	s.networks = make([]*net.Network, 0)
	s.configureChain(chain.ChainAID, validatorConfigsChainA)

	if !s.skipIBC {
		s.configureChain(chain.ChainBID, validatorConfigsChainB)
	}

	for _, network := range s.networks {
		validatorResources, err := network.RunValidators()
		s.Require().NoError(err)
		s.dockerResources.Validators[network.GetChain().ChainMeta.Id] = validatorResources

		s.Require().NoError(network.WaitUntilHeight(3))
	}

	if !s.skipIBC {
		// Run a relayer between every possible pair of chains.
		for i := 0; i < len(s.networks); i++ {
			for j := i + 1; j < len(s.networks); j++ {
				s.runIBCRelayer(s.networks[i].GetChain(), s.networks[j].GetChain())
			}
		}
	}

	if !s.skipUpgrade {
		s.createPreUpgradeState()
		s.upgrade()
		s.runPostUpgradeTests()
	}

	if !s.skipStateSync {
		s.configureStateSync(s.networks[0])
	}
}

func (s *IntegrationTestSuite) TearDownSuite() {
	if str := os.Getenv("OSMOSIS_E2E_SKIP_CLEANUP"); len(str) > 0 {
		skipCleanup, err := strconv.ParseBool(str)
		s.Require().NoError(err)

		if skipCleanup {
			return
		}
	}

	s.T().Log("tearing down e2e integration test suite...")

	err := s.dockerResources.Purge()
	s.Require().NoError(err)

	for _, network := range s.networks {
		os.RemoveAll(network.GetChain().ChainMeta.DataDir)
	}

	for _, td := range s.tmpDirs {
		os.RemoveAll(td)
	}
}

func (s *IntegrationTestSuite) runIBCRelayer(chainA *chain.Chain, chainB *chain.Chain) {
	s.T().Log("starting Hermes relayer container...")

	tmpDir, err := ioutil.TempDir("", "osmosis-e2e-testnet-hermes-")
	s.Require().NoError(err)
	s.tmpDirs = append(s.tmpDirs, tmpDir)

	osmoAVal := chainA.Validators[0]
	osmoBVal := chainB.Validators[0]
	hermesCfgPath := path.Join(tmpDir, "hermes")

	s.Require().NoError(os.MkdirAll(hermesCfgPath, 0o755))
	_, err = util.CopyFile(
		filepath.Join("./scripts/", "hermes_bootstrap.sh"),
		filepath.Join(hermesCfgPath, "hermes_bootstrap.sh"),
	)
	s.Require().NoError(err)

	s.dockerResources.Hermes, err = s.dockerResources.Pool.RunWithOptions(
		&dockertest.RunOptions{
			Name:       fmt.Sprintf("%s-%s-relayer", chainA.ChainMeta.Id, chainB.ChainMeta.Id),
			Repository: s.dockerImages.RelayerRepository,
			Tag:        s.dockerImages.RelayerTag,
			NetworkID:  s.dockerResources.Network.Network.ID,
			Cmd: []string{
				"start",
			},
			User: "root:root",
			Mounts: []string{
				fmt.Sprintf("%s/:/root/hermes", hermesCfgPath),
			},
			ExposedPorts: []string{
				"3031",
			},
			PortBindings: map[docker.Port][]docker.PortBinding{
				"3031/tcp": {{HostIP: "", HostPort: "3031"}},
			},
			Env: []string{
				fmt.Sprintf("OSMO_A_E2E_CHAIN_ID=%s", chainA.ChainMeta.Id),
				fmt.Sprintf("OSMO_B_E2E_CHAIN_ID=%s", chainB.ChainMeta.Id),
				fmt.Sprintf("OSMO_A_E2E_VAL_MNEMONIC=%s", osmoAVal.Mnemonic),
				fmt.Sprintf("OSMO_B_E2E_VAL_MNEMONIC=%s", osmoBVal.Mnemonic),
				fmt.Sprintf("OSMO_A_E2E_VAL_HOST=%s", s.dockerResources.Validators[chainA.ChainMeta.Id][0].Container.Name[1:]),
				fmt.Sprintf("OSMO_B_E2E_VAL_HOST=%s", s.dockerResources.Validators[chainB.ChainMeta.Id][0].Container.Name[1:]),
			},
			Entrypoint: []string{
				"sh",
				"-c",
				"chmod +x /root/hermes/hermes_bootstrap.sh && /root/hermes/hermes_bootstrap.sh",
			},
		},
		noRestart,
	)
	s.Require().NoError(err)

	endpoint := fmt.Sprintf("http://%s/state", s.dockerResources.Hermes.GetHostPort("3031/tcp"))
	s.Require().Eventually(
		func() bool {
			resp, err := http.Get(endpoint)
			if err != nil {
				return false
			}

			defer resp.Body.Close()

			bz, err := io.ReadAll(resp.Body)
			if err != nil {
				return false
			}

			var respBody map[string]interface{}
			if err := json.Unmarshal(bz, &respBody); err != nil {
				return false
			}

			status := respBody["status"].(string)
			result := respBody["result"].(map[string]interface{})

			return status == "success" && len(result["chains"].([]interface{})) == 2
		},
		5*time.Minute,
		time.Second,
		"hermes relayer not healthy",
	)

	s.T().Logf("started Hermes relayer container: %s", s.dockerResources.Hermes.Container.ID)

	// XXX: Give time to both networks to start, otherwise we might see gRPC
	// transport errors.
	time.Sleep(10 * time.Second)

	// create the client, connection and channel between the two Osmosis chains
	s.connectIBCChains(chainA, chainB)
}

func (s *IntegrationTestSuite) configureChain(chainId string, validatorConfigs []*chain.ValidatorConfig) {
	s.T().Logf("starting e2e infrastructure for chain-id: %s", chainId)
	tmpDir, err := ioutil.TempDir("", "osmosis-e2e-testnet-")

	s.T().Logf("temp directory for chain-id %v: %v", chainId, tmpDir)
	s.Require().NoError(err)

	validatorConfigBytes, err := json.Marshal(validatorConfigs)
	s.Require().NoError(err)

	newNetwork := net.New(s.T(), len(s.networks), len(validatorConfigs), s.dockerResources, s.dockerImages, s.workingDirectory)

	votingPeriodDuration := time.Duration(int(newNetwork.GetVotingPeriod()) * 1000000000)

	initResource, err := s.dockerResources.Pool.RunWithOptions(
		&dockertest.RunOptions{
			Name:       fmt.Sprintf("%s", chainId),
			Repository: s.dockerImages.InitRepository,
			Tag:        s.dockerImages.InitTag,
			NetworkID:  s.dockerResources.Network.Network.ID,
			Cmd: []string{
				fmt.Sprintf("--data-dir=%s", tmpDir),
				fmt.Sprintf("--chain-id=%s", chainId),
				fmt.Sprintf("--config=%s", validatorConfigBytes),
				fmt.Sprintf("--voting-period=%v", votingPeriodDuration),
			},
			User: "root:root",
			Mounts: []string{
				fmt.Sprintf("%s:%s", tmpDir, tmpDir),
			},
		},
		noRestart,
	)
	s.Require().NoError(err)

	fileName := fmt.Sprintf("%v/%v-encode", tmpDir, chainId)
	s.T().Logf("serialized init file for chain-id %v: %v", chainId, fileName)

	// loop through the reading and unmarshaling of the init file a total of maxRetries or until error is nil
	// without this, test attempts to unmarshal file before docker container is finished writing
	for i := 0; i < maxRetries; i++ {
		encJson, _ := os.ReadFile(fileName)
		err = json.Unmarshal(encJson, newNetwork.GetChain())
		if err == nil {
			break
		}

		if i == maxRetries-1 {
			s.Require().NoError(err)
		}

		if i > 0 {
			time.Sleep(1 * time.Second)
		}
	}
	s.Require().NoError(s.dockerResources.Pool.Purge(initResource))

	s.networks = append(s.networks, newNetwork)
}

func noRestart(config *docker.HostConfig) {
	// in this case we don't want the nodes to restart on failure
	config.RestartPolicy = docker.RestartPolicy{
		Name: "no",
	}
}

func (s *IntegrationTestSuite) upgrade() {
	// submit, deposit, and vote for upgrade proposal
	// prop height = current height + voting period + time it takes to submit proposal + small buffer
	for _, network := range s.networks {
		currentHeight, err := network.GetCurrentHeightFromValidator(0)
		s.Require().NoError(err)

		network.CalclulateAndSetProposalHeight(currentHeight)
		proposalHeight := network.GetProposalHeight()

		curChain := network.GetChain()
		s.submitProposal(curChain, proposalHeight)
		s.depositProposal(curChain)
		s.voteProposal(network)
	}

	// wait till all chains halt at upgrade height
	for _, network := range s.networks {
		curChain := network.GetChain()

		for i := range curChain.Validators {

			// use counter to ensure no new blocks are being created
			counter := 0
			s.T().Logf("waiting to reach upgrade height on %s validator container: %s", s.dockerResources.Validators[curChain.ChainMeta.Id][i].Container.Name[1:], s.dockerResources.Validators[curChain.ChainMeta.Id][i].Container.ID)
			s.Require().Eventually(
				func() bool {
					currentHeight, err := network.GetCurrentHeightFromValidator(i)
					s.Require().NoError(err)
					propHeight := network.GetProposalHeight()

					if currentHeight != propHeight {
						s.T().Logf("current block height on %s is %v, waiting for block %v container: %s", s.dockerResources.Validators[curChain.ChainMeta.Id][i].Container.Name[1:], currentHeight, propHeight, s.dockerResources.Validators[curChain.ChainMeta.Id][i].Container.ID)
					}
					if currentHeight > propHeight {
						panic("chain did not halt at upgrade height")
					}
					if currentHeight == propHeight {
						counter++
					}
					return counter == 3
				},
				5*time.Minute,
				time.Second,
			)
			s.T().Logf("reached upgrade height on %s container: %s", s.dockerResources.Validators[curChain.ChainMeta.Id][i].Container.Name[1:], s.dockerResources.Validators[curChain.ChainMeta.Id][i].Container.ID)
		}
	}

	// remove all containers so we can upgrade them to the new version
	for _, network := range s.networks {
		curChain := network.GetChain()
		for i := range curChain.Validators {
			if err := network.RemoveValidatorContainer(i); err != nil {
				s.Require().NoError(err)
			}
		}
	}

	for _, network := range s.networks {
		s.upgradeContainers(network)
	}
}

func (s *IntegrationTestSuite) upgradeContainers(network *net.Network) {
	s.dockerImages.SwitchToUpgradeImages()
	// upgrade containers to the locally compiled daemon
	chain := network.GetChain()
	s.T().Logf("starting upgrade for chain-id: %s...", chain.ChainMeta.Id)

	newValidatorResources, err := network.RunValidators()
	s.Require().NoError(err)
	s.dockerResources.Validators[network.GetChain().ChainMeta.Id] = newValidatorResources

	propHeight := network.GetProposalHeight()
	s.Require().NoError(network.WaitUntilHeight(propHeight))
}

func (s *IntegrationTestSuite) configureStateSync(network *net.Network) {
	currentHeight, err := network.GetCurrentHeightFromValidator(0)
	s.Require().NoError(err)

	// Ensure that state sync trust height is slightly lower than the latest
	// snapshot of every node
	stateSyncTrustHeight := currentHeight
	stateSyncTrustHash, err := network.GetHashFromBlock(stateSyncTrustHeight)
	s.Require().NoError(err)

	for valIdx := range network.GetChain().Validators {
		// Stop a validator container to update its config
		if err := network.RemoveValidatorContainer(valIdx); err != nil {
			s.Require().NoError(err)
		}

		configDir := network.GetChain().Validators[valIdx].ConfigDir

		stateSyncResource, err := s.dockerResources.Pool.RunWithOptions(
			&dockertest.RunOptions{
				Name:       fmt.Sprintf("state-sync-%s", network.GetChain().Validators[valIdx].Name),
				Repository: s.dockerImages.InitRepository,
				Tag:        s.dockerImages.InitTag,
				NetworkID:  s.dockerResources.Network.Network.ID,
				Cmd: []string{
					fmt.Sprintf("--config-dir=%s", configDir),
					fmt.Sprintf("--trust-height=%d", stateSyncTrustHeight),
					fmt.Sprintf("--trust-hash=%s", stateSyncTrustHash),
				},
				User: "root:root",
				Mounts: []string{
					fmt.Sprintf("%s:%s", configDir, configDir),
				},
			},
			noRestart,
		)
		s.Require().NoError(err)
		s.Require().NoError(s.dockerResources.Pool.Purge(stateSyncResource))

		// Start this validator in state sync test.
		if valIdx == 3 {
			continue
		}

		_, err = network.RunValidator(valIdx)
		s.Require().NoError(err)
	}

	maxSnapshotInterval := uint64(0)

	for _, valConfig := range validatorConfigsChainA {
		if valConfig.SnapshotInterval > maxSnapshotInterval {
			maxSnapshotInterval = valConfig.SnapshotInterval
		}
	}

	// Ensure that all restarted validators are making progress.
	doneCondition := func(syncInfo coretypes.SyncInfo) bool {
		return !syncInfo.CatchingUp && syncInfo.LatestBlockHeight > stateSyncTrustHeight+int64(maxSnapshotInterval)+1
	}

	for valIdx := range network.GetChain().Validators {
		if valIdx == 3 {
			continue
		}

		if err = network.WaitUntil(valIdx, doneCondition); err != nil {
			s.T().Errorf("validator with index %d failed to start", valIdx)
			s.Require().NoError(err)
		}
	}
}

func (s *IntegrationTestSuite) createPreUpgradeState() {
	chainA := s.networks[0].GetChain()
	chainB := s.networks[1].GetChain()

	s.sendIBC(chainA, chainB, chainB.Validators[0].PublicAddress, chain.OsmoToken)
	s.sendIBC(chainB, chainA, chainA.Validators[0].PublicAddress, chain.OsmoToken)
	s.sendIBC(chainA, chainB, chainB.Validators[0].PublicAddress, chain.StakeToken)
	s.sendIBC(chainB, chainA, chainA.Validators[0].PublicAddress, chain.StakeToken)
	s.createPool(chainA, "pool1A.json")
	s.createPool(chainB, "pool1B.json")
}

func (s *IntegrationTestSuite) runPostUpgradeTests() {
	chainA := s.networks[0].GetChain()
	chainB := s.networks[1].GetChain()

	s.sendIBC(chainA, chainB, chainB.Validators[0].PublicAddress, chain.OsmoToken)
	s.sendIBC(chainB, chainA, chainA.Validators[0].PublicAddress, chain.OsmoToken)
	s.sendIBC(chainA, chainB, chainB.Validators[0].PublicAddress, chain.StakeToken)
	s.sendIBC(chainB, chainA, chainA.Validators[0].PublicAddress, chain.StakeToken)
	s.createPool(chainA, "pool2A.json")
	s.createPool(chainB, "pool2B.json")
}
