// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package adjudicator

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
	_ = abi.U256
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// AdjudicatorABI is the input ABI used to generate the binding from.
const AdjudicatorABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"}],\"name\":\"Concluded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"}],\"name\":\"FinalStateRegistered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"}],\"name\":\"PushOutcome\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"}],\"name\":\"Refuted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"}],\"name\":\"Registered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"}],\"name\":\"Responded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timeout\",\"type\":\"uint256\"}],\"name\":\"Stored\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"challengeDuration\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"app\",\"type\":\"address\"},{\"internalType\":\"address[]\",\"name\":\"participants\",\"type\":\"address[]\"}],\"internalType\":\"structPerunTypes.Params\",\"name\":\"p\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"},{\"components\":[{\"internalType\":\"address[]\",\"name\":\"assets\",\"type\":\"address[]\"},{\"internalType\":\"uint256[][]\",\"name\":\"balances\",\"type\":\"uint256[][]\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"ID\",\"type\":\"bytes32\"},{\"internalType\":\"uint256[]\",\"name\":\"balances\",\"type\":\"uint256[]\"}],\"internalType\":\"structPerunTypes.SubAlloc[]\",\"name\":\"locked\",\"type\":\"tuple[]\"}],\"internalType\":\"structPerunTypes.Allocation\",\"name\":\"outcome\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"appData\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"isFinal\",\"type\":\"bool\"}],\"internalType\":\"structPerunTypes.State\",\"name\":\"s\",\"type\":\"tuple\"},{\"internalType\":\"uint256\",\"name\":\"timeout\",\"type\":\"uint256\"},{\"internalType\":\"enumAdjudicator.DisputeState\",\"name\":\"disputeState\",\"type\":\"uint8\"}],\"name\":\"concludeChallenge\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"disputeRegistry\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"challengeDuration\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"app\",\"type\":\"address\"},{\"internalType\":\"address[]\",\"name\":\"participants\",\"type\":\"address[]\"}],\"internalType\":\"structPerunTypes.Params\",\"name\":\"p\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"},{\"components\":[{\"internalType\":\"address[]\",\"name\":\"assets\",\"type\":\"address[]\"},{\"internalType\":\"uint256[][]\",\"name\":\"balances\",\"type\":\"uint256[][]\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"ID\",\"type\":\"bytes32\"},{\"internalType\":\"uint256[]\",\"name\":\"balances\",\"type\":\"uint256[]\"}],\"internalType\":\"structPerunTypes.SubAlloc[]\",\"name\":\"locked\",\"type\":\"tuple[]\"}],\"internalType\":\"structPerunTypes.Allocation\",\"name\":\"outcome\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"appData\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"isFinal\",\"type\":\"bool\"}],\"internalType\":\"structPerunTypes.State\",\"name\":\"old\",\"type\":\"tuple\"},{\"internalType\":\"uint256\",\"name\":\"timeout\",\"type\":\"uint256\"},{\"internalType\":\"enumAdjudicator.DisputeState\",\"name\":\"disputeState\",\"type\":\"uint8\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"},{\"components\":[{\"internalType\":\"address[]\",\"name\":\"assets\",\"type\":\"address[]\"},{\"internalType\":\"uint256[][]\",\"name\":\"balances\",\"type\":\"uint256[][]\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"ID\",\"type\":\"bytes32\"},{\"internalType\":\"uint256[]\",\"name\":\"balances\",\"type\":\"uint256[]\"}],\"internalType\":\"structPerunTypes.SubAlloc[]\",\"name\":\"locked\",\"type\":\"tuple[]\"}],\"internalType\":\"structPerunTypes.Allocation\",\"name\":\"outcome\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"appData\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"isFinal\",\"type\":\"bool\"}],\"internalType\":\"structPerunTypes.State\",\"name\":\"s\",\"type\":\"tuple\"},{\"internalType\":\"uint256\",\"name\":\"moverIdx\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"sig\",\"type\":\"bytes\"}],\"name\":\"progress\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"challengeDuration\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"app\",\"type\":\"address\"},{\"internalType\":\"address[]\",\"name\":\"participants\",\"type\":\"address[]\"}],\"internalType\":\"structPerunTypes.Params\",\"name\":\"p\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"},{\"components\":[{\"internalType\":\"address[]\",\"name\":\"assets\",\"type\":\"address[]\"},{\"internalType\":\"uint256[][]\",\"name\":\"balances\",\"type\":\"uint256[][]\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"ID\",\"type\":\"bytes32\"},{\"internalType\":\"uint256[]\",\"name\":\"balances\",\"type\":\"uint256[]\"}],\"internalType\":\"structPerunTypes.SubAlloc[]\",\"name\":\"locked\",\"type\":\"tuple[]\"}],\"internalType\":\"structPerunTypes.Allocation\",\"name\":\"outcome\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"appData\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"isFinal\",\"type\":\"bool\"}],\"internalType\":\"structPerunTypes.State\",\"name\":\"old\",\"type\":\"tuple\"},{\"internalType\":\"uint256\",\"name\":\"timeout\",\"type\":\"uint256\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"},{\"components\":[{\"internalType\":\"address[]\",\"name\":\"assets\",\"type\":\"address[]\"},{\"internalType\":\"uint256[][]\",\"name\":\"balances\",\"type\":\"uint256[][]\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"ID\",\"type\":\"bytes32\"},{\"internalType\":\"uint256[]\",\"name\":\"balances\",\"type\":\"uint256[]\"}],\"internalType\":\"structPerunTypes.SubAlloc[]\",\"name\":\"locked\",\"type\":\"tuple[]\"}],\"internalType\":\"structPerunTypes.Allocation\",\"name\":\"outcome\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"appData\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"isFinal\",\"type\":\"bool\"}],\"internalType\":\"structPerunTypes.State\",\"name\":\"s\",\"type\":\"tuple\"},{\"internalType\":\"bytes[]\",\"name\":\"sigs\",\"type\":\"bytes[]\"}],\"name\":\"refute\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"challengeDuration\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"app\",\"type\":\"address\"},{\"internalType\":\"address[]\",\"name\":\"participants\",\"type\":\"address[]\"}],\"internalType\":\"structPerunTypes.Params\",\"name\":\"p\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"},{\"components\":[{\"internalType\":\"address[]\",\"name\":\"assets\",\"type\":\"address[]\"},{\"internalType\":\"uint256[][]\",\"name\":\"balances\",\"type\":\"uint256[][]\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"ID\",\"type\":\"bytes32\"},{\"internalType\":\"uint256[]\",\"name\":\"balances\",\"type\":\"uint256[]\"}],\"internalType\":\"structPerunTypes.SubAlloc[]\",\"name\":\"locked\",\"type\":\"tuple[]\"}],\"internalType\":\"structPerunTypes.Allocation\",\"name\":\"outcome\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"appData\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"isFinal\",\"type\":\"bool\"}],\"internalType\":\"structPerunTypes.State\",\"name\":\"s\",\"type\":\"tuple\"},{\"internalType\":\"bytes[]\",\"name\":\"sigs\",\"type\":\"bytes[]\"}],\"name\":\"register\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"challengeDuration\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"app\",\"type\":\"address\"},{\"internalType\":\"address[]\",\"name\":\"participants\",\"type\":\"address[]\"}],\"internalType\":\"structPerunTypes.Params\",\"name\":\"p\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"},{\"components\":[{\"internalType\":\"address[]\",\"name\":\"assets\",\"type\":\"address[]\"},{\"internalType\":\"uint256[][]\",\"name\":\"balances\",\"type\":\"uint256[][]\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"ID\",\"type\":\"bytes32\"},{\"internalType\":\"uint256[]\",\"name\":\"balances\",\"type\":\"uint256[]\"}],\"internalType\":\"structPerunTypes.SubAlloc[]\",\"name\":\"locked\",\"type\":\"tuple[]\"}],\"internalType\":\"structPerunTypes.Allocation\",\"name\":\"outcome\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"appData\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"isFinal\",\"type\":\"bool\"}],\"internalType\":\"structPerunTypes.State\",\"name\":\"s\",\"type\":\"tuple\"},{\"internalType\":\"bytes[]\",\"name\":\"sigs\",\"type\":\"bytes[]\"}],\"name\":\"registerFinalState\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// AdjudicatorFuncSigs maps the 4-byte function signature to its string representation.
var AdjudicatorFuncSigs = map[string]string{
	"2a79c104": "concludeChallenge((uint256,uint256,address,address[]),(bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool),uint256,uint8)",
	"0a4f3e26": "disputeRegistry(bytes32)",
	"2f1bf781": "progress((uint256,uint256,address,address[]),(bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool),uint256,uint8,(bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool),uint256,bytes)",
	"91a628a9": "refute((uint256,uint256,address,address[]),(bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool),uint256,(bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool),bytes[])",
	"170e6715": "register((uint256,uint256,address,address[]),(bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool),bytes[])",
	"2f5dcf9a": "registerFinalState((uint256,uint256,address,address[]),(bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool),bytes[])",
}

// AdjudicatorBin is the compiled bytecode used for deploying new contracts.
var AdjudicatorBin = "0x608060405234801561001057600080fd5b5061260b806100206000396000f3fe608060405234801561001057600080fd5b50600436106100625760003560e01c80630a4f3e2614610067578063170e6715146100905780632a79c104146100a55780632f1bf781146100b85780632f5dcf9a146100cb57806391a628a9146100de575b600080fd5b61007a6100753660046114f2565b6100f1565b604051610087919061218e565b60405180910390f35b6100a361009e366004611510565b610103565b005b6100a36100b33660046115a1565b6101c0565b6100a36100c636600461162f565b610264565b6100a36100d9366004611510565b6103c5565b6100a36100ec366004611726565b61048c565b60006020819052908152604090205481565b600061010e846105a9565b8351909150811461013a5760405162461bcd60e51b815260040161013190612312565b60405180910390fd5b600081815260208190526040902054156101665760405162461bcd60e51b8152600401610131906123b2565b6101718484846105ef565b61017e8484836000610682565b807fc8704a622f3eb8c9fc5a2ddf1775b5ea7695359b13dec8111874e266a4d5fbc984602001516040516101b29190612478565b60405180910390a250505050565b81804210156101e15760405162461bcd60e51b815260040161013190612352565b60006101ec866105a9565b90506101fa868686866106fb565b600082815260208190526040902054146102265760405162461bcd60e51b815260040161013190612382565b61023181878761073f565b60405181907f3e659e9176c25a527f4575e010a270b3e1f8e9d1e94f5e49d4a91dd2c59e8cf890600090a2505050505050565b600084600181111561027257fe5b141561029857844210156102985760405162461bcd60e51b815260040161013190612352565b60006102a3886105a9565b845190915081146102c65760405162461bcd60e51b815260040161013190612372565b6102d2888888886106fb565b600082815260208190526040902054146102fe5760405162461bcd60e51b815260040161013190612382565b600061030a8584610a56565b9050886060015151841061031d57600080fd5b806001600160a01b03168960600151858151811061033757fe5b60200260200101516001600160a01b0316146103655760405162461bcd60e51b815260040161013190612392565b61037189898787610bcd565b61037e8986846001610682565b817ff4793471fad8d7bbe3211ef7eed6bbef53a8f2e0593826ec24a97931e249b42386602001516040516103b29190612478565b60405180910390a2505050505050505050565b608082015115156001146103eb5760405162461bcd60e51b8152600401610131906122b2565b60006103f6846105a9565b835190915081146104195760405162461bcd60e51b815260040161013190612312565b600081815260208190526040902054156104455760405162461bcd60e51b8152600401610131906123b2565b6104508484846105ef565b61045b81858561073f565b60405181907ffa302ab93de3c7a9581de1f9182591df6335562d06dc23ea6c8af24a0e3d5c1890600090a250505050565b828042106104ac5760405162461bcd60e51b815260040161013190612362565b84602001516001600160401b031683602001516001600160401b0316116104e55760405162461bcd60e51b8152600401610131906123a2565b60006104f0876105a9565b845190915081146105135760405162461bcd60e51b8152600401610131906122f2565b61052087878760006106fb565b6000828152602081905260409020541461054c5760405162461bcd60e51b815260040161013190612382565b6105578785856105ef565b6105648785836000610682565b807fd478cbccdd5ca6d246b145bb539b375b45c30ce42f63235b10ee19e4bc0f63c785602001516040516105989190612478565b60405180910390a250505050505050565b600081600001518260200151836040015184606001516040516020016105d2949392919061243e565b604051602081830303815290604052805190602001209050919050565b8051836060015151146105fe57fe5b60005b815181101561067c5760006106298484848151811061061c57fe5b6020026020010151610a56565b9050806001600160a01b03168560600151838151811061064557fe5b60200260200101516001600160a01b0316146106735760405162461bcd60e51b8152600401610131906122c2565b50600101610601565b50505050565b835160009061069890429063ffffffff610ccb16565b90506106a6858583856106fb565b60008085815260200190815260200160002081905550827fde02b1ac594e3d12f3797b91ed3e93213c5fcb9a6963fe4003c2fc8287e67c31826040516106ec919061218e565b60405180910390a25050505050565b600084848484600181111561070c57fe5b60405160200161071f949392919061240a565b604051602081830303815290604052805190602001209050949350505050565b60608160400151600001515160405190808252806020026020018201604052801561077e57816020015b60608152602001906001900390816107695790505b5090506060826040015160400151516040519080825280602002602001820160405280156107b6578160200160208202803883390190505b50905060005b8360400151604001515181101561090c5783604001516040015181815181106107e157fe5b6020026020010151600001518282815181106107f957fe5b602090810291909101015260005b604085015151518110156109035781610866578460400151604001515160405190808252806020026020018201604052801561084d578160200160208202803883390190505b5084828151811061085a57fe5b60200260200101819052505b6108d1856040015160400151838151811061087d57fe5b602002602001015160200151828151811061089457fe5b60200260200101518583815181106108a857fe5b602002602001015184815181106108bb57fe5b6020026020010151610ccb90919063ffffffff16565b8482815181106108dd57fe5b602002602001015183815181106108f057fe5b6020908102919091010152600101610807565b506001016107bc565b5060005b60408401515151811015610a23576000846040015160000151828151811061093457fe5b60200260200101519050856060015151856040015160200151838151811061095857fe5b6020026020010151511461097e5760405162461bcd60e51b815260040161013190612322565b806001600160a01b03166379aad62e88886060015188604001516020015186815181106109a757fe5b6020026020010151878988815181106109bc57fe5b60200260200101516040518663ffffffff1660e01b81526004016109e495949392919061219c565b600060405180830381600087803b1580156109fe57600080fd5b505af1158015610a12573d6000803e3d6000fd5b505060019093019250610910915050565b5060405185907f18a580a4aab39f3e138ed4cf306861cb9702f09856253189563ccaec335f0ffb90600090a25050505050565b600060606040518060400160405280601c81526020017f19457468657265756d205369676e6564204d6573736167653a0a33320000000081525090506060846040015160400151600081518110610aa957fe5b602002602001015160000151856040015160400151600081518110610aca57fe5b602002602001015160200151604051602001610ae7929190612203565b60408051601f19818403018152828252908701518051602091820151929450606093610b17939192869101612155565b60405160208183030381529060405290506060866000015187602001518389606001518a60800151604051602001610b53959493929190612223565b604051602081830303815290604052905060008180519060200120905060008582604051602001610b85929190612133565b6040516020818303038152906040528051906020012090506000610ba9828a610cf7565b90506001600160a01b038116610bbe57600080fd5b96505050505050505b92915050565b82602001516001016001600160401b031682602001516001600160401b031614610c095760405162461bcd60e51b8152600401610131906122d2565b610c2183604001518360400151866060015151610dd3565b6040808501519051637614eebf60e11b81526001600160a01b0382169063ec29dd7e90610c589088908890889088906004016123c2565b60206040518083038186803b158015610c7057600080fd5b505afa158015610c84573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250610ca891908101906114cc565b610cc45760405162461bcd60e51b815260040161013190612332565b5050505050565b600082820183811015610cf05760405162461bcd60e51b815260040161013190612302565b9392505050565b60008151604114610d0a57506000610bc7565b60208201516040830151606084015160001a7f7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a0821115610d505760009350505050610bc7565b8060ff16601b14158015610d6857508060ff16601c14155b15610d795760009350505050610bc7565b60018682858560405160008152602001604052604051610d9c949392919061227d565b6020604051602081039080840390855afa158015610dbe573d6000803e3d6000fd5b5050604051601f190151979650505050505050565b81602001515183602001515114610de657fe5b81515183515114610df357fe5b60005b82515181101561067c578251805182908110610e0e57fe5b60200260200101516001600160a01b031684600001518281518110610e2f57fe5b60200260200101516001600160a01b031614610e5d5760405162461bcd60e51b815260040161013190612342565b60208301518051600091829184908110610e7357fe5b60200260200101515186602001518481518110610e8c57fe5b60200260200101515114610e9c57fe5b8386602001518481518110610ead57fe5b60200260200101515114610ebd57fe5b60005b85602001518481518110610ed057fe5b602002602001015151811015610f6657610f1d87602001518581518110610ef357fe5b60200260200101518281518110610f0657fe5b602002602001015184610ccb90919063ffffffff16565b9250610f5c86602001518581518110610f3257fe5b60200260200101518281518110610f4557fe5b602002602001015183610ccb90919063ffffffff16565b9150600101610ec0565b5060005b866040015151811015610fd757610fa187604001518281518110610f8a57fe5b6020026020010151602001518581518110610f0657fe5b9250610fcd86604001518281518110610fb657fe5b6020026020010151602001518581518110610f4557fe5b9150600101610f6a565b50808214610ff75760405162461bcd60e51b8152600401610131906122e2565b5050600101610df6565b8035610bc781612589565b600082601f83011261101d57600080fd5b813561103061102b826124ac565b612486565b9150818183526020840193506020810190508385602084028201111561105557600080fd5b60005b83811015611081578161106b8882611001565b8452506020928301929190910190600101611058565b5050505092915050565b600082601f83011261109c57600080fd5b81356110aa61102b826124ac565b81815260209384019390925082018360005b8381101561108157813586016110d288826111a2565b84525060209283019291909101906001016110bc565b600082601f8301126110f957600080fd5b813561110761102b826124ac565b81815260209384019390925082018360005b83811015611081578135860161112f8882611233565b8452506020928301929190910190600101611119565b600082601f83011261115657600080fd5b813561116461102b826124ac565b81815260209384019390925082018360005b83811015611081578135860161118c8882611463565b8452506020928301929190910190600101611176565b600082601f8301126111b357600080fd5b81356111c161102b826124ac565b915081818352602084019350602081019050838560208402820111156111e657600080fd5b60005b8381101561108157816111fc8882611228565b84525060209283019291909101906001016111e9565b8035610bc7816125a0565b8051610bc7816125a0565b8035610bc7816125a9565b600082601f83011261124457600080fd5b813561125261102b826124cc565b9150808252602083016020830185838301111561126e57600080fd5b611279838284612547565b50505092915050565b8035610bc7816125b2565b60006060828403121561129f57600080fd5b6112a96060612486565b905081356001600160401b038111156112c157600080fd5b6112cd8482850161100c565b82525060208201356001600160401b038111156112e957600080fd5b6112f58482850161108b565b60208301525060408201356001600160401b0381111561131457600080fd5b61132084828501611145565b60408301525092915050565b60006080828403121561133e57600080fd5b6113486080612486565b905060006113568484611228565b825250602061136784848301611228565b602083015250604061137b84828501611001565b60408301525060608201356001600160401b0381111561139a57600080fd5b6113a68482850161100c565b60608301525092915050565b600060a082840312156113c457600080fd5b6113ce60a0612486565b905060006113dc8484611228565b82525060206113ed848483016114c1565b60208301525060408201356001600160401b0381111561140c57600080fd5b6114188482850161128d565b60408301525060608201356001600160401b0381111561143757600080fd5b61144384828501611233565b606083015250608061145784828501611212565b60808301525092915050565b60006040828403121561147557600080fd5b61147f6040612486565b9050600061148d8484611228565b82525060208201356001600160401b038111156114a957600080fd5b6114b5848285016111a2565b60208301525092915050565b8035610bc7816125bf565b6000602082840312156114de57600080fd5b60006114ea848461121d565b949350505050565b60006020828403121561150457600080fd5b60006114ea8484611228565b60008060006060848603121561152557600080fd5b83356001600160401b0381111561153b57600080fd5b6115478682870161132c565b93505060208401356001600160401b0381111561156357600080fd5b61156f868287016113b2565b92505060408401356001600160401b0381111561158b57600080fd5b611597868287016110e8565b9150509250925092565b600080600080608085870312156115b757600080fd5b84356001600160401b038111156115cd57600080fd5b6115d98782880161132c565b94505060208501356001600160401b038111156115f557600080fd5b611601878288016113b2565b935050604061161287828801611228565b925050606061162387828801611282565b91505092959194509250565b600080600080600080600060e0888a03121561164a57600080fd5b87356001600160401b0381111561166057600080fd5b61166c8a828b0161132c565b97505060208801356001600160401b0381111561168857600080fd5b6116948a828b016113b2565b96505060406116a58a828b01611228565b95505060606116b68a828b01611282565b94505060808801356001600160401b038111156116d257600080fd5b6116de8a828b016113b2565b93505060a06116ef8a828b01611228565b92505060c08801356001600160401b0381111561170b57600080fd5b6117178a828b01611233565b91505092959891949750929550565b600080600080600060a0868803121561173e57600080fd5b85356001600160401b0381111561175457600080fd5b6117608882890161132c565b95505060208601356001600160401b0381111561177c57600080fd5b611788888289016113b2565b945050604061179988828901611228565b93505060608601356001600160401b038111156117b557600080fd5b6117c1888289016113b2565b92505060808601356001600160401b038111156117dd57600080fd5b6117e9888289016110e8565b9150509295509295909350565b6000611802838361182e565b505060200190565b6000610cf08383611a62565b60006118028383611b07565b6000610cf083836120ec565b6118378161250b565b82525050565b6000611848826124f9565b61185281856124fd565b935061185d836124f3565b8060005b8381101561188b57815161187588826117f6565b9750611880836124f3565b925050600101611861565b509495945050505050565b60006118a1826124f9565b6118ab81856124fd565b93506118b6836124f3565b8060005b8381101561188b5781516118ce88826117f6565b97506118d9836124f3565b9250506001016118ba565b60006118ef826124f9565b6118f981856124fd565b93508360208202850161190b856124f3565b8060005b858110156119455784840389528151611928858261180a565b9450611933836124f3565b60209a909a019992505060010161190f565b5091979650505050505050565b600061195d826124f9565b61196781856124fd565b935083602082028501611979856124f3565b8060005b858110156119455784840389528151611996858261180a565b94506119a1836124f3565b60209a909a019992505060010161197d565b60006119be826124f9565b6119c881856124fd565b93506119d3836124f3565b8060005b8381101561188b5781516119eb8882611816565b97506119f6836124f3565b9250506001016119d7565b6000611a0c826124f9565b611a1681856124fd565b935083602082028501611a28856124f3565b8060005b858110156119455784840389528151611a458582611822565b9450611a50836124f3565b60209a909a0199925050600101611a2c565b6000611a6d826124f9565b611a7781856124fd565b9350611a82836124f3565b8060005b8381101561188b578151611a9a8882611816565b9750611aa5836124f3565b925050600101611a86565b6000611abb826124f9565b611ac581856124fd565b9350611ad0836124f3565b8060005b8381101561188b578151611ae88882611816565b9750611af3836124f3565b925050600101611ad4565b61183781612516565b6118378161251b565b611837611b1c8261251b565b61251b565b6000611b2c826124f9565b611b3681856124fd565b9350611b46818560208601612553565b611b4f8161257f565b9093019392505050565b6000611b64826124f9565b611b6e8185612506565b9350611b7e818560208601612553565b9290920192915050565b6000611b956018836124fd565b7f6f6e6c79206163636570742066696e616c207374617465730000000000000000815260200192915050565b6000611bce6011836124fd565b70696e76616c6964207369676e617475726560781b815260200192915050565b6000611bfb602b836124fd565b7f63616e206f6e6c7920616476616e6365207468652076657273696f6e20636f7581526a6e746572206279206f6e6560a81b602082015260400192915050565b6000611c48602a836124fd565b7f53756d206f662062616c616e63657320666f7220616e206173736574206d75738152691d08189948195c5d585b60b21b602082015260400192915050565b6000611c946027836124fd565b7f74726965642072656675746174696f6e207769746820696e76616c6964206368815266185b9b995b125160ca1b602082015260400192915050565b6000611cdd601b836124fd565b7f536166654d6174683a206164646974696f6e206f766572666c6f770000000000815260200192915050565b6000611d166023836124fd565b7f7472696564207265676973746572696e6720696e76616c6964206368616e6e658152621b125160ea1b602082015260400192915050565b6000611d5b6030836124fd565b7f62616c616e636573206c656e6774682073686f756c64206d617463682070617281526f0e8d2c6d2e0c2dce8e640d8cadccee8d60831b602082015260400192915050565b6000611dad6011836124fd565b70696e76616c6964206e657720737461746560781b815260200192915050565b6000611dda6018836124fd565b7f617373657420616464726573736573206d69736d617463680000000000000000815260200192915050565b6000611e13601e836124fd565b7f66756e6374696f6e2063616c6c6564206265666f72652074696d656f75740000815260200192915050565b6000611e4c601d836124fd565b7f66756e6374696f6e2063616c6c65642061667465722074696d656f7574000000815260200192915050565b6000611e856027836124fd565b7f747269656420746f20726573706f6e64207769746820696e76616c6964206368815266185b9b995b125160ca1b602082015260400192915050565b6000611ece6020836124fd565b7f70726f76696465642077726f6e67206f6c642073746174652f74696d656f7574815260200192915050565b6000611f07602b836124fd565b7f6d6f766572496478206973206e6f742073657420746f20746865206964206f6681526a103a34329039b2b73232b960a91b602082015260400192915050565b6000611f54602d836124fd565b7f6f6e6c7920612072656675746174696f6e20776974682061206e65776572207381526c1d185d19481a5cc81d985b1a59609a1b602082015260400192915050565b6000611fa36020836124fd565b7f6120646973707574652077617320616c72656164792072656769737465726564815260200192915050565b8051606080845260009190840190611fe7828261183d565b9150506020830151848203602086015261200182826118e4565b9150506040830151848203604086015261201b8282611a01565b95945050505050565b805160009060808401906120388582611b07565b50602083015161204b6020860182611b07565b50604083015161205e604086018261182e565b506060830151848203606086015261201b828261183d565b805160009060a084019061208a8582611b07565b50602083015161209d6020860182612121565b50604083015184820360408601526120b58282611fcf565b915050606083015184820360608601526120cf8282611b21565b91505060808301516120e46080860182611afe565b509392505050565b805160009060408401906121008582611b07565b506020830151848203602086015261201b8282611a62565b6118378161253c565b6118378161252a565b61183781612536565b600061213f8285611b59565b915061214b8284611b10565b5060200192915050565b606080825281016121668186611896565b9050818103602083015261217a8185611952565b9050818103604083015261201b8184611b21565b60208101610bc78284611b07565b60a081016121aa8288611b07565b81810360208301526121bc8187611896565b905081810360408301526121d08186611ab0565b905081810360608301526121e481856119b3565b905081810360808301526121f88184611ab0565b979650505050505050565b604081016122118285611b07565b81810360208301526114ea8184611ab0565b60a081016122318288611b07565b61223e6020830187612121565b81810360408301526122508186611b21565b905081810360608301526122648185611b21565b90506122736080830184611afe565b9695505050505050565b6080810161228b8287611b07565b612298602083018661212a565b6122a56040830185611b07565b61201b6060830184611b07565b60208082528101610bc781611b88565b60208082528101610bc781611bc1565b60208082528101610bc781611bee565b60208082528101610bc781611c3b565b60208082528101610bc781611c87565b60208082528101610bc781611cd0565b60208082528101610bc781611d09565b60208082528101610bc781611d4e565b60208082528101610bc781611da0565b60208082528101610bc781611dcd565b60208082528101610bc781611e06565b60208082528101610bc781611e3f565b60208082528101610bc781611e78565b60208082528101610bc781611ec1565b60208082528101610bc781611efa565b60208082528101610bc781611f47565b60208082528101610bc781611f96565b608080825281016123d38187612024565b905081810360208301526123e78186612076565b905081810360408301526123fb8185612076565b905061201b6060830184611b07565b6080808252810161241b8187612024565b9050818103602083015261242f8186612076565b90506122a56040830185611b07565b6080810161244c8287611b07565b6124596020830186611b07565b612466604083018561182e565b81810360608301526122738184611896565b60208101610bc78284612118565b6040518181016001600160401b03811182821017156124a457600080fd5b604052919050565b60006001600160401b038211156124c257600080fd5b5060209081020190565b60006001600160401b038211156124e257600080fd5b506020601f91909101601f19160190565b60200190565b5190565b90815260200190565b919050565b6000610bc78261251e565b151590565b90565b6001600160a01b031690565b6001600160401b031690565b60ff1690565b6000610bc78261252a565b82818337506000910152565b60005b8381101561256e578181015183820152602001612556565b8381111561067c5750506000910152565b601f01601f191690565b6125928161250b565b811461259d57600080fd5b50565b61259281612516565b6125928161251b565b6002811061259d57600080fd5b6125928161252a56fea365627a7a72315820ae0eaf34e9894b81cd150851d7dda43cb658df6f15dd62d3c72b1c23d512f4c26c6578706572696d656e74616cf564736f6c634300050c0040"

// DeployAdjudicator deploys a new Ethereum contract, binding an instance of Adjudicator to it.
func DeployAdjudicator(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Adjudicator, error) {
	parsed, err := abi.JSON(strings.NewReader(AdjudicatorABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(AdjudicatorBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Adjudicator{AdjudicatorCaller: AdjudicatorCaller{contract: contract}, AdjudicatorTransactor: AdjudicatorTransactor{contract: contract}, AdjudicatorFilterer: AdjudicatorFilterer{contract: contract}}, nil
}

// Adjudicator is an auto generated Go binding around an Ethereum contract.
type Adjudicator struct {
	AdjudicatorCaller     // Read-only binding to the contract
	AdjudicatorTransactor // Write-only binding to the contract
	AdjudicatorFilterer   // Log filterer for contract events
}

// AdjudicatorCaller is an auto generated read-only Go binding around an Ethereum contract.
type AdjudicatorCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AdjudicatorTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AdjudicatorTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AdjudicatorFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AdjudicatorFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AdjudicatorSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AdjudicatorSession struct {
	Contract     *Adjudicator      // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// AdjudicatorCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AdjudicatorCallerSession struct {
	Contract *AdjudicatorCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts      // Call options to use throughout this session
}

// AdjudicatorTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AdjudicatorTransactorSession struct {
	Contract     *AdjudicatorTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// AdjudicatorRaw is an auto generated low-level Go binding around an Ethereum contract.
type AdjudicatorRaw struct {
	Contract *Adjudicator // Generic contract binding to access the raw methods on
}

// AdjudicatorCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AdjudicatorCallerRaw struct {
	Contract *AdjudicatorCaller // Generic read-only contract binding to access the raw methods on
}

// AdjudicatorTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AdjudicatorTransactorRaw struct {
	Contract *AdjudicatorTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAdjudicator creates a new instance of Adjudicator, bound to a specific deployed contract.
func NewAdjudicator(address common.Address, backend bind.ContractBackend) (*Adjudicator, error) {
	contract, err := bindAdjudicator(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Adjudicator{AdjudicatorCaller: AdjudicatorCaller{contract: contract}, AdjudicatorTransactor: AdjudicatorTransactor{contract: contract}, AdjudicatorFilterer: AdjudicatorFilterer{contract: contract}}, nil
}

// NewAdjudicatorCaller creates a new read-only instance of Adjudicator, bound to a specific deployed contract.
func NewAdjudicatorCaller(address common.Address, caller bind.ContractCaller) (*AdjudicatorCaller, error) {
	contract, err := bindAdjudicator(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AdjudicatorCaller{contract: contract}, nil
}

// NewAdjudicatorTransactor creates a new write-only instance of Adjudicator, bound to a specific deployed contract.
func NewAdjudicatorTransactor(address common.Address, transactor bind.ContractTransactor) (*AdjudicatorTransactor, error) {
	contract, err := bindAdjudicator(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AdjudicatorTransactor{contract: contract}, nil
}

// NewAdjudicatorFilterer creates a new log filterer instance of Adjudicator, bound to a specific deployed contract.
func NewAdjudicatorFilterer(address common.Address, filterer bind.ContractFilterer) (*AdjudicatorFilterer, error) {
	contract, err := bindAdjudicator(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AdjudicatorFilterer{contract: contract}, nil
}

// bindAdjudicator binds a generic wrapper to an already deployed contract.
func bindAdjudicator(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(AdjudicatorABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Adjudicator *AdjudicatorRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Adjudicator.Contract.AdjudicatorCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Adjudicator *AdjudicatorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Adjudicator.Contract.AdjudicatorTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Adjudicator *AdjudicatorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Adjudicator.Contract.AdjudicatorTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Adjudicator *AdjudicatorCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Adjudicator.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Adjudicator *AdjudicatorTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Adjudicator.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Adjudicator *AdjudicatorTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Adjudicator.Contract.contract.Transact(opts, method, params...)
}

// DisputeRegistry is a free data retrieval call binding the contract method 0x0a4f3e26.
//
// Solidity: function disputeRegistry(bytes32 ) constant returns(bytes32)
func (_Adjudicator *AdjudicatorCaller) DisputeRegistry(opts *bind.CallOpts, arg0 [32]byte) ([32]byte, error) {
	var (
		ret0 = new([32]byte)
	)
	out := ret0
	err := _Adjudicator.contract.Call(opts, out, "disputeRegistry", arg0)
	return *ret0, err
}

// DisputeRegistry is a free data retrieval call binding the contract method 0x0a4f3e26.
//
// Solidity: function disputeRegistry(bytes32 ) constant returns(bytes32)
func (_Adjudicator *AdjudicatorSession) DisputeRegistry(arg0 [32]byte) ([32]byte, error) {
	return _Adjudicator.Contract.DisputeRegistry(&_Adjudicator.CallOpts, arg0)
}

// DisputeRegistry is a free data retrieval call binding the contract method 0x0a4f3e26.
//
// Solidity: function disputeRegistry(bytes32 ) constant returns(bytes32)
func (_Adjudicator *AdjudicatorCallerSession) DisputeRegistry(arg0 [32]byte) ([32]byte, error) {
	return _Adjudicator.Contract.DisputeRegistry(&_Adjudicator.CallOpts, arg0)
}

// ConcludeChallenge is a paid mutator transaction binding the contract method 0x2a79c104.
//
// Solidity: function concludeChallenge(PerunTypesParams p, PerunTypesState s, uint256 timeout, uint8 disputeState) returns()
func (_Adjudicator *AdjudicatorTransactor) ConcludeChallenge(opts *bind.TransactOpts, p PerunTypesParams, s PerunTypesState, timeout *big.Int, disputeState uint8) (*types.Transaction, error) {
	return _Adjudicator.contract.Transact(opts, "concludeChallenge", p, s, timeout, disputeState)
}

// ConcludeChallenge is a paid mutator transaction binding the contract method 0x2a79c104.
//
// Solidity: function concludeChallenge(PerunTypesParams p, PerunTypesState s, uint256 timeout, uint8 disputeState) returns()
func (_Adjudicator *AdjudicatorSession) ConcludeChallenge(p PerunTypesParams, s PerunTypesState, timeout *big.Int, disputeState uint8) (*types.Transaction, error) {
	return _Adjudicator.Contract.ConcludeChallenge(&_Adjudicator.TransactOpts, p, s, timeout, disputeState)
}

// ConcludeChallenge is a paid mutator transaction binding the contract method 0x2a79c104.
//
// Solidity: function concludeChallenge(PerunTypesParams p, PerunTypesState s, uint256 timeout, uint8 disputeState) returns()
func (_Adjudicator *AdjudicatorTransactorSession) ConcludeChallenge(p PerunTypesParams, s PerunTypesState, timeout *big.Int, disputeState uint8) (*types.Transaction, error) {
	return _Adjudicator.Contract.ConcludeChallenge(&_Adjudicator.TransactOpts, p, s, timeout, disputeState)
}

// Progress is a paid mutator transaction binding the contract method 0x2f1bf781.
//
// Solidity: function progress(PerunTypesParams p, PerunTypesState old, uint256 timeout, uint8 disputeState, PerunTypesState s, uint256 moverIdx, bytes sig) returns()
func (_Adjudicator *AdjudicatorTransactor) Progress(opts *bind.TransactOpts, p PerunTypesParams, old PerunTypesState, timeout *big.Int, disputeState uint8, s PerunTypesState, moverIdx *big.Int, sig []byte) (*types.Transaction, error) {
	return _Adjudicator.contract.Transact(opts, "progress", p, old, timeout, disputeState, s, moverIdx, sig)
}

// Progress is a paid mutator transaction binding the contract method 0x2f1bf781.
//
// Solidity: function progress(PerunTypesParams p, PerunTypesState old, uint256 timeout, uint8 disputeState, PerunTypesState s, uint256 moverIdx, bytes sig) returns()
func (_Adjudicator *AdjudicatorSession) Progress(p PerunTypesParams, old PerunTypesState, timeout *big.Int, disputeState uint8, s PerunTypesState, moverIdx *big.Int, sig []byte) (*types.Transaction, error) {
	return _Adjudicator.Contract.Progress(&_Adjudicator.TransactOpts, p, old, timeout, disputeState, s, moverIdx, sig)
}

// Progress is a paid mutator transaction binding the contract method 0x2f1bf781.
//
// Solidity: function progress(PerunTypesParams p, PerunTypesState old, uint256 timeout, uint8 disputeState, PerunTypesState s, uint256 moverIdx, bytes sig) returns()
func (_Adjudicator *AdjudicatorTransactorSession) Progress(p PerunTypesParams, old PerunTypesState, timeout *big.Int, disputeState uint8, s PerunTypesState, moverIdx *big.Int, sig []byte) (*types.Transaction, error) {
	return _Adjudicator.Contract.Progress(&_Adjudicator.TransactOpts, p, old, timeout, disputeState, s, moverIdx, sig)
}

// Refute is a paid mutator transaction binding the contract method 0x91a628a9.
//
// Solidity: function refute(PerunTypesParams p, PerunTypesState old, uint256 timeout, PerunTypesState s, bytes[] sigs) returns()
func (_Adjudicator *AdjudicatorTransactor) Refute(opts *bind.TransactOpts, p PerunTypesParams, old PerunTypesState, timeout *big.Int, s PerunTypesState, sigs [][]byte) (*types.Transaction, error) {
	return _Adjudicator.contract.Transact(opts, "refute", p, old, timeout, s, sigs)
}

// Refute is a paid mutator transaction binding the contract method 0x91a628a9.
//
// Solidity: function refute(PerunTypesParams p, PerunTypesState old, uint256 timeout, PerunTypesState s, bytes[] sigs) returns()
func (_Adjudicator *AdjudicatorSession) Refute(p PerunTypesParams, old PerunTypesState, timeout *big.Int, s PerunTypesState, sigs [][]byte) (*types.Transaction, error) {
	return _Adjudicator.Contract.Refute(&_Adjudicator.TransactOpts, p, old, timeout, s, sigs)
}

// Refute is a paid mutator transaction binding the contract method 0x91a628a9.
//
// Solidity: function refute(PerunTypesParams p, PerunTypesState old, uint256 timeout, PerunTypesState s, bytes[] sigs) returns()
func (_Adjudicator *AdjudicatorTransactorSession) Refute(p PerunTypesParams, old PerunTypesState, timeout *big.Int, s PerunTypesState, sigs [][]byte) (*types.Transaction, error) {
	return _Adjudicator.Contract.Refute(&_Adjudicator.TransactOpts, p, old, timeout, s, sigs)
}

// Register is a paid mutator transaction binding the contract method 0x170e6715.
//
// Solidity: function register(PerunTypesParams p, PerunTypesState s, bytes[] sigs) returns()
func (_Adjudicator *AdjudicatorTransactor) Register(opts *bind.TransactOpts, p PerunTypesParams, s PerunTypesState, sigs [][]byte) (*types.Transaction, error) {
	return _Adjudicator.contract.Transact(opts, "register", p, s, sigs)
}

// Register is a paid mutator transaction binding the contract method 0x170e6715.
//
// Solidity: function register(PerunTypesParams p, PerunTypesState s, bytes[] sigs) returns()
func (_Adjudicator *AdjudicatorSession) Register(p PerunTypesParams, s PerunTypesState, sigs [][]byte) (*types.Transaction, error) {
	return _Adjudicator.Contract.Register(&_Adjudicator.TransactOpts, p, s, sigs)
}

// Register is a paid mutator transaction binding the contract method 0x170e6715.
//
// Solidity: function register(PerunTypesParams p, PerunTypesState s, bytes[] sigs) returns()
func (_Adjudicator *AdjudicatorTransactorSession) Register(p PerunTypesParams, s PerunTypesState, sigs [][]byte) (*types.Transaction, error) {
	return _Adjudicator.Contract.Register(&_Adjudicator.TransactOpts, p, s, sigs)
}

// RegisterFinalState is a paid mutator transaction binding the contract method 0x2f5dcf9a.
//
// Solidity: function registerFinalState(PerunTypesParams p, PerunTypesState s, bytes[] sigs) returns()
func (_Adjudicator *AdjudicatorTransactor) RegisterFinalState(opts *bind.TransactOpts, p PerunTypesParams, s PerunTypesState, sigs [][]byte) (*types.Transaction, error) {
	return _Adjudicator.contract.Transact(opts, "registerFinalState", p, s, sigs)
}

// RegisterFinalState is a paid mutator transaction binding the contract method 0x2f5dcf9a.
//
// Solidity: function registerFinalState(PerunTypesParams p, PerunTypesState s, bytes[] sigs) returns()
func (_Adjudicator *AdjudicatorSession) RegisterFinalState(p PerunTypesParams, s PerunTypesState, sigs [][]byte) (*types.Transaction, error) {
	return _Adjudicator.Contract.RegisterFinalState(&_Adjudicator.TransactOpts, p, s, sigs)
}

// RegisterFinalState is a paid mutator transaction binding the contract method 0x2f5dcf9a.
//
// Solidity: function registerFinalState(PerunTypesParams p, PerunTypesState s, bytes[] sigs) returns()
func (_Adjudicator *AdjudicatorTransactorSession) RegisterFinalState(p PerunTypesParams, s PerunTypesState, sigs [][]byte) (*types.Transaction, error) {
	return _Adjudicator.Contract.RegisterFinalState(&_Adjudicator.TransactOpts, p, s, sigs)
}

// AdjudicatorConcludedIterator is returned from FilterConcluded and is used to iterate over the raw logs and unpacked data for Concluded events raised by the Adjudicator contract.
type AdjudicatorConcludedIterator struct {
	Event *AdjudicatorConcluded // Event containing the contract specifics and raw log

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
func (it *AdjudicatorConcludedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AdjudicatorConcluded)
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
		it.Event = new(AdjudicatorConcluded)
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
func (it *AdjudicatorConcludedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AdjudicatorConcludedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AdjudicatorConcluded represents a Concluded event raised by the Adjudicator contract.
type AdjudicatorConcluded struct {
	ChannelID [32]byte
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterConcluded is a free log retrieval operation binding the contract event 0x3e659e9176c25a527f4575e010a270b3e1f8e9d1e94f5e49d4a91dd2c59e8cf8.
//
// Solidity: event Concluded(bytes32 indexed channelID)
func (_Adjudicator *AdjudicatorFilterer) FilterConcluded(opts *bind.FilterOpts, channelID [][32]byte) (*AdjudicatorConcludedIterator, error) {

	var channelIDRule []interface{}
	for _, channelIDItem := range channelID {
		channelIDRule = append(channelIDRule, channelIDItem)
	}

	logs, sub, err := _Adjudicator.contract.FilterLogs(opts, "Concluded", channelIDRule)
	if err != nil {
		return nil, err
	}
	return &AdjudicatorConcludedIterator{contract: _Adjudicator.contract, event: "Concluded", logs: logs, sub: sub}, nil
}

// WatchConcluded is a free log subscription operation binding the contract event 0x3e659e9176c25a527f4575e010a270b3e1f8e9d1e94f5e49d4a91dd2c59e8cf8.
//
// Solidity: event Concluded(bytes32 indexed channelID)
func (_Adjudicator *AdjudicatorFilterer) WatchConcluded(opts *bind.WatchOpts, sink chan<- *AdjudicatorConcluded, channelID [][32]byte) (event.Subscription, error) {

	var channelIDRule []interface{}
	for _, channelIDItem := range channelID {
		channelIDRule = append(channelIDRule, channelIDItem)
	}

	logs, sub, err := _Adjudicator.contract.WatchLogs(opts, "Concluded", channelIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AdjudicatorConcluded)
				if err := _Adjudicator.contract.UnpackLog(event, "Concluded", log); err != nil {
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

// ParseConcluded is a log parse operation binding the contract event 0x3e659e9176c25a527f4575e010a270b3e1f8e9d1e94f5e49d4a91dd2c59e8cf8.
//
// Solidity: event Concluded(bytes32 indexed channelID)
func (_Adjudicator *AdjudicatorFilterer) ParseConcluded(log types.Log) (*AdjudicatorConcluded, error) {
	event := new(AdjudicatorConcluded)
	if err := _Adjudicator.contract.UnpackLog(event, "Concluded", log); err != nil {
		return nil, err
	}
	return event, nil
}

// AdjudicatorFinalStateRegisteredIterator is returned from FilterFinalStateRegistered and is used to iterate over the raw logs and unpacked data for FinalStateRegistered events raised by the Adjudicator contract.
type AdjudicatorFinalStateRegisteredIterator struct {
	Event *AdjudicatorFinalStateRegistered // Event containing the contract specifics and raw log

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
func (it *AdjudicatorFinalStateRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AdjudicatorFinalStateRegistered)
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
		it.Event = new(AdjudicatorFinalStateRegistered)
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
func (it *AdjudicatorFinalStateRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AdjudicatorFinalStateRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AdjudicatorFinalStateRegistered represents a FinalStateRegistered event raised by the Adjudicator contract.
type AdjudicatorFinalStateRegistered struct {
	ChannelID [32]byte
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterFinalStateRegistered is a free log retrieval operation binding the contract event 0xfa302ab93de3c7a9581de1f9182591df6335562d06dc23ea6c8af24a0e3d5c18.
//
// Solidity: event FinalStateRegistered(bytes32 indexed channelID)
func (_Adjudicator *AdjudicatorFilterer) FilterFinalStateRegistered(opts *bind.FilterOpts, channelID [][32]byte) (*AdjudicatorFinalStateRegisteredIterator, error) {

	var channelIDRule []interface{}
	for _, channelIDItem := range channelID {
		channelIDRule = append(channelIDRule, channelIDItem)
	}

	logs, sub, err := _Adjudicator.contract.FilterLogs(opts, "FinalStateRegistered", channelIDRule)
	if err != nil {
		return nil, err
	}
	return &AdjudicatorFinalStateRegisteredIterator{contract: _Adjudicator.contract, event: "FinalStateRegistered", logs: logs, sub: sub}, nil
}

// WatchFinalStateRegistered is a free log subscription operation binding the contract event 0xfa302ab93de3c7a9581de1f9182591df6335562d06dc23ea6c8af24a0e3d5c18.
//
// Solidity: event FinalStateRegistered(bytes32 indexed channelID)
func (_Adjudicator *AdjudicatorFilterer) WatchFinalStateRegistered(opts *bind.WatchOpts, sink chan<- *AdjudicatorFinalStateRegistered, channelID [][32]byte) (event.Subscription, error) {

	var channelIDRule []interface{}
	for _, channelIDItem := range channelID {
		channelIDRule = append(channelIDRule, channelIDItem)
	}

	logs, sub, err := _Adjudicator.contract.WatchLogs(opts, "FinalStateRegistered", channelIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AdjudicatorFinalStateRegistered)
				if err := _Adjudicator.contract.UnpackLog(event, "FinalStateRegistered", log); err != nil {
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

// ParseFinalStateRegistered is a log parse operation binding the contract event 0xfa302ab93de3c7a9581de1f9182591df6335562d06dc23ea6c8af24a0e3d5c18.
//
// Solidity: event FinalStateRegistered(bytes32 indexed channelID)
func (_Adjudicator *AdjudicatorFilterer) ParseFinalStateRegistered(log types.Log) (*AdjudicatorFinalStateRegistered, error) {
	event := new(AdjudicatorFinalStateRegistered)
	if err := _Adjudicator.contract.UnpackLog(event, "FinalStateRegistered", log); err != nil {
		return nil, err
	}
	return event, nil
}

// AdjudicatorPushOutcomeIterator is returned from FilterPushOutcome and is used to iterate over the raw logs and unpacked data for PushOutcome events raised by the Adjudicator contract.
type AdjudicatorPushOutcomeIterator struct {
	Event *AdjudicatorPushOutcome // Event containing the contract specifics and raw log

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
func (it *AdjudicatorPushOutcomeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AdjudicatorPushOutcome)
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
		it.Event = new(AdjudicatorPushOutcome)
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
func (it *AdjudicatorPushOutcomeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AdjudicatorPushOutcomeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AdjudicatorPushOutcome represents a PushOutcome event raised by the Adjudicator contract.
type AdjudicatorPushOutcome struct {
	ChannelID [32]byte
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterPushOutcome is a free log retrieval operation binding the contract event 0x18a580a4aab39f3e138ed4cf306861cb9702f09856253189563ccaec335f0ffb.
//
// Solidity: event PushOutcome(bytes32 indexed channelID)
func (_Adjudicator *AdjudicatorFilterer) FilterPushOutcome(opts *bind.FilterOpts, channelID [][32]byte) (*AdjudicatorPushOutcomeIterator, error) {

	var channelIDRule []interface{}
	for _, channelIDItem := range channelID {
		channelIDRule = append(channelIDRule, channelIDItem)
	}

	logs, sub, err := _Adjudicator.contract.FilterLogs(opts, "PushOutcome", channelIDRule)
	if err != nil {
		return nil, err
	}
	return &AdjudicatorPushOutcomeIterator{contract: _Adjudicator.contract, event: "PushOutcome", logs: logs, sub: sub}, nil
}

// WatchPushOutcome is a free log subscription operation binding the contract event 0x18a580a4aab39f3e138ed4cf306861cb9702f09856253189563ccaec335f0ffb.
//
// Solidity: event PushOutcome(bytes32 indexed channelID)
func (_Adjudicator *AdjudicatorFilterer) WatchPushOutcome(opts *bind.WatchOpts, sink chan<- *AdjudicatorPushOutcome, channelID [][32]byte) (event.Subscription, error) {

	var channelIDRule []interface{}
	for _, channelIDItem := range channelID {
		channelIDRule = append(channelIDRule, channelIDItem)
	}

	logs, sub, err := _Adjudicator.contract.WatchLogs(opts, "PushOutcome", channelIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AdjudicatorPushOutcome)
				if err := _Adjudicator.contract.UnpackLog(event, "PushOutcome", log); err != nil {
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

// ParsePushOutcome is a log parse operation binding the contract event 0x18a580a4aab39f3e138ed4cf306861cb9702f09856253189563ccaec335f0ffb.
//
// Solidity: event PushOutcome(bytes32 indexed channelID)
func (_Adjudicator *AdjudicatorFilterer) ParsePushOutcome(log types.Log) (*AdjudicatorPushOutcome, error) {
	event := new(AdjudicatorPushOutcome)
	if err := _Adjudicator.contract.UnpackLog(event, "PushOutcome", log); err != nil {
		return nil, err
	}
	return event, nil
}

// AdjudicatorRefutedIterator is returned from FilterRefuted and is used to iterate over the raw logs and unpacked data for Refuted events raised by the Adjudicator contract.
type AdjudicatorRefutedIterator struct {
	Event *AdjudicatorRefuted // Event containing the contract specifics and raw log

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
func (it *AdjudicatorRefutedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AdjudicatorRefuted)
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
		it.Event = new(AdjudicatorRefuted)
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
func (it *AdjudicatorRefutedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AdjudicatorRefutedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AdjudicatorRefuted represents a Refuted event raised by the Adjudicator contract.
type AdjudicatorRefuted struct {
	ChannelID [32]byte
	Version   *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterRefuted is a free log retrieval operation binding the contract event 0xd478cbccdd5ca6d246b145bb539b375b45c30ce42f63235b10ee19e4bc0f63c7.
//
// Solidity: event Refuted(bytes32 indexed channelID, uint256 version)
func (_Adjudicator *AdjudicatorFilterer) FilterRefuted(opts *bind.FilterOpts, channelID [][32]byte) (*AdjudicatorRefutedIterator, error) {

	var channelIDRule []interface{}
	for _, channelIDItem := range channelID {
		channelIDRule = append(channelIDRule, channelIDItem)
	}

	logs, sub, err := _Adjudicator.contract.FilterLogs(opts, "Refuted", channelIDRule)
	if err != nil {
		return nil, err
	}
	return &AdjudicatorRefutedIterator{contract: _Adjudicator.contract, event: "Refuted", logs: logs, sub: sub}, nil
}

// WatchRefuted is a free log subscription operation binding the contract event 0xd478cbccdd5ca6d246b145bb539b375b45c30ce42f63235b10ee19e4bc0f63c7.
//
// Solidity: event Refuted(bytes32 indexed channelID, uint256 version)
func (_Adjudicator *AdjudicatorFilterer) WatchRefuted(opts *bind.WatchOpts, sink chan<- *AdjudicatorRefuted, channelID [][32]byte) (event.Subscription, error) {

	var channelIDRule []interface{}
	for _, channelIDItem := range channelID {
		channelIDRule = append(channelIDRule, channelIDItem)
	}

	logs, sub, err := _Adjudicator.contract.WatchLogs(opts, "Refuted", channelIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AdjudicatorRefuted)
				if err := _Adjudicator.contract.UnpackLog(event, "Refuted", log); err != nil {
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

// ParseRefuted is a log parse operation binding the contract event 0xd478cbccdd5ca6d246b145bb539b375b45c30ce42f63235b10ee19e4bc0f63c7.
//
// Solidity: event Refuted(bytes32 indexed channelID, uint256 version)
func (_Adjudicator *AdjudicatorFilterer) ParseRefuted(log types.Log) (*AdjudicatorRefuted, error) {
	event := new(AdjudicatorRefuted)
	if err := _Adjudicator.contract.UnpackLog(event, "Refuted", log); err != nil {
		return nil, err
	}
	return event, nil
}

// AdjudicatorRegisteredIterator is returned from FilterRegistered and is used to iterate over the raw logs and unpacked data for Registered events raised by the Adjudicator contract.
type AdjudicatorRegisteredIterator struct {
	Event *AdjudicatorRegistered // Event containing the contract specifics and raw log

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
func (it *AdjudicatorRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AdjudicatorRegistered)
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
		it.Event = new(AdjudicatorRegistered)
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
func (it *AdjudicatorRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AdjudicatorRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AdjudicatorRegistered represents a Registered event raised by the Adjudicator contract.
type AdjudicatorRegistered struct {
	ChannelID [32]byte
	Version   *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterRegistered is a free log retrieval operation binding the contract event 0xc8704a622f3eb8c9fc5a2ddf1775b5ea7695359b13dec8111874e266a4d5fbc9.
//
// Solidity: event Registered(bytes32 indexed channelID, uint256 version)
func (_Adjudicator *AdjudicatorFilterer) FilterRegistered(opts *bind.FilterOpts, channelID [][32]byte) (*AdjudicatorRegisteredIterator, error) {

	var channelIDRule []interface{}
	for _, channelIDItem := range channelID {
		channelIDRule = append(channelIDRule, channelIDItem)
	}

	logs, sub, err := _Adjudicator.contract.FilterLogs(opts, "Registered", channelIDRule)
	if err != nil {
		return nil, err
	}
	return &AdjudicatorRegisteredIterator{contract: _Adjudicator.contract, event: "Registered", logs: logs, sub: sub}, nil
}

// WatchRegistered is a free log subscription operation binding the contract event 0xc8704a622f3eb8c9fc5a2ddf1775b5ea7695359b13dec8111874e266a4d5fbc9.
//
// Solidity: event Registered(bytes32 indexed channelID, uint256 version)
func (_Adjudicator *AdjudicatorFilterer) WatchRegistered(opts *bind.WatchOpts, sink chan<- *AdjudicatorRegistered, channelID [][32]byte) (event.Subscription, error) {

	var channelIDRule []interface{}
	for _, channelIDItem := range channelID {
		channelIDRule = append(channelIDRule, channelIDItem)
	}

	logs, sub, err := _Adjudicator.contract.WatchLogs(opts, "Registered", channelIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AdjudicatorRegistered)
				if err := _Adjudicator.contract.UnpackLog(event, "Registered", log); err != nil {
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

// ParseRegistered is a log parse operation binding the contract event 0xc8704a622f3eb8c9fc5a2ddf1775b5ea7695359b13dec8111874e266a4d5fbc9.
//
// Solidity: event Registered(bytes32 indexed channelID, uint256 version)
func (_Adjudicator *AdjudicatorFilterer) ParseRegistered(log types.Log) (*AdjudicatorRegistered, error) {
	event := new(AdjudicatorRegistered)
	if err := _Adjudicator.contract.UnpackLog(event, "Registered", log); err != nil {
		return nil, err
	}
	return event, nil
}

// AdjudicatorRespondedIterator is returned from FilterResponded and is used to iterate over the raw logs and unpacked data for Responded events raised by the Adjudicator contract.
type AdjudicatorRespondedIterator struct {
	Event *AdjudicatorResponded // Event containing the contract specifics and raw log

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
func (it *AdjudicatorRespondedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AdjudicatorResponded)
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
		it.Event = new(AdjudicatorResponded)
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
func (it *AdjudicatorRespondedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AdjudicatorRespondedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AdjudicatorResponded represents a Responded event raised by the Adjudicator contract.
type AdjudicatorResponded struct {
	ChannelID [32]byte
	Version   *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterResponded is a free log retrieval operation binding the contract event 0xf4793471fad8d7bbe3211ef7eed6bbef53a8f2e0593826ec24a97931e249b423.
//
// Solidity: event Responded(bytes32 indexed channelID, uint256 version)
func (_Adjudicator *AdjudicatorFilterer) FilterResponded(opts *bind.FilterOpts, channelID [][32]byte) (*AdjudicatorRespondedIterator, error) {

	var channelIDRule []interface{}
	for _, channelIDItem := range channelID {
		channelIDRule = append(channelIDRule, channelIDItem)
	}

	logs, sub, err := _Adjudicator.contract.FilterLogs(opts, "Responded", channelIDRule)
	if err != nil {
		return nil, err
	}
	return &AdjudicatorRespondedIterator{contract: _Adjudicator.contract, event: "Responded", logs: logs, sub: sub}, nil
}

// WatchResponded is a free log subscription operation binding the contract event 0xf4793471fad8d7bbe3211ef7eed6bbef53a8f2e0593826ec24a97931e249b423.
//
// Solidity: event Responded(bytes32 indexed channelID, uint256 version)
func (_Adjudicator *AdjudicatorFilterer) WatchResponded(opts *bind.WatchOpts, sink chan<- *AdjudicatorResponded, channelID [][32]byte) (event.Subscription, error) {

	var channelIDRule []interface{}
	for _, channelIDItem := range channelID {
		channelIDRule = append(channelIDRule, channelIDItem)
	}

	logs, sub, err := _Adjudicator.contract.WatchLogs(opts, "Responded", channelIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AdjudicatorResponded)
				if err := _Adjudicator.contract.UnpackLog(event, "Responded", log); err != nil {
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

// ParseResponded is a log parse operation binding the contract event 0xf4793471fad8d7bbe3211ef7eed6bbef53a8f2e0593826ec24a97931e249b423.
//
// Solidity: event Responded(bytes32 indexed channelID, uint256 version)
func (_Adjudicator *AdjudicatorFilterer) ParseResponded(log types.Log) (*AdjudicatorResponded, error) {
	event := new(AdjudicatorResponded)
	if err := _Adjudicator.contract.UnpackLog(event, "Responded", log); err != nil {
		return nil, err
	}
	return event, nil
}

// AdjudicatorStoredIterator is returned from FilterStored and is used to iterate over the raw logs and unpacked data for Stored events raised by the Adjudicator contract.
type AdjudicatorStoredIterator struct {
	Event *AdjudicatorStored // Event containing the contract specifics and raw log

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
func (it *AdjudicatorStoredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AdjudicatorStored)
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
		it.Event = new(AdjudicatorStored)
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
func (it *AdjudicatorStoredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AdjudicatorStoredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AdjudicatorStored represents a Stored event raised by the Adjudicator contract.
type AdjudicatorStored struct {
	ChannelID [32]byte
	Timeout   *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterStored is a free log retrieval operation binding the contract event 0xde02b1ac594e3d12f3797b91ed3e93213c5fcb9a6963fe4003c2fc8287e67c31.
//
// Solidity: event Stored(bytes32 indexed channelID, uint256 timeout)
func (_Adjudicator *AdjudicatorFilterer) FilterStored(opts *bind.FilterOpts, channelID [][32]byte) (*AdjudicatorStoredIterator, error) {

	var channelIDRule []interface{}
	for _, channelIDItem := range channelID {
		channelIDRule = append(channelIDRule, channelIDItem)
	}

	logs, sub, err := _Adjudicator.contract.FilterLogs(opts, "Stored", channelIDRule)
	if err != nil {
		return nil, err
	}
	return &AdjudicatorStoredIterator{contract: _Adjudicator.contract, event: "Stored", logs: logs, sub: sub}, nil
}

// WatchStored is a free log subscription operation binding the contract event 0xde02b1ac594e3d12f3797b91ed3e93213c5fcb9a6963fe4003c2fc8287e67c31.
//
// Solidity: event Stored(bytes32 indexed channelID, uint256 timeout)
func (_Adjudicator *AdjudicatorFilterer) WatchStored(opts *bind.WatchOpts, sink chan<- *AdjudicatorStored, channelID [][32]byte) (event.Subscription, error) {

	var channelIDRule []interface{}
	for _, channelIDItem := range channelID {
		channelIDRule = append(channelIDRule, channelIDItem)
	}

	logs, sub, err := _Adjudicator.contract.WatchLogs(opts, "Stored", channelIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AdjudicatorStored)
				if err := _Adjudicator.contract.UnpackLog(event, "Stored", log); err != nil {
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

// ParseStored is a log parse operation binding the contract event 0xde02b1ac594e3d12f3797b91ed3e93213c5fcb9a6963fe4003c2fc8287e67c31.
//
// Solidity: event Stored(bytes32 indexed channelID, uint256 timeout)
func (_Adjudicator *AdjudicatorFilterer) ParseStored(log types.Log) (*AdjudicatorStored, error) {
	event := new(AdjudicatorStored)
	if err := _Adjudicator.contract.UnpackLog(event, "Stored", log); err != nil {
		return nil, err
	}
	return event, nil
}

// AssetHolderABI is the input ABI used to generate the binding from.
const AssetHolderABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"participantID\",\"type\":\"bytes32\"}],\"name\":\"Deposited\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"}],\"name\":\"OutcomeSet\",\"type\":\"event\"},{\"constant\":true,\"inputs\":[],\"name\":\"Adjudicator\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"participantID\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"deposit\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"holdings\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"},{\"internalType\":\"address[]\",\"name\":\"parts\",\"type\":\"address[]\"},{\"internalType\":\"uint256[]\",\"name\":\"newBals\",\"type\":\"uint256[]\"},{\"internalType\":\"bytes32[]\",\"name\":\"subAllocs\",\"type\":\"bytes32[]\"},{\"internalType\":\"uint256[]\",\"name\":\"subBalances\",\"type\":\"uint256[]\"}],\"name\":\"setOutcome\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"settled\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"participant\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"receiver\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structAssetHolder.WithdrawalAuth\",\"name\":\"authorization\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"withdraw\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// AssetHolderFuncSigs maps the 4-byte function signature to its string representation.
var AssetHolderFuncSigs = map[string]string{
	"47c4aadf": "Adjudicator()",
	"1de26e16": "deposit(bytes32,uint256)",
	"ae9ee18c": "holdings(bytes32)",
	"79aad62e": "setOutcome(bytes32,address[],uint256[],bytes32[],uint256[])",
	"d945af1d": "settled(bytes32)",
	"4ed4283c": "withdraw((bytes32,address,address,uint256),bytes)",
}

// AssetHolder is an auto generated Go binding around an Ethereum contract.
type AssetHolder struct {
	AssetHolderCaller     // Read-only binding to the contract
	AssetHolderTransactor // Write-only binding to the contract
	AssetHolderFilterer   // Log filterer for contract events
}

// AssetHolderCaller is an auto generated read-only Go binding around an Ethereum contract.
type AssetHolderCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AssetHolderTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AssetHolderTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AssetHolderFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AssetHolderFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AssetHolderSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AssetHolderSession struct {
	Contract     *AssetHolder      // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// AssetHolderCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AssetHolderCallerSession struct {
	Contract *AssetHolderCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts      // Call options to use throughout this session
}

// AssetHolderTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AssetHolderTransactorSession struct {
	Contract     *AssetHolderTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// AssetHolderRaw is an auto generated low-level Go binding around an Ethereum contract.
type AssetHolderRaw struct {
	Contract *AssetHolder // Generic contract binding to access the raw methods on
}

// AssetHolderCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AssetHolderCallerRaw struct {
	Contract *AssetHolderCaller // Generic read-only contract binding to access the raw methods on
}

// AssetHolderTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AssetHolderTransactorRaw struct {
	Contract *AssetHolderTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAssetHolder creates a new instance of AssetHolder, bound to a specific deployed contract.
func NewAssetHolder(address common.Address, backend bind.ContractBackend) (*AssetHolder, error) {
	contract, err := bindAssetHolder(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &AssetHolder{AssetHolderCaller: AssetHolderCaller{contract: contract}, AssetHolderTransactor: AssetHolderTransactor{contract: contract}, AssetHolderFilterer: AssetHolderFilterer{contract: contract}}, nil
}

// NewAssetHolderCaller creates a new read-only instance of AssetHolder, bound to a specific deployed contract.
func NewAssetHolderCaller(address common.Address, caller bind.ContractCaller) (*AssetHolderCaller, error) {
	contract, err := bindAssetHolder(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AssetHolderCaller{contract: contract}, nil
}

// NewAssetHolderTransactor creates a new write-only instance of AssetHolder, bound to a specific deployed contract.
func NewAssetHolderTransactor(address common.Address, transactor bind.ContractTransactor) (*AssetHolderTransactor, error) {
	contract, err := bindAssetHolder(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AssetHolderTransactor{contract: contract}, nil
}

// NewAssetHolderFilterer creates a new log filterer instance of AssetHolder, bound to a specific deployed contract.
func NewAssetHolderFilterer(address common.Address, filterer bind.ContractFilterer) (*AssetHolderFilterer, error) {
	contract, err := bindAssetHolder(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AssetHolderFilterer{contract: contract}, nil
}

// bindAssetHolder binds a generic wrapper to an already deployed contract.
func bindAssetHolder(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(AssetHolderABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AssetHolder *AssetHolderRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _AssetHolder.Contract.AssetHolderCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AssetHolder *AssetHolderRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AssetHolder.Contract.AssetHolderTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AssetHolder *AssetHolderRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AssetHolder.Contract.AssetHolderTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AssetHolder *AssetHolderCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _AssetHolder.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AssetHolder *AssetHolderTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AssetHolder.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AssetHolder *AssetHolderTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AssetHolder.Contract.contract.Transact(opts, method, params...)
}

// AssetHolderWithdrawalAuth is an auto generated low-level Go binding around an user-defined struct.
type AssetHolderWithdrawalAuth struct {
	ChannelID   [32]byte
	Participant common.Address
	Receiver    common.Address
	Amount      *big.Int
}

// Adjudicator is a free data retrieval call binding the contract method 0x47c4aadf.
//
// Solidity: function Adjudicator() constant returns(address)
func (_AssetHolder *AssetHolderCaller) Adjudicator(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _AssetHolder.contract.Call(opts, out, "Adjudicator")
	return *ret0, err
}

// Adjudicator is a free data retrieval call binding the contract method 0x47c4aadf.
//
// Solidity: function Adjudicator() constant returns(address)
func (_AssetHolder *AssetHolderSession) Adjudicator() (common.Address, error) {
	return _AssetHolder.Contract.Adjudicator(&_AssetHolder.CallOpts)
}

// Adjudicator is a free data retrieval call binding the contract method 0x47c4aadf.
//
// Solidity: function Adjudicator() constant returns(address)
func (_AssetHolder *AssetHolderCallerSession) Adjudicator() (common.Address, error) {
	return _AssetHolder.Contract.Adjudicator(&_AssetHolder.CallOpts)
}

// Holdings is a free data retrieval call binding the contract method 0xae9ee18c.
//
// Solidity: function holdings(bytes32 ) constant returns(uint256)
func (_AssetHolder *AssetHolderCaller) Holdings(opts *bind.CallOpts, arg0 [32]byte) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _AssetHolder.contract.Call(opts, out, "holdings", arg0)
	return *ret0, err
}

// Holdings is a free data retrieval call binding the contract method 0xae9ee18c.
//
// Solidity: function holdings(bytes32 ) constant returns(uint256)
func (_AssetHolder *AssetHolderSession) Holdings(arg0 [32]byte) (*big.Int, error) {
	return _AssetHolder.Contract.Holdings(&_AssetHolder.CallOpts, arg0)
}

// Holdings is a free data retrieval call binding the contract method 0xae9ee18c.
//
// Solidity: function holdings(bytes32 ) constant returns(uint256)
func (_AssetHolder *AssetHolderCallerSession) Holdings(arg0 [32]byte) (*big.Int, error) {
	return _AssetHolder.Contract.Holdings(&_AssetHolder.CallOpts, arg0)
}

// Settled is a free data retrieval call binding the contract method 0xd945af1d.
//
// Solidity: function settled(bytes32 ) constant returns(bool)
func (_AssetHolder *AssetHolderCaller) Settled(opts *bind.CallOpts, arg0 [32]byte) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _AssetHolder.contract.Call(opts, out, "settled", arg0)
	return *ret0, err
}

// Settled is a free data retrieval call binding the contract method 0xd945af1d.
//
// Solidity: function settled(bytes32 ) constant returns(bool)
func (_AssetHolder *AssetHolderSession) Settled(arg0 [32]byte) (bool, error) {
	return _AssetHolder.Contract.Settled(&_AssetHolder.CallOpts, arg0)
}

// Settled is a free data retrieval call binding the contract method 0xd945af1d.
//
// Solidity: function settled(bytes32 ) constant returns(bool)
func (_AssetHolder *AssetHolderCallerSession) Settled(arg0 [32]byte) (bool, error) {
	return _AssetHolder.Contract.Settled(&_AssetHolder.CallOpts, arg0)
}

// Deposit is a paid mutator transaction binding the contract method 0x1de26e16.
//
// Solidity: function deposit(bytes32 participantID, uint256 amount) returns()
func (_AssetHolder *AssetHolderTransactor) Deposit(opts *bind.TransactOpts, participantID [32]byte, amount *big.Int) (*types.Transaction, error) {
	return _AssetHolder.contract.Transact(opts, "deposit", participantID, amount)
}

// Deposit is a paid mutator transaction binding the contract method 0x1de26e16.
//
// Solidity: function deposit(bytes32 participantID, uint256 amount) returns()
func (_AssetHolder *AssetHolderSession) Deposit(participantID [32]byte, amount *big.Int) (*types.Transaction, error) {
	return _AssetHolder.Contract.Deposit(&_AssetHolder.TransactOpts, participantID, amount)
}

// Deposit is a paid mutator transaction binding the contract method 0x1de26e16.
//
// Solidity: function deposit(bytes32 participantID, uint256 amount) returns()
func (_AssetHolder *AssetHolderTransactorSession) Deposit(participantID [32]byte, amount *big.Int) (*types.Transaction, error) {
	return _AssetHolder.Contract.Deposit(&_AssetHolder.TransactOpts, participantID, amount)
}

// SetOutcome is a paid mutator transaction binding the contract method 0x79aad62e.
//
// Solidity: function setOutcome(bytes32 channelID, address[] parts, uint256[] newBals, bytes32[] subAllocs, uint256[] subBalances) returns()
func (_AssetHolder *AssetHolderTransactor) SetOutcome(opts *bind.TransactOpts, channelID [32]byte, parts []common.Address, newBals []*big.Int, subAllocs [][32]byte, subBalances []*big.Int) (*types.Transaction, error) {
	return _AssetHolder.contract.Transact(opts, "setOutcome", channelID, parts, newBals, subAllocs, subBalances)
}

// SetOutcome is a paid mutator transaction binding the contract method 0x79aad62e.
//
// Solidity: function setOutcome(bytes32 channelID, address[] parts, uint256[] newBals, bytes32[] subAllocs, uint256[] subBalances) returns()
func (_AssetHolder *AssetHolderSession) SetOutcome(channelID [32]byte, parts []common.Address, newBals []*big.Int, subAllocs [][32]byte, subBalances []*big.Int) (*types.Transaction, error) {
	return _AssetHolder.Contract.SetOutcome(&_AssetHolder.TransactOpts, channelID, parts, newBals, subAllocs, subBalances)
}

// SetOutcome is a paid mutator transaction binding the contract method 0x79aad62e.
//
// Solidity: function setOutcome(bytes32 channelID, address[] parts, uint256[] newBals, bytes32[] subAllocs, uint256[] subBalances) returns()
func (_AssetHolder *AssetHolderTransactorSession) SetOutcome(channelID [32]byte, parts []common.Address, newBals []*big.Int, subAllocs [][32]byte, subBalances []*big.Int) (*types.Transaction, error) {
	return _AssetHolder.Contract.SetOutcome(&_AssetHolder.TransactOpts, channelID, parts, newBals, subAllocs, subBalances)
}

// Withdraw is a paid mutator transaction binding the contract method 0x4ed4283c.
//
// Solidity: function withdraw(AssetHolderWithdrawalAuth authorization, bytes signature) returns()
func (_AssetHolder *AssetHolderTransactor) Withdraw(opts *bind.TransactOpts, authorization AssetHolderWithdrawalAuth, signature []byte) (*types.Transaction, error) {
	return _AssetHolder.contract.Transact(opts, "withdraw", authorization, signature)
}

// Withdraw is a paid mutator transaction binding the contract method 0x4ed4283c.
//
// Solidity: function withdraw(AssetHolderWithdrawalAuth authorization, bytes signature) returns()
func (_AssetHolder *AssetHolderSession) Withdraw(authorization AssetHolderWithdrawalAuth, signature []byte) (*types.Transaction, error) {
	return _AssetHolder.Contract.Withdraw(&_AssetHolder.TransactOpts, authorization, signature)
}

// Withdraw is a paid mutator transaction binding the contract method 0x4ed4283c.
//
// Solidity: function withdraw(AssetHolderWithdrawalAuth authorization, bytes signature) returns()
func (_AssetHolder *AssetHolderTransactorSession) Withdraw(authorization AssetHolderWithdrawalAuth, signature []byte) (*types.Transaction, error) {
	return _AssetHolder.Contract.Withdraw(&_AssetHolder.TransactOpts, authorization, signature)
}

// AssetHolderDepositedIterator is returned from FilterDeposited and is used to iterate over the raw logs and unpacked data for Deposited events raised by the AssetHolder contract.
type AssetHolderDepositedIterator struct {
	Event *AssetHolderDeposited // Event containing the contract specifics and raw log

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
func (it *AssetHolderDepositedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AssetHolderDeposited)
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
		it.Event = new(AssetHolderDeposited)
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
func (it *AssetHolderDepositedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AssetHolderDepositedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AssetHolderDeposited represents a Deposited event raised by the AssetHolder contract.
type AssetHolderDeposited struct {
	ParticipantID [32]byte
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterDeposited is a free log retrieval operation binding the contract event 0x5d6bcb8d3a72f0688817569bc00aba553820f312e9260b6c7da291b97bf13367.
//
// Solidity: event Deposited(bytes32 indexed participantID)
func (_AssetHolder *AssetHolderFilterer) FilterDeposited(opts *bind.FilterOpts, participantID [][32]byte) (*AssetHolderDepositedIterator, error) {

	var participantIDRule []interface{}
	for _, participantIDItem := range participantID {
		participantIDRule = append(participantIDRule, participantIDItem)
	}

	logs, sub, err := _AssetHolder.contract.FilterLogs(opts, "Deposited", participantIDRule)
	if err != nil {
		return nil, err
	}
	return &AssetHolderDepositedIterator{contract: _AssetHolder.contract, event: "Deposited", logs: logs, sub: sub}, nil
}

// WatchDeposited is a free log subscription operation binding the contract event 0x5d6bcb8d3a72f0688817569bc00aba553820f312e9260b6c7da291b97bf13367.
//
// Solidity: event Deposited(bytes32 indexed participantID)
func (_AssetHolder *AssetHolderFilterer) WatchDeposited(opts *bind.WatchOpts, sink chan<- *AssetHolderDeposited, participantID [][32]byte) (event.Subscription, error) {

	var participantIDRule []interface{}
	for _, participantIDItem := range participantID {
		participantIDRule = append(participantIDRule, participantIDItem)
	}

	logs, sub, err := _AssetHolder.contract.WatchLogs(opts, "Deposited", participantIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AssetHolderDeposited)
				if err := _AssetHolder.contract.UnpackLog(event, "Deposited", log); err != nil {
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

// ParseDeposited is a log parse operation binding the contract event 0x5d6bcb8d3a72f0688817569bc00aba553820f312e9260b6c7da291b97bf13367.
//
// Solidity: event Deposited(bytes32 indexed participantID)
func (_AssetHolder *AssetHolderFilterer) ParseDeposited(log types.Log) (*AssetHolderDeposited, error) {
	event := new(AssetHolderDeposited)
	if err := _AssetHolder.contract.UnpackLog(event, "Deposited", log); err != nil {
		return nil, err
	}
	return event, nil
}

// AssetHolderOutcomeSetIterator is returned from FilterOutcomeSet and is used to iterate over the raw logs and unpacked data for OutcomeSet events raised by the AssetHolder contract.
type AssetHolderOutcomeSetIterator struct {
	Event *AssetHolderOutcomeSet // Event containing the contract specifics and raw log

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
func (it *AssetHolderOutcomeSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AssetHolderOutcomeSet)
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
		it.Event = new(AssetHolderOutcomeSet)
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
func (it *AssetHolderOutcomeSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AssetHolderOutcomeSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AssetHolderOutcomeSet represents a OutcomeSet event raised by the AssetHolder contract.
type AssetHolderOutcomeSet struct {
	ChannelID [32]byte
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterOutcomeSet is a free log retrieval operation binding the contract event 0xef898d6cd3395b6dfe67a3c1923e5c726c1b154e979fb0a25a9c41d0093168b8.
//
// Solidity: event OutcomeSet(bytes32 indexed channelID)
func (_AssetHolder *AssetHolderFilterer) FilterOutcomeSet(opts *bind.FilterOpts, channelID [][32]byte) (*AssetHolderOutcomeSetIterator, error) {

	var channelIDRule []interface{}
	for _, channelIDItem := range channelID {
		channelIDRule = append(channelIDRule, channelIDItem)
	}

	logs, sub, err := _AssetHolder.contract.FilterLogs(opts, "OutcomeSet", channelIDRule)
	if err != nil {
		return nil, err
	}
	return &AssetHolderOutcomeSetIterator{contract: _AssetHolder.contract, event: "OutcomeSet", logs: logs, sub: sub}, nil
}

// WatchOutcomeSet is a free log subscription operation binding the contract event 0xef898d6cd3395b6dfe67a3c1923e5c726c1b154e979fb0a25a9c41d0093168b8.
//
// Solidity: event OutcomeSet(bytes32 indexed channelID)
func (_AssetHolder *AssetHolderFilterer) WatchOutcomeSet(opts *bind.WatchOpts, sink chan<- *AssetHolderOutcomeSet, channelID [][32]byte) (event.Subscription, error) {

	var channelIDRule []interface{}
	for _, channelIDItem := range channelID {
		channelIDRule = append(channelIDRule, channelIDItem)
	}

	logs, sub, err := _AssetHolder.contract.WatchLogs(opts, "OutcomeSet", channelIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AssetHolderOutcomeSet)
				if err := _AssetHolder.contract.UnpackLog(event, "OutcomeSet", log); err != nil {
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
func (_AssetHolder *AssetHolderFilterer) ParseOutcomeSet(log types.Log) (*AssetHolderOutcomeSet, error) {
	event := new(AssetHolderOutcomeSet)
	if err := _AssetHolder.contract.UnpackLog(event, "OutcomeSet", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ECDSAABI is the input ABI used to generate the binding from.
const ECDSAABI = "[]"

// ECDSABin is the compiled bytecode used for deploying new contracts.
var ECDSABin = "0x60556023600b82828239805160001a607314601657fe5b30600052607381538281f3fe73000000000000000000000000000000000000000030146080604052600080fdfea265627a7a72315820342d710f451a2eff59d1aa19ac201440efa44bb9a1066bfdd87a318beb77e52764736f6c634300050c0032"

// DeployECDSA deploys a new Ethereum contract, binding an instance of ECDSA to it.
func DeployECDSA(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *ECDSA, error) {
	parsed, err := abi.JSON(strings.NewReader(ECDSAABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(ECDSABin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &ECDSA{ECDSACaller: ECDSACaller{contract: contract}, ECDSATransactor: ECDSATransactor{contract: contract}, ECDSAFilterer: ECDSAFilterer{contract: contract}}, nil
}

// ECDSA is an auto generated Go binding around an Ethereum contract.
type ECDSA struct {
	ECDSACaller     // Read-only binding to the contract
	ECDSATransactor // Write-only binding to the contract
	ECDSAFilterer   // Log filterer for contract events
}

// ECDSACaller is an auto generated read-only Go binding around an Ethereum contract.
type ECDSACaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ECDSATransactor is an auto generated write-only Go binding around an Ethereum contract.
type ECDSATransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ECDSAFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ECDSAFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ECDSASession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ECDSASession struct {
	Contract     *ECDSA            // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ECDSACallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ECDSACallerSession struct {
	Contract *ECDSACaller  // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// ECDSATransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ECDSATransactorSession struct {
	Contract     *ECDSATransactor  // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ECDSARaw is an auto generated low-level Go binding around an Ethereum contract.
type ECDSARaw struct {
	Contract *ECDSA // Generic contract binding to access the raw methods on
}

// ECDSACallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ECDSACallerRaw struct {
	Contract *ECDSACaller // Generic read-only contract binding to access the raw methods on
}

// ECDSATransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ECDSATransactorRaw struct {
	Contract *ECDSATransactor // Generic write-only contract binding to access the raw methods on
}

// NewECDSA creates a new instance of ECDSA, bound to a specific deployed contract.
func NewECDSA(address common.Address, backend bind.ContractBackend) (*ECDSA, error) {
	contract, err := bindECDSA(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ECDSA{ECDSACaller: ECDSACaller{contract: contract}, ECDSATransactor: ECDSATransactor{contract: contract}, ECDSAFilterer: ECDSAFilterer{contract: contract}}, nil
}

// NewECDSACaller creates a new read-only instance of ECDSA, bound to a specific deployed contract.
func NewECDSACaller(address common.Address, caller bind.ContractCaller) (*ECDSACaller, error) {
	contract, err := bindECDSA(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ECDSACaller{contract: contract}, nil
}

// NewECDSATransactor creates a new write-only instance of ECDSA, bound to a specific deployed contract.
func NewECDSATransactor(address common.Address, transactor bind.ContractTransactor) (*ECDSATransactor, error) {
	contract, err := bindECDSA(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ECDSATransactor{contract: contract}, nil
}

// NewECDSAFilterer creates a new log filterer instance of ECDSA, bound to a specific deployed contract.
func NewECDSAFilterer(address common.Address, filterer bind.ContractFilterer) (*ECDSAFilterer, error) {
	contract, err := bindECDSA(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ECDSAFilterer{contract: contract}, nil
}

// bindECDSA binds a generic wrapper to an already deployed contract.
func bindECDSA(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ECDSAABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ECDSA *ECDSARaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _ECDSA.Contract.ECDSACaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ECDSA *ECDSARaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ECDSA.Contract.ECDSATransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ECDSA *ECDSARaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ECDSA.Contract.ECDSATransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ECDSA *ECDSACallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _ECDSA.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ECDSA *ECDSATransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ECDSA.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ECDSA *ECDSATransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ECDSA.Contract.contract.Transact(opts, method, params...)
}

// PerunTypesABI is the input ABI used to generate the binding from.
const PerunTypesABI = "[]"

// PerunTypesBin is the compiled bytecode used for deploying new contracts.
var PerunTypesBin = "0x60636023600b82828239805160001a607314601657fe5b30600052607381538281f3fe73000000000000000000000000000000000000000030146080604052600080fdfea365627a7a72315820f5dcb7643880c56f4ea9a0584a88d8b7f19fe35e6a2aa577caef516d2af84fdc6c6578706572696d656e74616cf564736f6c634300050c0040"

// DeployPerunTypes deploys a new Ethereum contract, binding an instance of PerunTypes to it.
func DeployPerunTypes(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *PerunTypes, error) {
	parsed, err := abi.JSON(strings.NewReader(PerunTypesABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(PerunTypesBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &PerunTypes{PerunTypesCaller: PerunTypesCaller{contract: contract}, PerunTypesTransactor: PerunTypesTransactor{contract: contract}, PerunTypesFilterer: PerunTypesFilterer{contract: contract}}, nil
}

// PerunTypes is an auto generated Go binding around an Ethereum contract.
type PerunTypes struct {
	PerunTypesCaller     // Read-only binding to the contract
	PerunTypesTransactor // Write-only binding to the contract
	PerunTypesFilterer   // Log filterer for contract events
}

// PerunTypesCaller is an auto generated read-only Go binding around an Ethereum contract.
type PerunTypesCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PerunTypesTransactor is an auto generated write-only Go binding around an Ethereum contract.
type PerunTypesTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PerunTypesFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type PerunTypesFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PerunTypesSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type PerunTypesSession struct {
	Contract     *PerunTypes       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// PerunTypesCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type PerunTypesCallerSession struct {
	Contract *PerunTypesCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// PerunTypesTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type PerunTypesTransactorSession struct {
	Contract     *PerunTypesTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// PerunTypesRaw is an auto generated low-level Go binding around an Ethereum contract.
type PerunTypesRaw struct {
	Contract *PerunTypes // Generic contract binding to access the raw methods on
}

// PerunTypesCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type PerunTypesCallerRaw struct {
	Contract *PerunTypesCaller // Generic read-only contract binding to access the raw methods on
}

// PerunTypesTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type PerunTypesTransactorRaw struct {
	Contract *PerunTypesTransactor // Generic write-only contract binding to access the raw methods on
}

// NewPerunTypes creates a new instance of PerunTypes, bound to a specific deployed contract.
func NewPerunTypes(address common.Address, backend bind.ContractBackend) (*PerunTypes, error) {
	contract, err := bindPerunTypes(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &PerunTypes{PerunTypesCaller: PerunTypesCaller{contract: contract}, PerunTypesTransactor: PerunTypesTransactor{contract: contract}, PerunTypesFilterer: PerunTypesFilterer{contract: contract}}, nil
}

// NewPerunTypesCaller creates a new read-only instance of PerunTypes, bound to a specific deployed contract.
func NewPerunTypesCaller(address common.Address, caller bind.ContractCaller) (*PerunTypesCaller, error) {
	contract, err := bindPerunTypes(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &PerunTypesCaller{contract: contract}, nil
}

// NewPerunTypesTransactor creates a new write-only instance of PerunTypes, bound to a specific deployed contract.
func NewPerunTypesTransactor(address common.Address, transactor bind.ContractTransactor) (*PerunTypesTransactor, error) {
	contract, err := bindPerunTypes(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &PerunTypesTransactor{contract: contract}, nil
}

// NewPerunTypesFilterer creates a new log filterer instance of PerunTypes, bound to a specific deployed contract.
func NewPerunTypesFilterer(address common.Address, filterer bind.ContractFilterer) (*PerunTypesFilterer, error) {
	contract, err := bindPerunTypes(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &PerunTypesFilterer{contract: contract}, nil
}

// bindPerunTypes binds a generic wrapper to an already deployed contract.
func bindPerunTypes(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(PerunTypesABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PerunTypes *PerunTypesRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _PerunTypes.Contract.PerunTypesCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PerunTypes *PerunTypesRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PerunTypes.Contract.PerunTypesTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PerunTypes *PerunTypesRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PerunTypes.Contract.PerunTypesTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PerunTypes *PerunTypesCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _PerunTypes.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PerunTypes *PerunTypesTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PerunTypes.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PerunTypes *PerunTypesTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PerunTypes.Contract.contract.Transact(opts, method, params...)
}

// SafeMathABI is the input ABI used to generate the binding from.
const SafeMathABI = "[]"

// SafeMathBin is the compiled bytecode used for deploying new contracts.
var SafeMathBin = "0x60556023600b82828239805160001a607314601657fe5b30600052607381538281f3fe73000000000000000000000000000000000000000030146080604052600080fdfea265627a7a7231582056ed00141f84dd6aa03a937f6081f662a9e2dec0fabdb9a6bb944fd39723ff0164736f6c634300050c0032"

// DeploySafeMath deploys a new Ethereum contract, binding an instance of SafeMath to it.
func DeploySafeMath(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *SafeMath, error) {
	parsed, err := abi.JSON(strings.NewReader(SafeMathABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(SafeMathBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &SafeMath{SafeMathCaller: SafeMathCaller{contract: contract}, SafeMathTransactor: SafeMathTransactor{contract: contract}, SafeMathFilterer: SafeMathFilterer{contract: contract}}, nil
}

// SafeMath is an auto generated Go binding around an Ethereum contract.
type SafeMath struct {
	SafeMathCaller     // Read-only binding to the contract
	SafeMathTransactor // Write-only binding to the contract
	SafeMathFilterer   // Log filterer for contract events
}

// SafeMathCaller is an auto generated read-only Go binding around an Ethereum contract.
type SafeMathCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SafeMathTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SafeMathTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SafeMathFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SafeMathFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SafeMathSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SafeMathSession struct {
	Contract     *SafeMath         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SafeMathCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SafeMathCallerSession struct {
	Contract *SafeMathCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// SafeMathTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SafeMathTransactorSession struct {
	Contract     *SafeMathTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// SafeMathRaw is an auto generated low-level Go binding around an Ethereum contract.
type SafeMathRaw struct {
	Contract *SafeMath // Generic contract binding to access the raw methods on
}

// SafeMathCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SafeMathCallerRaw struct {
	Contract *SafeMathCaller // Generic read-only contract binding to access the raw methods on
}

// SafeMathTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SafeMathTransactorRaw struct {
	Contract *SafeMathTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSafeMath creates a new instance of SafeMath, bound to a specific deployed contract.
func NewSafeMath(address common.Address, backend bind.ContractBackend) (*SafeMath, error) {
	contract, err := bindSafeMath(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SafeMath{SafeMathCaller: SafeMathCaller{contract: contract}, SafeMathTransactor: SafeMathTransactor{contract: contract}, SafeMathFilterer: SafeMathFilterer{contract: contract}}, nil
}

// NewSafeMathCaller creates a new read-only instance of SafeMath, bound to a specific deployed contract.
func NewSafeMathCaller(address common.Address, caller bind.ContractCaller) (*SafeMathCaller, error) {
	contract, err := bindSafeMath(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SafeMathCaller{contract: contract}, nil
}

// NewSafeMathTransactor creates a new write-only instance of SafeMath, bound to a specific deployed contract.
func NewSafeMathTransactor(address common.Address, transactor bind.ContractTransactor) (*SafeMathTransactor, error) {
	contract, err := bindSafeMath(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SafeMathTransactor{contract: contract}, nil
}

// NewSafeMathFilterer creates a new log filterer instance of SafeMath, bound to a specific deployed contract.
func NewSafeMathFilterer(address common.Address, filterer bind.ContractFilterer) (*SafeMathFilterer, error) {
	contract, err := bindSafeMath(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SafeMathFilterer{contract: contract}, nil
}

// bindSafeMath binds a generic wrapper to an already deployed contract.
func bindSafeMath(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SafeMathABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SafeMath *SafeMathRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _SafeMath.Contract.SafeMathCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SafeMath *SafeMathRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SafeMath.Contract.SafeMathTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SafeMath *SafeMathRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SafeMath.Contract.SafeMathTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SafeMath *SafeMathCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _SafeMath.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SafeMath *SafeMathTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SafeMath.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SafeMath *SafeMathTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SafeMath.Contract.contract.Transact(opts, method, params...)
}

// ValidTransitionerABI is the input ABI used to generate the binding from.
const ValidTransitionerABI = "[{\"constant\":true,\"inputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"challengeDuration\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"app\",\"type\":\"address\"},{\"internalType\":\"address[]\",\"name\":\"participants\",\"type\":\"address[]\"}],\"internalType\":\"structPerunTypes.Params\",\"name\":\"params\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"},{\"components\":[{\"internalType\":\"address[]\",\"name\":\"assets\",\"type\":\"address[]\"},{\"internalType\":\"uint256[][]\",\"name\":\"balances\",\"type\":\"uint256[][]\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"ID\",\"type\":\"bytes32\"},{\"internalType\":\"uint256[]\",\"name\":\"balances\",\"type\":\"uint256[]\"}],\"internalType\":\"structPerunTypes.SubAlloc[]\",\"name\":\"locked\",\"type\":\"tuple[]\"}],\"internalType\":\"structPerunTypes.Allocation\",\"name\":\"outcome\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"appData\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"isFinal\",\"type\":\"bool\"}],\"internalType\":\"structPerunTypes.State\",\"name\":\"from\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"},{\"components\":[{\"internalType\":\"address[]\",\"name\":\"assets\",\"type\":\"address[]\"},{\"internalType\":\"uint256[][]\",\"name\":\"balances\",\"type\":\"uint256[][]\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"ID\",\"type\":\"bytes32\"},{\"internalType\":\"uint256[]\",\"name\":\"balances\",\"type\":\"uint256[]\"}],\"internalType\":\"structPerunTypes.SubAlloc[]\",\"name\":\"locked\",\"type\":\"tuple[]\"}],\"internalType\":\"structPerunTypes.Allocation\",\"name\":\"outcome\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"appData\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"isFinal\",\"type\":\"bool\"}],\"internalType\":\"structPerunTypes.State\",\"name\":\"to\",\"type\":\"tuple\"},{\"internalType\":\"uint256\",\"name\":\"moverIdx\",\"type\":\"uint256\"}],\"name\":\"validTransition\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"}]"

// ValidTransitionerFuncSigs maps the 4-byte function signature to its string representation.
var ValidTransitionerFuncSigs = map[string]string{
	"ec29dd7e": "validTransition((uint256,uint256,address,address[]),(bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool),(bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool),uint256)",
}

// ValidTransitioner is an auto generated Go binding around an Ethereum contract.
type ValidTransitioner struct {
	ValidTransitionerCaller     // Read-only binding to the contract
	ValidTransitionerTransactor // Write-only binding to the contract
	ValidTransitionerFilterer   // Log filterer for contract events
}

// ValidTransitionerCaller is an auto generated read-only Go binding around an Ethereum contract.
type ValidTransitionerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ValidTransitionerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ValidTransitionerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ValidTransitionerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ValidTransitionerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ValidTransitionerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ValidTransitionerSession struct {
	Contract     *ValidTransitioner // Generic contract binding to set the session for
	CallOpts     bind.CallOpts      // Call options to use throughout this session
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// ValidTransitionerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ValidTransitionerCallerSession struct {
	Contract *ValidTransitionerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts            // Call options to use throughout this session
}

// ValidTransitionerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ValidTransitionerTransactorSession struct {
	Contract     *ValidTransitionerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts            // Transaction auth options to use throughout this session
}

// ValidTransitionerRaw is an auto generated low-level Go binding around an Ethereum contract.
type ValidTransitionerRaw struct {
	Contract *ValidTransitioner // Generic contract binding to access the raw methods on
}

// ValidTransitionerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ValidTransitionerCallerRaw struct {
	Contract *ValidTransitionerCaller // Generic read-only contract binding to access the raw methods on
}

// ValidTransitionerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ValidTransitionerTransactorRaw struct {
	Contract *ValidTransitionerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewValidTransitioner creates a new instance of ValidTransitioner, bound to a specific deployed contract.
func NewValidTransitioner(address common.Address, backend bind.ContractBackend) (*ValidTransitioner, error) {
	contract, err := bindValidTransitioner(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ValidTransitioner{ValidTransitionerCaller: ValidTransitionerCaller{contract: contract}, ValidTransitionerTransactor: ValidTransitionerTransactor{contract: contract}, ValidTransitionerFilterer: ValidTransitionerFilterer{contract: contract}}, nil
}

// NewValidTransitionerCaller creates a new read-only instance of ValidTransitioner, bound to a specific deployed contract.
func NewValidTransitionerCaller(address common.Address, caller bind.ContractCaller) (*ValidTransitionerCaller, error) {
	contract, err := bindValidTransitioner(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ValidTransitionerCaller{contract: contract}, nil
}

// NewValidTransitionerTransactor creates a new write-only instance of ValidTransitioner, bound to a specific deployed contract.
func NewValidTransitionerTransactor(address common.Address, transactor bind.ContractTransactor) (*ValidTransitionerTransactor, error) {
	contract, err := bindValidTransitioner(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ValidTransitionerTransactor{contract: contract}, nil
}

// NewValidTransitionerFilterer creates a new log filterer instance of ValidTransitioner, bound to a specific deployed contract.
func NewValidTransitionerFilterer(address common.Address, filterer bind.ContractFilterer) (*ValidTransitionerFilterer, error) {
	contract, err := bindValidTransitioner(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ValidTransitionerFilterer{contract: contract}, nil
}

// bindValidTransitioner binds a generic wrapper to an already deployed contract.
func bindValidTransitioner(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ValidTransitionerABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ValidTransitioner *ValidTransitionerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _ValidTransitioner.Contract.ValidTransitionerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ValidTransitioner *ValidTransitionerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ValidTransitioner.Contract.ValidTransitionerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ValidTransitioner *ValidTransitionerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ValidTransitioner.Contract.ValidTransitionerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ValidTransitioner *ValidTransitionerCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _ValidTransitioner.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ValidTransitioner *ValidTransitionerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ValidTransitioner.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ValidTransitioner *ValidTransitionerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ValidTransitioner.Contract.contract.Transact(opts, method, params...)
}

// PerunTypesAllocation is an auto generated low-level Go binding around an user-defined struct.
type PerunTypesAllocation struct {
	Assets   []common.Address
	Balances [][]*big.Int
	Locked   []PerunTypesSubAlloc
}

// PerunTypesParams is an auto generated low-level Go binding around an user-defined struct.
type PerunTypesParams struct {
	ChallengeDuration *big.Int
	Nonce             *big.Int
	App               common.Address
	Participants      []common.Address
}

// PerunTypesState is an auto generated low-level Go binding around an user-defined struct.
type PerunTypesState struct {
	ChannelID [32]byte
	Version   uint64
	Outcome   PerunTypesAllocation
	AppData   []byte
	IsFinal   bool
}

// PerunTypesSubAlloc is an auto generated low-level Go binding around an user-defined struct.
type PerunTypesSubAlloc struct {
	ID       [32]byte
	Balances []*big.Int
}

// ValidTransition is a free data retrieval call binding the contract method 0xec29dd7e.
//
// Solidity: function validTransition(PerunTypesParams params, PerunTypesState from, PerunTypesState to, uint256 moverIdx) constant returns(bool)
func (_ValidTransitioner *ValidTransitionerCaller) ValidTransition(opts *bind.CallOpts, params PerunTypesParams, from PerunTypesState, to PerunTypesState, moverIdx *big.Int) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _ValidTransitioner.contract.Call(opts, out, "validTransition", params, from, to, moverIdx)
	return *ret0, err
}

// ValidTransition is a free data retrieval call binding the contract method 0xec29dd7e.
//
// Solidity: function validTransition(PerunTypesParams params, PerunTypesState from, PerunTypesState to, uint256 moverIdx) constant returns(bool)
func (_ValidTransitioner *ValidTransitionerSession) ValidTransition(params PerunTypesParams, from PerunTypesState, to PerunTypesState, moverIdx *big.Int) (bool, error) {
	return _ValidTransitioner.Contract.ValidTransition(&_ValidTransitioner.CallOpts, params, from, to, moverIdx)
}

// ValidTransition is a free data retrieval call binding the contract method 0xec29dd7e.
//
// Solidity: function validTransition(PerunTypesParams params, PerunTypesState from, PerunTypesState to, uint256 moverIdx) constant returns(bool)
func (_ValidTransitioner *ValidTransitionerCallerSession) ValidTransition(params PerunTypesParams, from PerunTypesState, to PerunTypesState, moverIdx *big.Int) (bool, error) {
	return _ValidTransitioner.Contract.ValidTransition(&_ValidTransitioner.CallOpts, params, from, to, moverIdx)
}
