// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package peruntoken

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

// PeruntokenABI is the input ABI used to generate the binding from.
const PeruntokenABI = "[{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"accounts\",\"type\":\"address[]\"},{\"internalType\":\"uint256\",\"name\":\"initBalance\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"subtractedValue\",\"type\":\"uint256\"}],\"name\":\"decreaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"addedValue\",\"type\":\"uint256\"}],\"name\":\"increaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// PeruntokenBin is the compiled bytecode used for deploying new contracts.
var PeruntokenBin = "0x60806040523480156200001157600080fd5b5060405162000e2c38038062000e2c833981016040819052620000349162000335565b6040518060400160405280600a8152602001692832b93ab72a37b5b2b760b11b8152506040518060400160405280600381526020016228292760e91b81525081600390805190602001906200008b9291906200026c565b508051620000a19060049060208401906200026c565b50506005805460ff191660121790555060005b8251811015620000ed57620000e4838281518110620000cf57fe5b602002602001015183620000f660201b60201c565b600101620000b4565b50505062000413565b6001600160a01b03821662000152576040805162461bcd60e51b815260206004820152601f60248201527f45524332303a206d696e7420746f20746865207a65726f206164647265737300604482015290519081900360640190fd5b620001606000838362000205565b6200017c816002546200020a60201b6200044e1790919060201c565b6002556001600160a01b03821660009081526020818152604090912054620001af9183906200044e6200020a821b17901c565b6001600160a01b0383166000818152602081815260408083209490945583518581529351929391927fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef9281900390910190a35050565b505050565b60008282018381101562000265576040805162461bcd60e51b815260206004820152601b60248201527f536166654d6174683a206164646974696f6e206f766572666c6f770000000000604482015290519081900360640190fd5b9392505050565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282620002a45760008555620002ef565b82601f10620002bf57805160ff1916838001178555620002ef565b82800160010185558215620002ef579182015b82811115620002ef578251825591602001919060010190620002d2565b50620002fd92915062000301565b5090565b5b80821115620002fd576000815560010162000302565b80516001600160a01b03811681146200033057600080fd5b919050565b6000806040838503121562000348578182fd5b82516001600160401b03808211156200035f578384fd5b818501915085601f83011262000373578384fd5b8151818111156200038057fe5b6020915081810262000394838201620003ef565b8281528381019085850183870186018b1015620003af578889fd5b8896505b84871015620003dc57620003c78162000318565b835260019690960195918501918501620003b3565b5097909301519698969750505050505050565b6040518181016001600160401b03811182821017156200040b57fe5b604052919050565b610a0980620004236000396000f3fe608060405234801561001057600080fd5b50600436106100a95760003560e01c80633950935111610071578063395093511461012957806370a082311461013c57806395d89b411461014f578063a457c2d714610157578063a9059cbb1461016a578063dd62ed3e1461017d576100a9565b806306fdde03146100ae578063095ea7b3146100cc57806318160ddd146100ec57806323b872dd14610101578063313ce56714610114575b600080fd5b6100b6610190565b6040516100c39190610868565b60405180910390f35b6100df6100da366004610834565b610226565b6040516100c3919061085d565b6100f4610243565b6040516100c391906108bb565b6100df61010f3660046107f9565b610249565b61011c6102d0565b6040516100c391906108c4565b6100df610137366004610834565b6102d9565b6100f461014a3660046107ad565b610327565b6100b6610346565b6100df610165366004610834565b6103a7565b6100df610178366004610834565b61040f565b6100f461018b3660046107c7565b610423565b60038054604080516020601f600260001961010060018816150201909516949094049384018190048102820181019092528281526060939092909183018282801561021c5780601f106101f15761010080835404028352916020019161021c565b820191906000526020600020905b8154815290600101906020018083116101ff57829003601f168201915b5050505050905090565b600061023a6102336104af565b84846104b3565b50600192915050565b60025490565b600061025684848461059f565b6102c6846102626104af565b6102c18560405180606001604052806028815260200161093e602891396001600160a01b038a166000908152600160205260408120906102a06104af565b6001600160a01b0316815260208101919091526040016000205491906106fa565b6104b3565b5060019392505050565b60055460ff1690565b600061023a6102e66104af565b846102c185600160006102f76104af565b6001600160a01b03908116825260208083019390935260409182016000908120918c16815292529020549061044e565b6001600160a01b0381166000908152602081905260409020545b919050565b60048054604080516020601f600260001961010060018816150201909516949094049384018190048102820181019092528281526060939092909183018282801561021c5780601f106101f15761010080835404028352916020019161021c565b600061023a6103b46104af565b846102c1856040518060600160405280602581526020016109af60259139600160006103de6104af565b6001600160a01b03908116825260208083019390935260409182016000908120918d168152925290205491906106fa565b600061023a61041c6104af565b848461059f565b6001600160a01b03918216600090815260016020908152604080832093909416825291909152205490565b6000828201838110156104a8576040805162461bcd60e51b815260206004820152601b60248201527f536166654d6174683a206164646974696f6e206f766572666c6f770000000000604482015290519081900360640190fd5b9392505050565b3390565b6001600160a01b0383166104f85760405162461bcd60e51b815260040180806020018281038252602481526020018061098b6024913960400191505060405180910390fd5b6001600160a01b03821661053d5760405162461bcd60e51b81526004018080602001828103825260228152602001806108f66022913960400191505060405180910390fd5b6001600160a01b03808416600081815260016020908152604080832094871680845294825291829020859055815185815291517f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b9259281900390910190a3505050565b6001600160a01b0383166105e45760405162461bcd60e51b81526004018080602001828103825260258152602001806109666025913960400191505060405180910390fd5b6001600160a01b0382166106295760405162461bcd60e51b81526004018080602001828103825260238152602001806108d36023913960400191505060405180910390fd5b610634838383610791565b61067181604051806060016040528060268152602001610918602691396001600160a01b03861660009081526020819052604090205491906106fa565b6001600160a01b0380851660009081526020819052604080822093909355908416815220546106a0908261044e565b6001600160a01b038084166000818152602081815260409182902094909455805185815290519193928716927fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef92918290030190a3505050565b600081848411156107895760405162461bcd60e51b81526004018080602001828103825283818151815260200191508051906020019080838360005b8381101561074e578181015183820152602001610736565b50505050905090810190601f16801561077b5780820380516001836020036101000a031916815260200191505b509250505060405180910390fd5b505050900390565b505050565b80356001600160a01b038116811461034157600080fd5b6000602082840312156107be578081fd5b6104a882610796565b600080604083850312156107d9578081fd5b6107e283610796565b91506107f060208401610796565b90509250929050565b60008060006060848603121561080d578081fd5b61081684610796565b925061082460208501610796565b9150604084013590509250925092565b60008060408385031215610846578182fd5b61084f83610796565b946020939093013593505050565b901515815260200190565b6000602080835283518082850152825b8181101561089457858101830151858201604001528201610878565b818111156108a55783604083870101525b50601f01601f1916929092016040019392505050565b90815260200190565b60ff9190911681526020019056fe45524332303a207472616e7366657220746f20746865207a65726f206164647265737345524332303a20617070726f766520746f20746865207a65726f206164647265737345524332303a207472616e7366657220616d6f756e7420657863656564732062616c616e636545524332303a207472616e7366657220616d6f756e74206578636565647320616c6c6f77616e636545524332303a207472616e736665722066726f6d20746865207a65726f206164647265737345524332303a20617070726f76652066726f6d20746865207a65726f206164647265737345524332303a2064656372656173656420616c6c6f77616e63652062656c6f77207a65726fa264697066735822122037474073085c66a29bfbd0f5f32114fd1ca80f646beac1ebc808dda1667e7bc064736f6c63430007040033"

