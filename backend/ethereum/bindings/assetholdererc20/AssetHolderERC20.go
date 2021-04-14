// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package assetholdererc20

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

// Assetholdererc20ABI is the input ABI used to generate the binding from.
const Assetholdererc20ABI = "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_adjudicator\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"fundingID\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Deposited\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"}],\"name\":\"OutcomeSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"fundingID\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"receiver\",\"type\":\"address\"}],\"name\":\"Withdrawn\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"adjudicator\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"fundingID\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"deposit\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"holdings\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"},{\"internalType\":\"address[]\",\"name\":\"parts\",\"type\":\"address[]\"},{\"internalType\":\"uint256[]\",\"name\":\"newBals\",\"type\":\"uint256[]\"}],\"name\":\"setOutcome\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"settled\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"token\",\"outputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"participant\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"receiver\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structAssetHolder.WithdrawalAuth\",\"name\":\"authorization\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"withdraw\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// Assetholdererc20Bin is the compiled bytecode used for deploying new contracts.
var Assetholdererc20Bin = "0x60a060405234801561001057600080fd5b5060405161118938038061118983398101604081905261002f91610081565b600280546001600160a01b0319166001600160a01b03939093169290921790915560601b6001600160601b0319166080526100b3565b80516001600160a01b038116811461007c57600080fd5b919050565b60008060408385031215610093578182fd5b61009c83610065565b91506100aa60208401610065565b90509250929050565b60805160601c6110af6100da600039806103b6528061069052806107f652506110af6000f3fe6080604052600436106100705760003560e01c8063ae9ee18c1161004e578063ae9ee18c146100d5578063d945af1d14610102578063fc0c546a1461012f578063fc79a66d1461014457610070565b80631de26e16146100755780634ed4283c1461008a57806353c2ed8e146100aa575b600080fd5b610088610083366004610ca2565b610164565b005b34801561009657600080fd5b506100886100a5366004610cc3565b6101dc565b3480156100b657600080fd5b506100bf61037e565b6040516100cc9190610d48565b60405180910390f35b3480156100e157600080fd5b506100f56100f0366004610c13565b61038d565b6040516100cc9190611014565b34801561010e57600080fd5b5061012261011d366004610c13565b61039f565b6040516100cc9190610d99565b34801561013b57600080fd5b506100bf6103b4565b34801561015057600080fd5b5061008861015f366004610c2b565b6103d8565b61016e82826105f6565b6000828152602081905260409020546101879082610618565b6000838152602081905260409020556101a08282610679565b817fcd2fe07293de5928c5df9505b65a8d6506f8668dfe81af09090920687edc48a9826040516101d09190611014565b60405180910390a25050565b823560009081526001602052604090205460ff166102155760405162461bcd60e51b815260040161020c90610f00565b60405180910390fd5b61027d836040516020016102299190610fc8565b60408051601f198184030181526020601f860181900481028401810190925284835291908590859081908401838280828437600092019190915250610278925050506040870160208801610bd7565b610737565b6102995760405162461bcd60e51b815260040161020c90610ec9565b60006102b584356102b06040870160208801610bd7565b610772565b600081815260208190526040902054909150606085013511156102ea5760405162461bcd60e51b815260040161020c90610e4d565b6102f58484846107a5565b6000818152602081905260409020546103129060608601356107aa565b60008281526020819052604090205561032c8484846107ec565b807fd0b6e7d0170f56c62f87de6a8a47a0ccf41c86ffb5084d399d8eb62e823f2a81606086018035906103629060408901610bd7565b604051610370929190610da4565b60405180910390a250505050565b6002546001600160a01b031681565b60006020819052908152604090205481565b60016020526000908152604090205460ff1681565b7f000000000000000000000000000000000000000000000000000000000000000081565b6002546001600160a01b031633146104025760405162461bcd60e51b815260040161020c90610f83565b8281146104215760405162461bcd60e51b815260040161020c90610e04565b60008581526001602052604090205460ff16156104505760405162461bcd60e51b815260040161020c90610e84565b60008581526020819052604081208054908290559060608567ffffffffffffffff8111801561047e57600080fd5b506040519080825280602002602001820160405280156104a8578160200160208202803683370190505b50905060005b8681101561054f5760006104dd8a8a8a858181106104c857fe5b90506020020160208101906102b09190610bd7565b9050808383815181106104ec57fe5b60200260200101818152505061051d600080838152602001908152602001600020548661061890919063ffffffff16565b945061054487878481811061052e57fe5b905060200201358561061890919063ffffffff16565b9350506001016104ae565b508183106105a95760005b868110156105a75785858281811061056e57fe5b9050602002013560008084848151811061058457fe5b60209081029190910181015182528101919091526040016000205560010161055a565b505b6000888152600160208190526040808320805460ff19169092179091555189917fef898d6cd3395b6dfe67a3c1923e5c726c1b154e979fb0a25a9c41d0093168b891a25050505050505050565b34156106145760405162461bcd60e51b815260040161020c90610dbb565b5050565b600082820183811015610672576040805162461bcd60e51b815260206004820152601b60248201527f536166654d6174683a206164646974696f6e206f766572666c6f770000000000604482015290519081900360640190fd5b9392505050565b6040516323b872dd60e01b81526001600160a01b037f000000000000000000000000000000000000000000000000000000000000000016906323b872dd906106c990339030908690600401610d5c565b602060405180830381600087803b1580156106e357600080fd5b505af11580156106f7573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061071b9190610bf3565b6106145760405162461bcd60e51b815260040161020c90610f2d565b60008061074a85805190602001206108bb565b90506000610758828661090c565b6001600160a01b0390811690851614925050509392505050565b60008282604051602001610787929190610da4565b60405160208183030381529060405280519060200120905092915050565b505050565b600061067283836040518060400160405280601e81526020017f536166654d6174683a207375627472616374696f6e206f766572666c6f770000815250610af7565b6001600160a01b037f00000000000000000000000000000000000000000000000000000000000000001663a9059cbb61082b6060860160408701610bd7565b85606001356040518363ffffffff1660e01b815260040161084d929190610d80565b602060405180830381600087803b15801561086757600080fd5b505af115801561087b573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061089f9190610bf3565b6107a55760405162461bcd60e51b815260040161020c90610f5a565b604080517f19457468657265756d205369676e6564204d6573736167653a0a333200000000602080830191909152603c8083019490945282518083039094018452605c909101909152815191012090565b60008151604114610964576040805162461bcd60e51b815260206004820152601f60248201527f45434453413a20696e76616c6964207369676e6174757265206c656e67746800604482015290519081900360640190fd5b60208201516040830151606084015160001a7f7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a08211156109d55760405162461bcd60e51b81526004018080602001828103825260228152602001806110366022913960400191505060405180910390fd5b8060ff16601b141580156109ed57508060ff16601c14155b15610a295760405162461bcd60e51b81526004018080602001828103825260228152602001806110586022913960400191505060405180910390fd5b600060018783868660405160008152602001604052604051808581526020018460ff1681526020018381526020018281526020019450505050506020604051602081039080840390855afa158015610a85573d6000803e3d6000fd5b5050604051601f1901519150506001600160a01b038116610aed576040805162461bcd60e51b815260206004820152601860248201527f45434453413a20696e76616c6964207369676e61747572650000000000000000604482015290519081900360640190fd5b9695505050505050565b60008184841115610b865760405162461bcd60e51b81526004018080602001828103825283818151815260200191508051906020019080838360005b83811015610b4b578181015183820152602001610b33565b50505050905090810190601f168015610b785780820380516001836020036101000a031916815260200191505b509250505060405180910390fd5b505050900390565b60008083601f840112610b9f578182fd5b50813567ffffffffffffffff811115610bb6578182fd5b6020830191508360208083028501011115610bd057600080fd5b9250929050565b600060208284031215610be8578081fd5b81356106728161101d565b600060208284031215610c04578081fd5b81518015158114610672578182fd5b600060208284031215610c24578081fd5b5035919050565b600080600080600060608688031215610c42578081fd5b85359450602086013567ffffffffffffffff80821115610c60578283fd5b610c6c89838a01610b8e565b90965094506040880135915080821115610c84578283fd5b50610c9188828901610b8e565b969995985093965092949392505050565b60008060408385031215610cb4578182fd5b50508035926020909101359150565b600080600083850360a0811215610cd8578384fd5b6080811215610ce5578384fd5b50839250608084013567ffffffffffffffff80821115610d03578384fd5b818601915086601f830112610d16578384fd5b813581811115610d24578485fd5b876020828501011115610d35578485fd5b6020830194508093505050509250925092565b6001600160a01b0391909116815260200190565b6001600160a01b039384168152919092166020820152604081019190915260600190565b6001600160a01b03929092168252602082015260400190565b901515815260200190565b9182526001600160a01b0316602082015260400190565b60208082526029908201527f6d6573736167652076616c7565206d757374206265203020666f7220746f6b656040820152681b8819195c1bdcda5d60ba1b606082015260800190565b60208082526029908201527f7061727469636970616e7473206c656e6774682073686f756c6420657175616c6040820152682062616c616e63657360b81b606082015260800190565b6020808252601f908201527f696e73756666696369656e742045544820666f72207769746864726177616c00604082015260600190565b60208082526025908201527f747279696e6720746f2073657420616c726561647920736574746c6564206368604082015264185b9b995b60da1b606082015260800190565b6020808252601d908201527f7369676e617475726520766572696669636174696f6e206661696c6564000000604082015260600190565b60208082526013908201527218da185b9b995b081b9bdd081cd95d1d1b1959606a1b604082015260600190565b6020808252601390820152721d1c985b9cd9995c919c9bdb4819985a5b1959606a1b604082015260600190565b6020808252600f908201526e1d1c985b9cd9995c8819985a5b1959608a1b604082015260600190565b60208082526025908201527f63616e206f6e6c792062652063616c6c6564206279207468652061646a75646960408201526431b0ba37b960d91b606082015260800190565b81358152608081016020830135610fde8161101d565b6001600160a01b039081166020840152604084013590610ffd8261101d565b166040830152606092830135929091019190915290565b90815260200190565b6001600160a01b038116811461103257600080fd5b5056fe45434453413a20696e76616c6964207369676e6174757265202773272076616c756545434453413a20696e76616c6964207369676e6174757265202776272076616c7565a26469706673582212206e350da03b460efcd5aa35276f45eab280960765145bd0513f7d8c6b88d8ae3864736f6c63430007040033"

// DeployAssetholdererc20 deploys a new Ethereum contract, binding an instance of Assetholdererc20 to it.
func DeployAssetholdererc20(auth *bind.TransactOpts, backend bind.ContractBackend, _adjudicator common.Address, _token common.Address) (common.Address, *types.Transaction, *Assetholdererc20, error) {
	parsed, err := abi.JSON(strings.NewReader(Assetholdererc20ABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(Assetholdererc20Bin), backend, _adjudicator, _token)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Assetholdererc20{Assetholdererc20Caller: Assetholdererc20Caller{contract: contract}, Assetholdererc20Transactor: Assetholdererc20Transactor{contract: contract}, Assetholdererc20Filterer: Assetholdererc20Filterer{contract: contract}}, nil
}

// Assetholdererc20 is an auto generated Go binding around an Ethereum contract.
type Assetholdererc20 struct {
	Assetholdererc20Caller     // Read-only binding to the contract
	Assetholdererc20Transactor // Write-only binding to the contract
	Assetholdererc20Filterer   // Log filterer for contract events
}

// Assetholdererc20Caller is an auto generated read-only Go binding around an Ethereum contract.
type Assetholdererc20Caller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// Assetholdererc20Transactor is an auto generated write-only Go binding around an Ethereum contract.
type Assetholdererc20Transactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// Assetholdererc20Filterer is an auto generated log filtering Go binding around an Ethereum contract events.
type Assetholdererc20Filterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// Assetholdererc20Session is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type Assetholdererc20Session struct {
	Contract     *Assetholdererc20 // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// Assetholdererc20CallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type Assetholdererc20CallerSession struct {
	Contract *Assetholdererc20Caller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts           // Call options to use throughout this session
}

// Assetholdererc20TransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type Assetholdererc20TransactorSession struct {
	Contract     *Assetholdererc20Transactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts           // Transaction auth options to use throughout this session
}

