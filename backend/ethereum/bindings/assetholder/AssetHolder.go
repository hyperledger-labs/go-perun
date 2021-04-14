// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package assetholder

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

// AssetHolderWithdrawalAuth is an auto generated low-level Go binding around an user-defined struct.
type AssetHolderWithdrawalAuth struct {
	ChannelID   [32]byte
	Participant common.Address
	Receiver    common.Address
	Amount      *big.Int
}

// AssetholderABI is the input ABI used to generate the binding from.
const AssetholderABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"fundingID\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Deposited\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"}],\"name\":\"OutcomeSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"fundingID\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"receiver\",\"type\":\"address\"}],\"name\":\"Withdrawn\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"adjudicator\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"fundingID\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"deposit\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"holdings\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"},{\"internalType\":\"address[]\",\"name\":\"parts\",\"type\":\"address[]\"},{\"internalType\":\"uint256[]\",\"name\":\"newBals\",\"type\":\"uint256[]\"}],\"name\":\"setOutcome\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"settled\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"participant\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"receiver\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structAssetHolder.WithdrawalAuth\",\"name\":\"authorization\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"withdraw\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// Assetholder is an auto generated Go binding around an Ethereum contract.
type Assetholder struct {
	AssetholderCaller     // Read-only binding to the contract
	AssetholderTransactor // Write-only binding to the contract
	AssetholderFilterer   // Log filterer for contract events
}

// AssetholderCaller is an auto generated read-only Go binding around an Ethereum contract.
type AssetholderCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AssetholderTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AssetholderTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AssetholderFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AssetholderFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AssetholderSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AssetholderSession struct {
	Contract     *Assetholder      // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// AssetholderCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AssetholderCallerSession struct {
	Contract *AssetholderCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts      // Call options to use throughout this session
}

// AssetholderTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AssetholderTransactorSession struct {
	Contract     *AssetholderTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// AssetholderRaw is an auto generated low-level Go binding around an Ethereum contract.
type AssetholderRaw struct {
	Contract *Assetholder // Generic contract binding to access the raw methods on
}

// AssetholderCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AssetholderCallerRaw struct {
	Contract *AssetholderCaller // Generic read-only contract binding to access the raw methods on
}

// AssetholderTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AssetholderTransactorRaw struct {
	Contract *AssetholderTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAssetholder creates a new instance of Assetholder, bound to a specific deployed contract.