// DeployPeruntoken deploys a new Ethereum contract, binding an instance of Peruntoken to it.
func DeployPeruntoken(auth *bind.TransactOpts, backend bind.ContractBackend, accounts []common.Address, initBalance *big.Int) (common.Address, *types.Transaction, *Peruntoken, error) {
	parsed, err := abi.JSON(strings.NewReader(PeruntokenABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(PeruntokenBin), backend, accounts, initBalance)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Peruntoken{PeruntokenCaller: PeruntokenCaller{contract: contract}, PeruntokenTransactor: PeruntokenTransactor{contract: contract}, PeruntokenFilterer: PeruntokenFilterer{contract: contract}}, nil
}

// Peruntoken is an auto generated Go binding around an Ethereum contract.
type Peruntoken struct {
	PeruntokenCaller     // Read-only binding to the contract
	PeruntokenTransactor // Write-only binding to the contract
	PeruntokenFilterer   // Log filterer for contract events
}

// PeruntokenCaller is an auto generated read-only Go binding around an Ethereum contract.
type PeruntokenCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PeruntokenTransactor is an auto generated write-only Go binding around an Ethereum contract.
type PeruntokenTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PeruntokenFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type PeruntokenFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PeruntokenSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type PeruntokenSession struct {
	Contract     *Peruntoken       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// PeruntokenCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type PeruntokenCallerSession struct {
	Contract *PeruntokenCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// PeruntokenTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type PeruntokenTransactorSession struct {
	Contract     *PeruntokenTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// PeruntokenRaw is an auto generated low-level Go binding around an Ethereum contract.
type PeruntokenRaw struct {
	Contract *Peruntoken // Generic contract binding to access the raw methods on
}

// PeruntokenCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type PeruntokenCallerRaw struct {
	Contract *PeruntokenCaller // Generic read-only contract binding to access the raw methods on
}

// PeruntokenTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type PeruntokenTransactorRaw struct {
	Contract *PeruntokenTransactor // Generic write-only contract binding to access the raw methods on
}