// Assetholdererc20Raw is an auto generated low-level Go binding around an Ethereum contract.
type Assetholdererc20Raw struct {
	Contract *Assetholdererc20 // Generic contract binding to access the raw methods on
}

// Assetholdererc20CallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type Assetholdererc20CallerRaw struct {
	Contract *Assetholdererc20Caller // Generic read-only contract binding to access the raw methods on
}

// Assetholdererc20TransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type Assetholdererc20TransactorRaw struct {
	Contract *Assetholdererc20Transactor // Generic write-only contract binding to access the raw methods on
}

// NewAssetholdererc20 creates a new instance of Assetholdererc20, bound to a specific deployed contract.
func NewAssetholdererc20(address common.Address, backend bind.ContractBackend) (*Assetholdererc20, error) {
	contract, err := bindAssetholdererc20(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Assetholdererc20{Assetholdererc20Caller: Assetholdererc20Caller{contract: contract}, Assetholdererc20Transactor: Assetholdererc20Transactor{contract: contract}, Assetholdererc20Filterer: Assetholdererc20Filterer{contract: contract}}, nil
}

// NewAssetholdererc20Caller creates a new read-only instance of Assetholdererc20, bound to a specific deployed contract.
func NewAssetholdererc20Caller(address common.Address, caller bind.ContractCaller) (*Assetholdererc20Caller, error) {
	contract, err := bindAssetholdererc20(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &Assetholdererc20Caller{contract: contract}, nil
}

// NewAssetholdererc20Transactor creates a new write-only instance of Assetholdererc20, bound to a specific deployed contract.
func NewAssetholdererc20Transactor(address common.Address, transactor bind.ContractTransactor) (*Assetholdererc20Transactor, error) {
	contract, err := bindAssetholdererc20(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &Assetholdererc20Transactor{contract: contract}, nil
}

// NewAssetholdererc20Filterer creates a new log filterer instance of Assetholdererc20, bound to a specific deployed contract.
func NewAssetholdererc20Filterer(address common.Address, filterer bind.ContractFilterer) (*Assetholdererc20Filterer, error) {
	contract, err := bindAssetholdererc20(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &Assetholdererc20Filterer{contract: contract}, nil
}

// bindAssetholdererc20 binds a generic wrapper to an already deployed contract.
func bindAssetholdererc20(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(Assetholdererc20ABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Assetholdererc20 *Assetholdererc20Raw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Assetholdererc20.Contract.Assetholdererc20Caller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Assetholdererc20 *Assetholdererc20Raw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Assetholdererc20.Contract.Assetholdererc20Transactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Assetholdererc20 *Assetholdererc20Raw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Assetholdererc20.Contract.Assetholdererc20Transactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Assetholdererc20 *Assetholdererc20CallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Assetholdererc20.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Assetholdererc20 *Assetholdererc20TransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Assetholdererc20.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Assetholdererc20 *Assetholdererc20TransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Assetholdererc20.Contract.contract.Transact(opts, method, params...)
}

// Adjudicator is a free data retrieval call binding the contract method 0x53c2ed8e.
//
// Solidity: function adjudicator() view returns(address)
func (_Assetholdererc20 *Assetholdererc20Caller) Adjudicator(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Assetholdererc20.contract.Call(opts, &out, "adjudicator")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Adjudicator is a free data retrieval call binding the contract method 0x53c2ed8e.
//
// Solidity: function adjudicator() view returns(address)
func (_Assetholdererc20 *Assetholdererc20Session) Adjudicator() (common.Address, error) {
	return _Assetholdererc20.Contract.Adjudicator(&_Assetholdererc20.CallOpts)
}

// Adjudicator is a free data retrieval call binding the contract method 0x53c2ed8e.
//
// Solidity: function adjudicator() view returns(address)
func (_Assetholdererc20 *Assetholdererc20CallerSession) Adjudicator() (common.Address, error) {
	return _Assetholdererc20.Contract.Adjudicator(&_Assetholdererc20.CallOpts)
}

// Holdings is a free data retrieval call binding the contract method 0xae9ee18c.
//
// Solidity: function holdings(bytes32 ) view returns(uint256)
func (_Assetholdererc20 *Assetholdererc20Caller) Holdings(opts *bind.CallOpts, arg0 [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _Assetholdererc20.contract.Call(opts, &out, "holdings", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Holdings is a free data retrieval call binding the contract method 0xae9ee18c.
//
// Solidity: function holdings(bytes32 ) view returns(uint256)
func (_Assetholdererc20 *Assetholdererc20Session) Holdings(arg0 [32]byte) (*big.Int, error) {
	return _Assetholdererc20.Contract.Holdings(&_Assetholdererc20.CallOpts, arg0)
}

// Holdings is a free data retrieval call binding the contract method 0xae9ee18c.
//
// Solidity: function holdings(bytes32 ) view returns(uint256)
func (_Assetholdererc20 *Assetholdererc20CallerSession) Holdings(arg0 [32]byte) (*big.Int, error) {
	return _Assetholdererc20.Contract.Holdings(&_Assetholdererc20.CallOpts, arg0)
}

// Settled is a free data retrieval call binding the contract method 0xd945af1d.
//
// Solidity: function settled(bytes32 ) view returns(bool)
func (_Assetholdererc20 *Assetholdererc20Caller) Settled(opts *bind.CallOpts, arg0 [32]byte) (bool, error) {
	var out []interface{}
	err := _Assetholdererc20.contract.Call(opts, &out, "settled", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Settled is a free data retrieval call binding the contract method 0xd945af1d.
//
// Solidity: function settled(bytes32 ) view returns(bool)
func (_Assetholdererc20 *Assetholdererc20Session) Settled(arg0 [32]byte) (bool, error) {
	return _Assetholdererc20.Contract.Settled(&_Assetholdererc20.CallOpts, arg0)
}

// Settled is a free data retrieval call binding the contract method 0xd945af1d.
//
// Solidity: function settled(bytes32 ) view returns(bool)
func (_Assetholdererc20 *Assetholdererc20CallerSession) Settled(arg0 [32]byte) (bool, error) {
	return _Assetholdererc20.Contract.Settled(&_Assetholdererc20.CallOpts, arg0)
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_Assetholdererc20 *Assetholdererc20Caller) Token(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Assetholdererc20.contract.Call(opts, &out, "token")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_Assetholdererc20 *Assetholdererc20Session) Token() (common.Address, error) {
	return _Assetholdererc20.Contract.Token(&_Assetholdererc20.CallOpts)
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_Assetholdererc20 *Assetholdererc20CallerSession) Token() (common.Address, error) {
	return _Assetholdererc20.Contract.Token(&_Assetholdererc20.CallOpts)
}

// Deposit is a paid mutator transaction binding the contract method 0x1de26e16.
//
// Solidity: function deposit(bytes32 fundingID, uint256 amount) payable returns()
func (_Assetholdererc20 *Assetholdererc20Transactor) Deposit(opts *bind.TransactOpts, fundingID [32]byte, amount *big.Int) (*types.Transaction, error) {
	return _Assetholdererc20.contract.Transact(opts, "deposit", fundingID, amount)
}

// Deposit is a paid mutator transaction binding the contract method 0x1de26e16.
//
// Solidity: function deposit(bytes32 fundingID, uint256 amount) payable returns()
func (_Assetholdererc20 *Assetholdererc20Session) Deposit(fundingID [32]byte, amount *big.Int) (*types.Transaction, error) {
	return _Assetholdererc20.Contract.Deposit(&_Assetholdererc20.TransactOpts, fundingID, amount)
}

// Deposit is a paid mutator transaction binding the contract method 0x1de26e16.
//
// Solidity: function deposit(bytes32 fundingID, uint256 amount) payable returns()
func (_Assetholdererc20 *Assetholdererc20TransactorSession) Deposit(fundingID [32]byte, amount *big.Int) (*types.Transaction, error) {
	return _Assetholdererc20.Contract.Deposit(&_Assetholdererc20.TransactOpts, fundingID, amount)
}

// SetOutcome is a paid mutator transaction binding the contract method 0xfc79a66d.
//
// Solidity: function setOutcome(bytes32 channelID, address[] parts, uint256[] newBals) returns()
func (_Assetholdererc20 *Assetholdererc20Transactor) SetOutcome(opts *bind.TransactOpts, channelID [32]byte, parts []common.Address, newBals []*big.Int) (*types.Transaction, error) {
	return _Assetholdererc20.contract.Transact(opts, "setOutcome", channelID, parts, newBals)
}

// SetOutcome is a paid mutator transaction binding the contract method 0xfc79a66d.
//
// Solidity: function setOutcome(bytes32 channelID, address[] parts, uint256[] newBals) returns()
func (_Assetholdererc20 *Assetholdererc20Session) SetOutcome(channelID [32]byte, parts []common.Address, newBals []*big.Int) (*types.Transaction, error) {
	return _Assetholdererc20.Contract.SetOutcome(&_Assetholdererc20.TransactOpts, channelID, parts, newBals)
}

// SetOutcome is a paid mutator transaction binding the contract method 0xfc79a66d.
//
// Solidity: function setOutcome(bytes32 channelID, address[] parts, uint256[] newBals) returns()
func (_Assetholdererc20 *Assetholdererc20TransactorSession) SetOutcome(channelID [32]byte, parts []common.Address, newBals []*big.Int) (*types.Transaction, error) {
	return _Assetholdererc20.Contract.SetOutcome(&_Assetholdererc20.TransactOpts, channelID, parts, newBals)
}

// Withdraw is a paid mutator transaction binding the contract method 0x4ed4283c.
//
// Solidity: function withdraw((bytes32,address,address,uint256) authorization, bytes signature) returns()
func (_Assetholdererc20 *Assetholdererc20Transactor) Withdraw(opts *bind.TransactOpts, authorization AssetHolderWithdrawalAuth, signature []byte) (*types.Transaction, error) {
	return _Assetholdererc20.contract.Transact(opts, "withdraw", authorization, signature)
}

// Withdraw is a paid mutator transaction binding the contract method 0x4ed4283c.
//
// Solidity: function withdraw((bytes32,address,address,uint256) authorization, bytes signature) returns()
func (_Assetholdererc20 *Assetholdererc20Session) Withdraw(authorization AssetHolderWithdrawalAuth, signature []byte) (*types.Transaction, error) {
	return _Assetholdererc20.Contract.Withdraw(&_Assetholdererc20.TransactOpts, authorization, signature)
}

// Withdraw is a paid mutator transaction binding the contract method 0x4ed4283c.
//
// Solidity: function withdraw((bytes32,address,address,uint256) authorization, bytes signature) returns()
func (_Assetholdererc20 *Assetholdererc20TransactorSession) Withdraw(authorization AssetHolderWithdrawalAuth, signature []byte) (*types.Transaction, error) {
	return _Assetholdererc20.Contract.Withdraw(&_Assetholdererc20.TransactOpts, authorization, signature)
}

// Assetholdererc20DepositedIterator is returned from FilterDeposited and is used to iterate over the raw logs and unpacked data for Deposited events raised by the Assetholdererc20 contract.
type Assetholdererc20DepositedIterator struct {
	Event *Assetholdererc20Deposited // Event containing the contract specifics and raw log

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
func (it *Assetholdererc20DepositedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Assetholdererc20Deposited)
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
		it.Event = new(Assetholdererc20Deposited)
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
func (it *Assetholdererc20DepositedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Assetholdererc20DepositedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Assetholdererc20Deposited represents a Deposited event raised by the Assetholdererc20 contract.
type Assetholdererc20Deposited struct {
	FundingID [32]byte
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterDeposited is a free log retrieval operation binding the contract event 0xcd2fe07293de5928c5df9505b65a8d6506f8668dfe81af09090920687edc48a9.
//
// Solidity: event Deposited(bytes32 indexed fundingID, uint256 amount)
func (_Assetholdererc20 *Assetholdererc20Filterer) FilterDeposited(opts *bind.FilterOpts, fundingID [][32]byte) (*Assetholdererc20DepositedIterator, error) {

	var fundingIDRule []interface{}
	for _, fundingIDItem := range fundingID {
		fundingIDRule = append(fundingIDRule, fundingIDItem)
	}

	logs, sub, err := _Assetholdererc20.contract.FilterLogs(opts, "Deposited", fundingIDRule)
	if err != nil {
		return nil, err
	}
	return &Assetholdererc20DepositedIterator{contract: _Assetholdererc20.contract, event: "Deposited", logs: logs, sub: sub}, nil
}

// WatchDeposited is a free log subscription operation binding the contract event 0xcd2fe07293de5928c5df9505b65a8d6506f8668dfe81af09090920687edc48a9.
//
// Solidity: event Deposited(bytes32 indexed fundingID, uint256 amount)
func (_Assetholdererc20 *Assetholdererc20Filterer) WatchDeposited(opts *bind.WatchOpts, sink chan<- *Assetholdererc20Deposited, fundingID [][32]byte) (event.Subscription, error) {

	var fundingIDRule []interface{}
	for _, fundingIDItem := range fundingID {
		fundingIDRule = append(fundingIDRule, fundingIDItem)
	}

	logs, sub, err := _Assetholdererc20.contract.WatchLogs(opts, "Deposited", fundingIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Assetholdererc20Deposited)
				if err := _Assetholdererc20.contract.UnpackLog(event, "Deposited", log); err != nil {
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
func (_Assetholdererc20 *Assetholdererc20Filterer) ParseDeposited(log types.Log) (*Assetholdererc20Deposited, error) {
	event := new(Assetholdererc20Deposited)
	if err := _Assetholdererc20.contract.UnpackLog(event, "Deposited", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// Assetholdererc20OutcomeSetIterator is returned from FilterOutcomeSet and is used to iterate over the raw logs and unpacked data for OutcomeSet events raised by the Assetholdererc20 contract.
type Assetholdererc20OutcomeSetIterator struct {
	Event *Assetholdererc20OutcomeSet // Event containing the contract specifics and raw log

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
func (it *Assetholdererc20OutcomeSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Assetholdererc20OutcomeSet)
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
		it.Event = new(Assetholdererc20OutcomeSet)
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
func (it *Assetholdererc20OutcomeSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Assetholdererc20OutcomeSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Assetholdererc20OutcomeSet represents a OutcomeSet event raised by the Assetholdererc20 contract.
type Assetholdererc20OutcomeSet struct {
	ChannelID [32]byte
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterOutcomeSet is a free log retrieval operation binding the contract event 0xef898d6cd3395b6dfe67a3c1923e5c726c1b154e979fb0a25a9c41d0093168b8.
//
// Solidity: event OutcomeSet(bytes32 indexed channelID)
func (_Assetholdererc20 *Assetholdererc20Filterer) FilterOutcomeSet(opts *bind.FilterOpts, channelID [][32]byte) (*Assetholdererc20OutcomeSetIterator, error) {

	var channelIDRule []interface{}
	for _, channelIDItem := range channelID {
		channelIDRule = append(channelIDRule, channelIDItem)
	}

	logs, sub, err := _Assetholdererc20.contract.FilterLogs(opts, "OutcomeSet", channelIDRule)
	if err != nil {
		return nil, err
	}
	return &Assetholdererc20OutcomeSetIterator{contract: _Assetholdererc20.contract, event: "OutcomeSet", logs: logs, sub: sub}, nil
}

// WatchOutcomeSet is a free log subscription operation binding the contract event 0xef898d6cd3395b6dfe67a3c1923e5c726c1b154e979fb0a25a9c41d0093168b8.
//
// Solidity: event OutcomeSet(bytes32 indexed channelID)
func (_Assetholdererc20 *Assetholdererc20Filterer) WatchOutcomeSet(opts *bind.WatchOpts, sink chan<- *Assetholdererc20OutcomeSet, channelID [][32]byte) (event.Subscription, error) {

	var channelIDRule []interface{}
	for _, channelIDItem := range channelID {
		channelIDRule = append(channelIDRule, channelIDItem)
	}

	logs, sub, err := _Assetholdererc20.contract.WatchLogs(opts, "OutcomeSet", channelIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Assetholdererc20OutcomeSet)
				if err := _Assetholdererc20.contract.UnpackLog(event, "OutcomeSet", log); err != nil {
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
func (_Assetholdererc20 *Assetholdererc20Filterer) ParseOutcomeSet(log types.Log) (*Assetholdererc20OutcomeSet, error) {
	event := new(Assetholdererc20OutcomeSet)
	if err := _Assetholdererc20.contract.UnpackLog(event, "OutcomeSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// Assetholdererc20WithdrawnIterator is returned from FilterWithdrawn and is used to iterate over the raw logs and unpacked data for Withdrawn events raised by the Assetholdererc20 contract.
type Assetholdererc20WithdrawnIterator struct {
	Event *Assetholdererc20Withdrawn // Event containing the contract specifics and raw log

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
func (it *Assetholdererc20WithdrawnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Assetholdererc20Withdrawn)
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
		it.Event = new(Assetholdererc20Withdrawn)
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
func (it *Assetholdererc20WithdrawnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Assetholdererc20WithdrawnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Assetholdererc20Withdrawn represents a Withdrawn event raised by the Assetholdererc20 contract.
type Assetholdererc20Withdrawn struct {
	FundingID [32]byte
	Amount    *big.Int
	Receiver  common.Address
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterWithdrawn is a free log retrieval operation binding the contract event 0xd0b6e7d0170f56c62f87de6a8a47a0ccf41c86ffb5084d399d8eb62e823f2a81.
//
// Solidity: event Withdrawn(bytes32 indexed fundingID, uint256 amount, address receiver)
func (_Assetholdererc20 *Assetholdererc20Filterer) FilterWithdrawn(opts *bind.FilterOpts, fundingID [][32]byte) (*Assetholdererc20WithdrawnIterator, error) {

	var fundingIDRule []interface{}
	for _, fundingIDItem := range fundingID {
		fundingIDRule = append(fundingIDRule, fundingIDItem)
	}

	logs, sub, err := _Assetholdererc20.contract.FilterLogs(opts, "Withdrawn", fundingIDRule)
	if err != nil {
		return nil, err
	}
	return &Assetholdererc20WithdrawnIterator{contract: _Assetholdererc20.contract, event: "Withdrawn", logs: logs, sub: sub}, nil
}

// WatchWithdrawn is a free log subscription operation binding the contract event 0xd0b6e7d0170f56c62f87de6a8a47a0ccf41c86ffb5084d399d8eb62e823f2a81.
//
// Solidity: event Withdrawn(bytes32 indexed fundingID, uint256 amount, address receiver)
func (_Assetholdererc20 *Assetholdererc20Filterer) WatchWithdrawn(opts *bind.WatchOpts, sink chan<- *Assetholdererc20Withdrawn, fundingID [][32]byte) (event.Subscription, error) {

	var fundingIDRule []interface{}
	for _, fundingIDItem := range fundingID {
		fundingIDRule = append(fundingIDRule, fundingIDItem)
	}

	logs, sub, err := _Assetholdererc20.contract.WatchLogs(opts, "Withdrawn", fundingIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Assetholdererc20Withdrawn)
				if err := _Assetholdererc20.contract.UnpackLog(event, "Withdrawn", log); err != nil {
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
func (_Assetholdererc20 *Assetholdererc20Filterer) ParseWithdrawn(log types.Log) (*Assetholdererc20Withdrawn, error) {
	event := new(Assetholdererc20Withdrawn)
	if err := _Assetholdererc20.contract.UnpackLog(event, "Withdrawn", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