func NewAssetholder(address common.Address, backend bind.ContractBackend) (*Assetholder, error) {
	contract, err := bindAssetholder(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Assetholder{AssetholderCaller: AssetholderCaller{contract: contract}, AssetholderTransactor: AssetholderTransactor{contract: contract}, AssetholderFilterer: AssetholderFilterer{contract: contract}}, nil
}

// NewAssetholderCaller creates a new read-only instance of Assetholder, bound to a specific deployed contract.
func NewAssetholderCaller(address common.Address, caller bind.ContractCaller) (*AssetholderCaller, error) {
	contract, err := bindAssetholder(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AssetholderCaller{contract: contract}, nil
}

// NewAssetholderTransactor creates a new write-only instance of Assetholder, bound to a specific deployed contract.
func NewAssetholderTransactor(address common.Address, transactor bind.ContractTransactor) (*AssetholderTransactor, error) {
	contract, err := bindAssetholder(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AssetholderTransactor{contract: contract}, nil
}

// NewAssetholderFilterer creates a new log filterer instance of Assetholder, bound to a specific deployed contract.
func NewAssetholderFilterer(address common.Address, filterer bind.ContractFilterer) (*AssetholderFilterer, error) {
	contract, err := bindAssetholder(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AssetholderFilterer{contract: contract}, nil
}

// bindAssetholder binds a generic wrapper to an already deployed contract.
func bindAssetholder(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(AssetholderABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Assetholder *AssetholderRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Assetholder.Contract.AssetholderCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Assetholder *AssetholderRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Assetholder.Contract.AssetholderTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Assetholder *AssetholderRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Assetholder.Contract.AssetholderTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Assetholder *AssetholderCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Assetholder.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Assetholder *AssetholderTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Assetholder.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Assetholder *AssetholderTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Assetholder.Contract.contract.Transact(opts, method, params...)
}

// Adjudicator is a free data retrieval call binding the contract method 0x53c2ed8e.
//
// Solidity: function adjudicator() view returns(address)
func (_Assetholder *AssetholderCaller) Adjudicator(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Assetholder.contract.Call(opts, &out, "adjudicator")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Adjudicator is a free data retrieval call binding the contract method 0x53c2ed8e.
//
// Solidity: function adjudicator() view returns(address)
func (_Assetholder *AssetholderSession) Adjudicator() (common.Address, error) {
	return _Assetholder.Contract.Adjudicator(&_Assetholder.CallOpts)
}

// Adjudicator is a free data retrieval call binding the contract method 0x53c2ed8e.
//
// Solidity: function adjudicator() view returns(address)
func (_Assetholder *AssetholderCallerSession) Adjudicator() (common.Address, error) {
	return _Assetholder.Contract.Adjudicator(&_Assetholder.CallOpts)
}

// Holdings is a free data retrieval call binding the contract method 0xae9ee18c.
//
// Solidity: function holdings(bytes32 ) view returns(uint256)
func (_Assetholder *AssetholderCaller) Holdings(opts *bind.CallOpts, arg0 [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _Assetholder.contract.Call(opts, &out, "holdings", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Holdings is a free data retrieval call binding the contract method 0xae9ee18c.
//
// Solidity: function holdings(bytes32 ) view returns(uint256)
func (_Assetholder *AssetholderSession) Holdings(arg0 [32]byte) (*big.Int, error) {
	return _Assetholder.Contract.Holdings(&_Assetholder.CallOpts, arg0)
}

// Holdings is a free data retrieval call binding the contract method 0xae9ee18c.
//
// Solidity: function holdings(bytes32 ) view returns(uint256)
func (_Assetholder *AssetholderCallerSession) Holdings(arg0 [32]byte) (*big.Int, error) {
	return _Assetholder.Contract.Holdings(&_Assetholder.CallOpts, arg0)
}

// Settled is a free data retrieval call binding the contract method 0xd945af1d.
//
// Solidity: function settled(bytes32 ) view returns(bool)
func (_Assetholder *AssetholderCaller) Settled(opts *bind.CallOpts, arg0 [32]byte) (bool, error) {
	var out []interface{}
	err := _Assetholder.contract.Call(opts, &out, "settled", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Settled is a free data retrieval call binding the contract method 0xd945af1d.
//
// Solidity: function settled(bytes32 ) view returns(bool)
func (_Assetholder *AssetholderSession) Settled(arg0 [32]byte) (bool, error) {
	return _Assetholder.Contract.Settled(&_Assetholder.CallOpts, arg0)
}

// Settled is a free data retrieval call binding the contract method 0xd945af1d.
//
// Solidity: function settled(bytes32 ) view returns(bool)
func (_Assetholder *AssetholderCallerSession) Settled(arg0 [32]byte) (bool, error) {
	return _Assetholder.Contract.Settled(&_Assetholder.CallOpts, arg0)
}

// Deposit is a paid mutator transaction binding the contract method 0x1de26e16.
//
// Solidity: function deposit(bytes32 fundingID, uint256 amount) payable returns()
func (_Assetholder *AssetholderTransactor) Deposit(opts *bind.TransactOpts, fundingID [32]byte, amount *big.Int) (*types.Transaction, error) {
	return _Assetholder.contract.Transact(opts, "deposit", fundingID, amount)
}

// Deposit is a paid mutator transaction binding the contract method 0x1de26e16.
//
// Solidity: function deposit(bytes32 fundingID, uint256 amount) payable returns()
func (_Assetholder *AssetholderSession) Deposit(fundingID [32]byte, amount *big.Int) (*types.Transaction, error) {
	return _Assetholder.Contract.Deposit(&_Assetholder.TransactOpts, fundingID, amount)
}

// Deposit is a paid mutator transaction binding the contract method 0x1de26e16.
//
// Solidity: function deposit(bytes32 fundingID, uint256 amount) payable returns()
func (_Assetholder *AssetholderTransactorSession) Deposit(fundingID [32]byte, amount *big.Int) (*types.Transaction, error) {
	return _Assetholder.Contract.Deposit(&_Assetholder.TransactOpts, fundingID, amount)
}

// SetOutcome is a paid mutator transaction binding the contract method 0xfc79a66d.
//
// Solidity: function setOutcome(bytes32 channelID, address[] parts, uint256[] newBals) returns()
func (_Assetholder *AssetholderTransactor) SetOutcome(opts *bind.TransactOpts, channelID [32]byte, parts []common.Address, newBals []*big.Int) (*types.Transaction, error) {
	return _Assetholder.contract.Transact(opts, "setOutcome", channelID, parts, newBals)
}

// SetOutcome is a paid mutator transaction binding the contract method 0xfc79a66d.
//
// Solidity: function setOutcome(bytes32 channelID, address[] parts, uint256[] newBals) returns()
func (_Assetholder *AssetholderSession) SetOutcome(channelID [32]byte, parts []common.Address, newBals []*big.Int) (*types.Transaction, error) {
	return _Assetholder.Contract.SetOutcome(&_Assetholder.TransactOpts, channelID, parts, newBals)
}

// SetOutcome is a paid mutator transaction binding the contract method 0xfc79a66d.
//
// Solidity: function setOutcome(bytes32 channelID, address[] parts, uint256[] newBals) returns()
func (_Assetholder *AssetholderTransactorSession) SetOutcome(channelID [32]byte, parts []common.Address, newBals []*big.Int) (*types.Transaction, error) {
	return _Assetholder.Contract.SetOutcome(&_Assetholder.TransactOpts, channelID, parts, newBals)
}

// Withdraw is a paid mutator transaction binding the contract method 0x4ed4283c.
//
// Solidity: function withdraw((bytes32,address,address,uint256) authorization, bytes signature) returns()
func (_Assetholder *AssetholderTransactor) Withdraw(opts *bind.TransactOpts, authorization AssetHolderWithdrawalAuth, signature []byte) (*types.Transaction, error) {
	return _Assetholder.contract.Transact(opts, "withdraw", authorization, signature)
}

// Withdraw is a paid mutator transaction binding the contract method 0x4ed4283c.
//
// Solidity: function withdraw((bytes32,address,address,uint256) authorization, bytes signature) returns()
func (_Assetholder *AssetholderSession) Withdraw(authorization AssetHolderWithdrawalAuth, signature []byte) (*types.Transaction, error) {
	return _Assetholder.Contract.Withdraw(&_Assetholder.TransactOpts, authorization, signature)
}

// Withdraw is a paid mutator transaction binding the contract method 0x4ed4283c.
//
// Solidity: function withdraw((bytes32,address,address,uint256) authorization, bytes signature) returns()
func (_Assetholder *AssetholderTransactorSession) Withdraw(authorization AssetHolderWithdrawalAuth, signature []byte) (*types.Transaction, error) {
	return _Assetholder.Contract.Withdraw(&_Assetholder.TransactOpts, authorization, signature)
}

// AssetholderDepositedIterator is returned from FilterDeposited and is used to iterate over the raw logs and unpacked data for Deposited events raised by the Assetholder contract.
type AssetholderDepositedIterator struct {
	Event *AssetholderDeposited // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AssetholderDepositedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AssetholderDeposited)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AssetholderDeposited)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AssetholderDepositedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AssetholderDepositedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AssetholderDeposited represents a Deposited event raised by the Assetholder contract.
type AssetholderDeposited struct {
	FundingID [32]byte
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterDeposited is a free log retrieval operation binding the contract event 0xcd2fe07293de5928c5df9505b65a8d6506f8668dfe81af09090920687edc48a9.
//
// Solidity: event Deposited(bytes32 indexed fundingID, uint256 amount)
func (_Assetholder *AssetholderFilterer) FilterDeposited(opts *bind.FilterOpts, fundingID [][32]byte) (*AssetholderDepositedIterator, error) {

	var fundingIDRule []interface{}
	for _, fundingIDItem := range fundingID {
		fundingIDRule = append(fundingIDRule, fundingIDItem)
	}

	logs, sub, err := _Assetholder.contract.FilterLogs(opts, "Deposited", fundingIDRule)
	if err != nil {
		return nil, err
	}
	return &AssetholderDepositedIterator{contract: _Assetholder.contract, event: "Deposited", logs: logs, sub: sub}, nil
}

// WatchDeposited is a free log subscription operation binding the contract event 0xcd2fe07293de5928c5df9505b65a8d6506f8668dfe81af09090920687edc48a9.
//
// Solidity: event Deposited(bytes32 indexed fundingID, uint256 amount)
func (_Assetholder *AssetholderFilterer) WatchDeposited(opts *bind.WatchOpts, sink chan<- *AssetholderDeposited, fundingID [][32]byte) (event.Subscription, error) {

	var fundingIDRule []interface{}
	for _, fundingIDItem := range fundingID {
		fundingIDRule = append(fundingIDRule, fundingIDItem)
	}

	logs, sub, err := _Assetholder.contract.WatchLogs(opts, "Deposited", fundingIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AssetholderDeposited)
				if err := _Assetholder.contract.UnpackLog(event, "Deposited", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDeposited is a log parse operation binding the contract event 0xcd2fe07293de5928c5df9505b65a8d6506f8668dfe81af09090920687edc48a9.
//
// Solidity: event Deposited(bytes32 indexed fundingID, uint256 amount)
func (_Assetholder *AssetholderFilterer) ParseDeposited(log types.Log) (*AssetholderDeposited, error) {
	event := new(AssetholderDeposited)
	if err := _Assetholder.contract.UnpackLog(event, "Deposited", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AssetholderOutcomeSetIterator is returned from FilterOutcomeSet and is used to iterate over the raw logs and unpacked data for OutcomeSet events raised by the Assetholder contract.
type AssetholderOutcomeSetIterator struct {
	Event *AssetholderOutcomeSet // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AssetholderOutcomeSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AssetholderOutcomeSet)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AssetholderOutcomeSet)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AssetholderOutcomeSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AssetholderOutcomeSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AssetholderOutcomeSet represents a OutcomeSet event raised by the Assetholder contract.
type AssetholderOutcomeSet struct {
	ChannelID [32]byte
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterOutcomeSet is a free log retrieval operation binding the contract event 0xef898d6cd3395b6dfe67a3c1923e5c726c1b154e979fb0a25a9c41d0093168b8.
//
// Solidity: event OutcomeSet(bytes32 indexed channelID)
func (_Assetholder *AssetholderFilterer) FilterOutcomeSet(opts *bind.FilterOpts, channelID [][32]byte) (*AssetholderOutcomeSetIterator, error) {

	var channelIDRule []interface{}
	for _, channelIDItem := range channelID {
		channelIDRule = append(channelIDRule, channelIDItem)
	}

	logs, sub, err := _Assetholder.contract.FilterLogs(opts, "OutcomeSet", channelIDRule)
	if err != nil {
		return nil, err
	}
	return &AssetholderOutcomeSetIterator{contract: _Assetholder.contract, event: "OutcomeSet", logs: logs, sub: sub}, nil
}

// WatchOutcomeSet is a free log subscription operation binding the contract event 0xef898d6cd3395b6dfe67a3c1923e5c726c1b154e979fb0a25a9c41d0093168b8.
//
// Solidity: event OutcomeSet(bytes32 indexed channelID)
func (_Assetholder *AssetholderFilterer) WatchOutcomeSet(opts *bind.WatchOpts, sink chan<- *AssetholderOutcomeSet, channelID [][32]byte) (event.Subscription, error) {

	var channelIDRule []interface{}
	for _, channelIDItem := range channelID {
		channelIDRule = append(channelIDRule, channelIDItem)
	}

	logs, sub, err := _Assetholder.contract.WatchLogs(opts, "OutcomeSet", channelIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AssetholderOutcomeSet)
				if err := _Assetholder.contract.UnpackLog(event, "OutcomeSet", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOutcomeSet is a log parse operation binding the contract event 0xef898d6cd3395b6dfe67a3c1923e5c726c1b154e979fb0a25a9c41d0093168b8.
//
// Solidity: event OutcomeSet(bytes32 indexed channelID)
func (_Assetholder *AssetholderFilterer) ParseOutcomeSet(log types.Log) (*AssetholderOutcomeSet, error) {
	event := new(AssetholderOutcomeSet)
	if err := _Assetholder.contract.UnpackLog(event, "OutcomeSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AssetholderWithdrawnIterator is returned from FilterWithdrawn and is used to iterate over the raw logs and unpacked data for Withdrawn events raised by the Assetholder contract.
type AssetholderWithdrawnIterator struct {
	Event *AssetholderWithdrawn // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AssetholderWithdrawnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AssetholderWithdrawn)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AssetholderWithdrawn)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AssetholderWithdrawnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AssetholderWithdrawnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AssetholderWithdrawn represents a Withdrawn event raised by the Assetholder contract.
type AssetholderWithdrawn struct {
	FundingID [32]byte
	Amount    *big.Int
	Receiver  common.Address
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterWithdrawn is a free log retrieval operation binding the contract event 0xd0b6e7d0170f56c62f87de6a8a47a0ccf41c86ffb5084d399d8eb62e823f2a81.
//
// Solidity: event Withdrawn(bytes32 indexed fundingID, uint256 amount, address receiver)
func (_Assetholder *AssetholderFilterer) FilterWithdrawn(opts *bind.FilterOpts, fundingID [][32]byte) (*AssetholderWithdrawnIterator, error) {

	var fundingIDRule []interface{}
	for _, fundingIDItem := range fundingID {
		fundingIDRule = append(fundingIDRule, fundingIDItem)
	}

	logs, sub, err := _Assetholder.contract.FilterLogs(opts, "Withdrawn", fundingIDRule)
	if err != nil {
		return nil, err
	}
	return &AssetholderWithdrawnIterator{contract: _Assetholder.contract, event: "Withdrawn", logs: logs, sub: sub}, nil
}

// WatchWithdrawn is a free log subscription operation binding the contract event 0xd0b6e7d0170f56c62f87de6a8a47a0ccf41c86ffb5084d399d8eb62e823f2a81.
//
// Solidity: event Withdrawn(bytes32 indexed fundingID, uint256 amount, address receiver)
func (_Assetholder *AssetholderFilterer) WatchWithdrawn(opts *bind.WatchOpts, sink chan<- *AssetholderWithdrawn, fundingID [][32]byte) (event.Subscription, error) {

	var fundingIDRule []interface{}
	for _, fundingIDItem := range fundingID {
		fundingIDRule = append(fundingIDRule, fundingIDItem)
	}

	logs, sub, err := _Assetholder.contract.WatchLogs(opts, "Withdrawn", fundingIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AssetholderWithdrawn)
				if err := _Assetholder.contract.UnpackLog(event, "Withdrawn", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseWithdrawn is a log parse operation binding the contract event 0xd0b6e7d0170f56c62f87de6a8a47a0ccf41c86ffb5084d399d8eb62e823f2a81.
//
// Solidity: event Withdrawn(bytes32 indexed fundingID, uint256 amount, address receiver)
func (_Assetholder *AssetholderFilterer) ParseWithdrawn(log types.Log) (*AssetholderWithdrawn, error) {
	event := new(AssetholderWithdrawn)
	if err := _Assetholder.contract.UnpackLog(event, "Withdrawn", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
