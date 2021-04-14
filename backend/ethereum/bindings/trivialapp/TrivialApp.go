// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package trivialapp

import (
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// ChannelAllocation is an auto generated low-level Go binding around an user-defined struct.
type ChannelAllocation struct {
	Assets   []common.Address
	Balances [][]*big.Int
	Locked   []ChannelSubAlloc
}

// ChannelParams is an auto generated low-level Go binding around an user-defined struct.
type ChannelParams struct {
	ChallengeDuration *big.Int
	Nonce             *big.Int
	App               common.Address
	Participants      []common.Address
}

// ChannelState is an auto generated low-level Go binding around an user-defined struct.
type ChannelState struct {
	ChannelID [32]byte
	Version   uint64
	Outcome   ChannelAllocation
	AppData   []byte
	IsFinal   bool
}

// ChannelSubAlloc is an auto generated low-level Go binding around an user-defined struct.
type ChannelSubAlloc struct {
	ID       [32]byte
	Balances []*big.Int
}

// TrivialappABI is the input ABI used to generate the binding from.
const TrivialappABI = "[{\"inputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"challengeDuration\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"app\",\"type\":\"address\"},{\"internalType\":\"address[]\",\"name\":\"participants\",\"type\":\"address[]\"}],\"internalType\":\"structChannel.Params\",\"name\":\"params\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"},{\"components\":[{\"internalType\":\"address[]\",\"name\":\"assets\",\"type\":\"address[]\"},{\"internalType\":\"uint256[][]\",\"name\":\"balances\",\"type\":\"uint256[][]\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"ID\",\"type\":\"bytes32\"},{\"internalType\":\"uint256[]\",\"name\":\"balances\",\"type\":\"uint256[]\"}],\"internalType\":\"structChannel.SubAlloc[]\",\"name\":\"locked\",\"type\":\"tuple[]\"}],\"internalType\":\"structChannel.Allocation\",\"name\":\"outcome\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"appData\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"isFinal\",\"type\":\"bool\"}],\"internalType\":\"structChannel.State\",\"name\":\"from\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"},{\"components\":[{\"internalType\":\"address[]\",\"name\":\"assets\",\"type\":\"address[]\"},{\"internalType\":\"uint256[][]\",\"name\":\"balances\",\"type\":\"uint256[][]\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"ID\",\"type\":\"bytes32\"},{\"internalType\":\"uint256[]\",\"name\":\"balances\",\"type\":\"uint256[]\"}],\"internalType\":\"structChannel.SubAlloc[]\",\"name\":\"locked\",\"type\":\"tuple[]\"}],\"internalType\":\"structChannel.Allocation\",\"name\":\"outcome\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"appData\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"isFinal\",\"type\":\"bool\"}],\"internalType\":\"structChannel.State\",\"name\":\"to\",\"type\":\"tuple\"},{\"internalType\":\"uint256\",\"name\":\"actorIdx\",\"type\":\"uint256\"}],\"name\":\"validTransition\",\"outputs\":[],\"stateMutability\":\"pure\",\"type\":\"function\"}]"

// TrivialappBin is the compiled bytecode used for deploying new contracts.
var TrivialappBin = "0x608060405234801561001057600080fd5b5061011a806100206000396000f3fe6080604052348015600f57600080fd5b506004361060285760003560e01c8063ec29dd7e14602d575b600080fd5b603c6038366004605a565b603e565b005b50505050565b600060a082840312156054578081fd5b50919050565b60008060008060808587031215606e578384fd5b843567ffffffffffffffff808211156084578586fd5b90860190608082890312156096578586fd5b9094506020860135908082111560aa578485fd5b60b4888389016044565b9450604087013591508082111560c8578384fd5b5060d3878288016044565b94979396509394606001359350505056fea264697066735822122061e642a9d0202df8de51cfa58bb42acb2bc0b8cd6ec4eb2667a539b63177acfe64736f6c63430007040033"

// DeployTrivialapp deploys a new Ethereum contract, binding an instance of Trivialapp to it.
func DeployTrivialapp(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Trivialapp, error) {
	parsed, err := abi.JSON(strings.NewReader(TrivialappABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(TrivialappBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Trivialapp{TrivialappCaller: TrivialappCaller{contract: contract}, TrivialappTransactor: TrivialappTransactor{contract: contract}, TrivialappFilterer: TrivialappFilterer{contract: contract}}, nil
}

// Trivialapp is an auto generated Go binding around an Ethereum contract.
type Trivialapp struct {
	TrivialappCaller     // Read-only binding to the contract
	TrivialappTransactor // Write-only binding to the contract
	TrivialappFilterer   // Log filterer for contract events
}

// TrivialappCaller is an auto generated read-only Go binding around an Ethereum contract.
type TrivialappCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TrivialappTransactor is an auto generated write-only Go binding around an Ethereum contract.
type TrivialappTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TrivialappFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type TrivialappFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TrivialappSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type TrivialappSession struct {
	Contract     *Trivialapp       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// TrivialappCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type TrivialappCallerSession struct {
	Contract *TrivialappCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// TrivialappTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type TrivialappTransactorSession struct {
	Contract     *TrivialappTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// TrivialappRaw is an auto generated low-level Go binding around an Ethereum contract.
type TrivialappRaw struct {
	Contract *Trivialapp // Generic contract binding to access the raw methods on
}

// TrivialappCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type TrivialappCallerRaw struct {
	Contract *TrivialappCaller // Generic read-only contract binding to access the raw methods on
}

// TrivialappTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type TrivialappTransactorRaw struct {
	Contract *TrivialappTransactor // Generic write-only contract binding to access the raw methods on
}

// NewTrivialapp creates a new instance of Trivialapp, bound to a specific deployed contract.
func NewTrivialapp(address common.Address, backend bind.ContractBackend) (*Trivialapp, error) {
	contract, err := bindTrivialapp(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Trivialapp{TrivialappCaller: TrivialappCaller{contract: contract}, TrivialappTransactor: TrivialappTransactor{contract: contract}, TrivialappFilterer: TrivialappFilterer{contract: contract}}, nil
}

// NewTrivialappCaller creates a new read-only instance of Trivialapp, bound to a specific deployed contract.
func NewTrivialappCaller(address common.Address, caller bind.ContractCaller) (*TrivialappCaller, error) {
	contract, err := bindTrivialapp(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &TrivialappCaller{contract: contract}, nil
}

// NewTrivialappTransactor creates a new write-only instance of Trivialapp, bound to a specific deployed contract.
func NewTrivialappTransactor(address common.Address, transactor bind.ContractTransactor) (*TrivialappTransactor, error) {
	contract, err := bindTrivialapp(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &TrivialappTransactor{contract: contract}, nil
}

// NewTrivialappFilterer creates a new log filterer instance of Trivialapp, bound to a specific deployed contract.
func NewTrivialappFilterer(address common.Address, filterer bind.ContractFilterer) (*TrivialappFilterer, error) {
	contract, err := bindTrivialapp(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &TrivialappFilterer{contract: contract}, nil
}

// bindTrivialapp binds a generic wrapper to an already deployed contract.
func bindTrivialapp(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(TrivialappABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Trivialapp *TrivialappRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Trivialapp.Contract.TrivialappCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Trivialapp *TrivialappRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Trivialapp.Contract.TrivialappTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Trivialapp *TrivialappRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Trivialapp.Contract.TrivialappTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Trivialapp *TrivialappCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Trivialapp.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Trivialapp *TrivialappTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Trivialapp.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Trivialapp *TrivialappTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Trivialapp.Contract.contract.Transact(opts, method, params...)
}

// ValidTransition is a free data retrieval call binding the contract method 0xec29dd7e.
//
// Solidity: function validTransition((uint256,uint256,address,address[]) params, (bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool) from, (bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool) to, uint256 actorIdx) pure returns()
func (_Trivialapp *TrivialappCaller) ValidTransition(opts *bind.CallOpts, params ChannelParams, from ChannelState, to ChannelState, actorIdx *big.Int) error {
	var out []interface{}
	err := _Trivialapp.contract.Call(opts, &out, "validTransition", params, from, to, actorIdx)

	if err != nil {
		return err
	}

	return err

}

// ValidTransition is a free data retrieval call binding the contract method 0xec29dd7e.
//
// Solidity: function validTransition((uint256,uint256,address,address[]) params, (bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool) from, (bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool) to, uint256 actorIdx) pure returns()
func (_Trivialapp *TrivialappSession) ValidTransition(params ChannelParams, from ChannelState, to ChannelState, actorIdx *big.Int) error {
	return _Trivialapp.Contract.ValidTransition(&_Trivialapp.CallOpts, params, from, to, actorIdx)
}

// ValidTransition is a free data retrieval call binding the contract method 0xec29dd7e.
//
// Solidity: function validTransition((uint256,uint256,address,address[]) params, (bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool) from, (bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool) to, uint256 actorIdx) pure returns()
func (_Trivialapp *TrivialappCallerSession) ValidTransition(params ChannelParams, from ChannelState, to ChannelState, actorIdx *big.Int) error {
	return _Trivialapp.Contract.ValidTransition(&_Trivialapp.CallOpts, params, from, to, actorIdx)
}