// NewPeruntoken creates a new instance of Peruntoken, bound to a specific deployed contract.
func NewPeruntoken(address common.Address, backend bind.ContractBackend) (*Peruntoken, error) {
	contract, err := bindPeruntoken(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Peruntoken{PeruntokenCaller: PeruntokenCaller{contract: contract}, PeruntokenTransactor: PeruntokenTransactor{contract: contract}, PeruntokenFilterer: PeruntokenFilterer{contract: contract}}, nil
}

// NewPeruntokenCaller creates a new read-only instance of Peruntoken, bound to a specific deployed contract.
func NewPeruntokenCaller(address common.Address, caller bind.ContractCaller) (*PeruntokenCaller, error) {
	contract, err := bindPeruntoken(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &PeruntokenCaller{contract: contract}, nil
}

// NewPeruntokenTransactor creates a new write-only instance of Peruntoken, bound to a specific deployed contract.
func NewPeruntokenTransactor(address common.Address, transactor bind.ContractTransactor) (*PeruntokenTransactor, error) {
	contract, err := bindPeruntoken(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &PeruntokenTransactor{contract: contract}, nil
}

// NewPeruntokenFilterer creates a new log filterer instance of Peruntoken, bound to a specific deployed contract.
func NewPeruntokenFilterer(address common.Address, filterer bind.ContractFilterer) (*PeruntokenFilterer, error) {
	contract, err := bindPeruntoken(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &PeruntokenFilterer{contract: contract}, nil
}

// bindPeruntoken binds a generic wrapper to an already deployed contract.
func bindPeruntoken(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(PeruntokenABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Peruntoken *PeruntokenRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Peruntoken.Contract.PeruntokenCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Peruntoken *PeruntokenRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Peruntoken.Contract.PeruntokenTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Peruntoken *PeruntokenRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Peruntoken.Contract.PeruntokenTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Peruntoken *PeruntokenCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Peruntoken.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Peruntoken *PeruntokenTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Peruntoken.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Peruntoken *PeruntokenTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Peruntoken.Contract.contract.Transact(opts, method, params...)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_Peruntoken *PeruntokenCaller) Allowance(opts *bind.CallOpts, owner common.Address, spender common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Peruntoken.contract.Call(opts, &out, "allowance", owner, spender)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_Peruntoken *PeruntokenSession) Allowance(owner common.Address, spender common.Address) (*big.Int, error) {
	return _Peruntoken.Contract.Allowance(&_Peruntoken.CallOpts, owner, spender)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_Peruntoken *PeruntokenCallerSession) Allowance(owner common.Address, spender common.Address) (*big.Int, error) {
	return _Peruntoken.Contract.Allowance(&_Peruntoken.CallOpts, owner, spender)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_Peruntoken *PeruntokenCaller) BalanceOf(opts *bind.CallOpts, account common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Peruntoken.contract.Call(opts, &out, "balanceOf", account)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_Peruntoken *PeruntokenSession) BalanceOf(account common.Address) (*big.Int, error) {
	return _Peruntoken.Contract.BalanceOf(&_Peruntoken.CallOpts, account)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_Peruntoken *PeruntokenCallerSession) BalanceOf(account common.Address) (*big.Int, error) {
	return _Peruntoken.Contract.BalanceOf(&_Peruntoken.CallOpts, account)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_Peruntoken *PeruntokenCaller) Decimals(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _Peruntoken.contract.Call(opts, &out, "decimals")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_Peruntoken *PeruntokenSession) Decimals() (uint8, error) {
	return _Peruntoken.Contract.Decimals(&_Peruntoken.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_Peruntoken *PeruntokenCallerSession) Decimals() (uint8, error) {
	return _Peruntoken.Contract.Decimals(&_Peruntoken.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_Peruntoken *PeruntokenCaller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Peruntoken.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_Peruntoken *PeruntokenSession) Name() (string, error) {
	return _Peruntoken.Contract.Name(&_Peruntoken.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_Peruntoken *PeruntokenCallerSession) Name() (string, error) {
	return _Peruntoken.Contract.Name(&_Peruntoken.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_Peruntoken *PeruntokenCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Peruntoken.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_Peruntoken *PeruntokenSession) Symbol() (string, error) {
	return _Peruntoken.Contract.Symbol(&_Peruntoken.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_Peruntoken *PeruntokenCallerSession) Symbol() (string, error) {
	return _Peruntoken.Contract.Symbol(&_Peruntoken.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_Peruntoken *PeruntokenCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Peruntoken.contract.Call(opts, &out, "totalSupply")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_Peruntoken *PeruntokenSession) TotalSupply() (*big.Int, error) {
	return _Peruntoken.Contract.TotalSupply(&_Peruntoken.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_Peruntoken *PeruntokenCallerSession) TotalSupply() (*big.Int, error) {
	return _Peruntoken.Contract.TotalSupply(&_Peruntoken.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_Peruntoken *PeruntokenTransactor) Approve(opts *bind.TransactOpts, spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Peruntoken.contract.Transact(opts, "approve", spender, amount)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_Peruntoken *PeruntokenSession) Approve(spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Peruntoken.Contract.Approve(&_Peruntoken.TransactOpts, spender, amount)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_Peruntoken *PeruntokenTransactorSession) Approve(spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Peruntoken.Contract.Approve(&_Peruntoken.TransactOpts, spender, amount)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address spender, uint256 subtractedValue) returns(bool)
func (_Peruntoken *PeruntokenTransactor) DecreaseAllowance(opts *bind.TransactOpts, spender common.Address, subtractedValue *big.Int) (*types.Transaction, error) {
	return _Peruntoken.contract.Transact(opts, "decreaseAllowance", spender, subtractedValue)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address spender, uint256 subtractedValue) returns(bool)
func (_Peruntoken *PeruntokenSession) DecreaseAllowance(spender common.Address, subtractedValue *big.Int) (*types.Transaction, error) {
	return _Peruntoken.Contract.DecreaseAllowance(&_Peruntoken.TransactOpts, spender, subtractedValue)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address spender, uint256 subtractedValue) returns(bool)
func (_Peruntoken *PeruntokenTransactorSession) DecreaseAllowance(spender common.Address, subtractedValue *big.Int) (*types.Transaction, error) {
	return _Peruntoken.Contract.DecreaseAllowance(&_Peruntoken.TransactOpts, spender, subtractedValue)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address spender, uint256 addedValue) returns(bool)
func (_Peruntoken *PeruntokenTransactor) IncreaseAllowance(opts *bind.TransactOpts, spender common.Address, addedValue *big.Int) (*types.Transaction, error) {
	return _Peruntoken.contract.Transact(opts, "increaseAllowance", spender, addedValue)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address spender, uint256 addedValue) returns(bool)
func (_Peruntoken *PeruntokenSession) IncreaseAllowance(spender common.Address, addedValue *big.Int) (*types.Transaction, error) {
	return _Peruntoken.Contract.IncreaseAllowance(&_Peruntoken.TransactOpts, spender, addedValue)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address spender, uint256 addedValue) returns(bool)
func (_Peruntoken *PeruntokenTransactorSession) IncreaseAllowance(spender common.Address, addedValue *big.Int) (*types.Transaction, error) {
	return _Peruntoken.Contract.IncreaseAllowance(&_Peruntoken.TransactOpts, spender, addedValue)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address recipient, uint256 amount) returns(bool)
func (_Peruntoken *PeruntokenTransactor) Transfer(opts *bind.TransactOpts, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Peruntoken.contract.Transact(opts, "transfer", recipient, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address recipient, uint256 amount) returns(bool)
func (_Peruntoken *PeruntokenSession) Transfer(recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Peruntoken.Contract.Transfer(&_Peruntoken.TransactOpts, recipient, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address recipient, uint256 amount) returns(bool)
func (_Peruntoken *PeruntokenTransactorSession) Transfer(recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Peruntoken.Contract.Transfer(&_Peruntoken.TransactOpts, recipient, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address sender, address recipient, uint256 amount) returns(bool)
func (_Peruntoken *PeruntokenTransactor) TransferFrom(opts *bind.TransactOpts, sender common.Address, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Peruntoken.contract.Transact(opts, "transferFrom", sender, recipient, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address sender, address recipient, uint256 amount) returns(bool)
func (_Peruntoken *PeruntokenSession) TransferFrom(sender common.Address, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Peruntoken.Contract.TransferFrom(&_Peruntoken.TransactOpts, sender, recipient, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address sender, address recipient, uint256 amount) returns(bool)
func (_Peruntoken *PeruntokenTransactorSession) TransferFrom(sender common.Address, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Peruntoken.Contract.TransferFrom(&_Peruntoken.TransactOpts, sender, recipient, amount)
}

// PeruntokenApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the Peruntoken contract.
type PeruntokenApprovalIterator struct {
	Event *PeruntokenApproval // Event containing the contract specifics and raw log

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
func (it *PeruntokenApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PeruntokenApproval)
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
		it.Event = new(PeruntokenApproval)
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
func (it *PeruntokenApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PeruntokenApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PeruntokenApproval represents a Approval event raised by the Peruntoken contract.
type PeruntokenApproval struct {
	Owner   common.Address
	Spender common.Address
	Value   *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_Peruntoken *PeruntokenFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, spender []common.Address) (*PeruntokenApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _Peruntoken.contract.FilterLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return &PeruntokenApprovalIterator{contract: _Peruntoken.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_Peruntoken *PeruntokenFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *PeruntokenApproval, owner []common.Address, spender []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _Peruntoken.contract.WatchLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PeruntokenApproval)
				if err := _Peruntoken.contract.UnpackLog(event, "Approval", log); err != nil {
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

// ParseApproval is a log parse operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_Peruntoken *PeruntokenFilterer) ParseApproval(log types.Log) (*PeruntokenApproval, error) {
	event := new(PeruntokenApproval)
	if err := _Peruntoken.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PeruntokenTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the Peruntoken contract.
type PeruntokenTransferIterator struct {
	Event *PeruntokenTransfer // Event containing the contract specifics and raw log

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
func (it *PeruntokenTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PeruntokenTransfer)
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
		it.Event = new(PeruntokenTransfer)
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
func (it *PeruntokenTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PeruntokenTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PeruntokenTransfer represents a Transfer event raised by the Peruntoken contract.
type PeruntokenTransfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_Peruntoken *PeruntokenFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*PeruntokenTransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _Peruntoken.contract.FilterLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &PeruntokenTransferIterator{contract: _Peruntoken.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_Peruntoken *PeruntokenFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *PeruntokenTransfer, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _Peruntoken.contract.WatchLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PeruntokenTransfer)
				if err := _Peruntoken.contract.UnpackLog(event, "Transfer", log); err != nil {
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

// ParseTransfer is a log parse operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_Peruntoken *PeruntokenFilterer) ParseTransfer(log types.Log) (*PeruntokenTransfer, error) {
	event := new(PeruntokenTransfer)
	if err := _Peruntoken.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
