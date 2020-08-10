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

// AdjudicatorABI is the input ABI used to generate the binding from.
const AdjudicatorABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"}],\"name\":\"Concluded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"}],\"name\":\"FinalConcluded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"}],\"name\":\"Progressed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"}],\"name\":\"Refuted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"}],\"name\":\"Registered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"timeout\",\"type\":\"uint64\"}],\"name\":\"Stored\",\"type\":\"event\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"challengeDuration\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"app\",\"type\":\"address\"},{\"internalType\":\"address[]\",\"name\":\"participants\",\"type\":\"address[]\"}],\"internalType\":\"structChannel.Params\",\"name\":\"params\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"},{\"components\":[{\"internalType\":\"address[]\",\"name\":\"assets\",\"type\":\"address[]\"},{\"internalType\":\"uint256[][]\",\"name\":\"balances\",\"type\":\"uint256[][]\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"ID\",\"type\":\"bytes32\"},{\"internalType\":\"uint256[]\",\"name\":\"balances\",\"type\":\"uint256[]\"}],\"internalType\":\"structChannel.SubAlloc[]\",\"name\":\"locked\",\"type\":\"tuple[]\"}],\"internalType\":\"structChannel.Allocation\",\"name\":\"outcome\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"appData\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"isFinal\",\"type\":\"bool\"}],\"internalType\":\"structChannel.State\",\"name\":\"state\",\"type\":\"tuple\"}],\"name\":\"conclude\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"challengeDuration\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"app\",\"type\":\"address\"},{\"internalType\":\"address[]\",\"name\":\"participants\",\"type\":\"address[]\"}],\"internalType\":\"structChannel.Params\",\"name\":\"params\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"},{\"components\":[{\"internalType\":\"address[]\",\"name\":\"assets\",\"type\":\"address[]\"},{\"internalType\":\"uint256[][]\",\"name\":\"balances\",\"type\":\"uint256[][]\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"ID\",\"type\":\"bytes32\"},{\"internalType\":\"uint256[]\",\"name\":\"balances\",\"type\":\"uint256[]\"}],\"internalType\":\"structChannel.SubAlloc[]\",\"name\":\"locked\",\"type\":\"tuple[]\"}],\"internalType\":\"structChannel.Allocation\",\"name\":\"outcome\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"appData\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"isFinal\",\"type\":\"bool\"}],\"internalType\":\"structChannel.State\",\"name\":\"state\",\"type\":\"tuple\"},{\"internalType\":\"bytes[]\",\"name\":\"sigs\",\"type\":\"bytes[]\"}],\"name\":\"concludeFinal\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"disputes\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"timeout\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"},{\"internalType\":\"uint8\",\"name\":\"disputePhase\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"stateHash\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"challengeDuration\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"app\",\"type\":\"address\"},{\"internalType\":\"address[]\",\"name\":\"participants\",\"type\":\"address[]\"}],\"internalType\":\"structChannel.Params\",\"name\":\"params\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"},{\"components\":[{\"internalType\":\"address[]\",\"name\":\"assets\",\"type\":\"address[]\"},{\"internalType\":\"uint256[][]\",\"name\":\"balances\",\"type\":\"uint256[][]\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"ID\",\"type\":\"bytes32\"},{\"internalType\":\"uint256[]\",\"name\":\"balances\",\"type\":\"uint256[]\"}],\"internalType\":\"structChannel.SubAlloc[]\",\"name\":\"locked\",\"type\":\"tuple[]\"}],\"internalType\":\"structChannel.Allocation\",\"name\":\"outcome\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"appData\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"isFinal\",\"type\":\"bool\"}],\"internalType\":\"structChannel.State\",\"name\":\"stateOld\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"},{\"components\":[{\"internalType\":\"address[]\",\"name\":\"assets\",\"type\":\"address[]\"},{\"internalType\":\"uint256[][]\",\"name\":\"balances\",\"type\":\"uint256[][]\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"ID\",\"type\":\"bytes32\"},{\"internalType\":\"uint256[]\",\"name\":\"balances\",\"type\":\"uint256[]\"}],\"internalType\":\"structChannel.SubAlloc[]\",\"name\":\"locked\",\"type\":\"tuple[]\"}],\"internalType\":\"structChannel.Allocation\",\"name\":\"outcome\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"appData\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"isFinal\",\"type\":\"bool\"}],\"internalType\":\"structChannel.State\",\"name\":\"state\",\"type\":\"tuple\"},{\"internalType\":\"uint256\",\"name\":\"actorIdx\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"sig\",\"type\":\"bytes\"}],\"name\":\"progress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"challengeDuration\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"app\",\"type\":\"address\"},{\"internalType\":\"address[]\",\"name\":\"participants\",\"type\":\"address[]\"}],\"internalType\":\"structChannel.Params\",\"name\":\"params\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"},{\"components\":[{\"internalType\":\"address[]\",\"name\":\"assets\",\"type\":\"address[]\"},{\"internalType\":\"uint256[][]\",\"name\":\"balances\",\"type\":\"uint256[][]\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"ID\",\"type\":\"bytes32\"},{\"internalType\":\"uint256[]\",\"name\":\"balances\",\"type\":\"uint256[]\"}],\"internalType\":\"structChannel.SubAlloc[]\",\"name\":\"locked\",\"type\":\"tuple[]\"}],\"internalType\":\"structChannel.Allocation\",\"name\":\"outcome\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"appData\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"isFinal\",\"type\":\"bool\"}],\"internalType\":\"structChannel.State\",\"name\":\"state\",\"type\":\"tuple\"},{\"internalType\":\"bytes[]\",\"name\":\"sigs\",\"type\":\"bytes[]\"}],\"name\":\"refute\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"challengeDuration\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"app\",\"type\":\"address\"},{\"internalType\":\"address[]\",\"name\":\"participants\",\"type\":\"address[]\"}],\"internalType\":\"structChannel.Params\",\"name\":\"params\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"},{\"components\":[{\"internalType\":\"address[]\",\"name\":\"assets\",\"type\":\"address[]\"},{\"internalType\":\"uint256[][]\",\"name\":\"balances\",\"type\":\"uint256[][]\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"ID\",\"type\":\"bytes32\"},{\"internalType\":\"uint256[]\",\"name\":\"balances\",\"type\":\"uint256[]\"}],\"internalType\":\"structChannel.SubAlloc[]\",\"name\":\"locked\",\"type\":\"tuple[]\"}],\"internalType\":\"structChannel.Allocation\",\"name\":\"outcome\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"appData\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"isFinal\",\"type\":\"bool\"}],\"internalType\":\"structChannel.State\",\"name\":\"state\",\"type\":\"tuple\"},{\"internalType\":\"bytes[]\",\"name\":\"sigs\",\"type\":\"bytes[]\"}],\"name\":\"register\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// AdjudicatorFuncSigs maps the 4-byte function signature to its string representation.
var AdjudicatorFuncSigs = map[string]string{
	"637a1e9d": "conclude((uint256,uint256,address,address[]),(bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool))",
	"6bbf706a": "concludeFinal((uint256,uint256,address,address[]),(bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool),bytes[])",
	"11be1997": "disputes(bytes32)",
	"36995831": "progress((uint256,uint256,address,address[]),(bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool),(bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool),uint256,bytes)",
	"9b24e091": "refute((uint256,uint256,address,address[]),(bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool),bytes[])",
	"170e6715": "register((uint256,uint256,address,address[]),(bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool),bytes[])",
}

// AdjudicatorBin is the compiled bytecode used for deploying new contracts.
var AdjudicatorBin = "0x608060405234801561001057600080fd5b50612148806100206000396000f3fe608060405234801561001057600080fd5b50600436106100625760003560e01c806311be199714610067578063170e67151461009357806336995831146100a8578063637a1e9d146100bb5780636bbf706a146100ce5780639b24e091146100e1575b600080fd5b61007a6100753660046113f9565b6100f4565b60405161008a9493929190612098565b60405180910390f35b6100a66100a1366004611471565b610130565b005b6100a66100b6366004611543565b6101f0565b6100a66100c9366004611411565b6103ce565b6100a66100dc366004611471565b61047a565b6100a66100ef366004611471565b610515565b600060208190529081526040902080546001909101546001600160401b0380831692600160401b810490911691600160801b90910460ff169084565b600061013b84610642565b835190915081146101675760405162461bcd60e51b815260040161015e90611e90565b60405180910390fd5b600081815260208190526040902060010154156101965760405162461bcd60e51b815260040161015e90611a9b565b6101a1848484610688565b6101ae8484836000610719565b807f614ec80aa8693b5d1c65e212c9150ed89cadbd09945b4c9dc8356ab775feea2c84602001516040516101e2919061206a565b60405180910390a250505050565b84606001515182106102145760405162461bcd60e51b815260040161015e90611c49565b600061021f86610642565b600081815260208190526040902054909150600160801b900460ff1661027d57600081815260208190526040902054426001600160401b0390911611156102785760405162461bcd60e51b815260040161015e90611df4565b6102b6565b60008181526020819052604090205460ff600160801b909104166001146102b65760405162461bcd60e51b815260040161015e90611f2d565b835181146102d65760405162461bcd60e51b815260040161015e90611e90565b846040516020016102e79190612023565b60408051601f1981840301815291815281516020928301206000848152928390529120600101541461032b5760405162461bcd60e51b815260040161015e90611b4f565b6103556103378561084c565b838860600151868151811061034857fe5b60200260200101516108ce565b6103715760405162461bcd60e51b815260040161015e90611ad2565b61037d86868686610909565b61038a8685836001610719565b807f7f204194f5cc64dfd0591b3a8ef245dc7c0f756d71a48344dd6d586d3f290ba985602001516040516103be919061206a565b60405180910390a2505050505050565b60006103d983610642565b600081815260208190526040902054909150426001600160401b0390911611156104155760405162461bcd60e51b815260040161015e90611df4565b816040516020016104269190612023565b60408051601f1981840301815291815281516020928301206000848152928390529120600101541461046a5760405162461bcd60e51b815260040161015e90611b4f565b6104758184846109ed565b505050565b608082015115156001146104a05760405162461bcd60e51b815260040161015e90611a72565b60006104ab84610642565b835190915081146104ce5760405162461bcd60e51b815260040161015e90611e90565b6104d9848484610688565b6104e48185856109ed565b60405181907fc66555504438611dd09cb33bebae3edcd139b3251aacc295dc7958eb55302a9590600090a250505050565b600061052084610642565b835190915081146105435760405162461bcd60e51b815260040161015e90611e90565b60008181526020818152604090912054908401516001600160401b03600160401b90920482169116116105885760405162461bcd60e51b815260040161015e90611b78565b600081815260208190526040902054426001600160401b03909116116105c05760405162461bcd60e51b815260040161015e906119b1565b600081815260208190526040902054600160801b900460ff16156105f65760405162461bcd60e51b815260040161015e90611e5b565b610601848484610688565b61060e8484836000610719565b807f501ceb562a76f0a7471e93a4cfe90c33ac65eeed378064384f3aa80c0c60d46484602001516040516101e2919061206a565b6000816000015182602001518360400151846060015160405160200161066b949392919061203d565b604051602081830303815290604052805190602001209050919050565b60606106938361084c565b90508151846060015151146106ba5760405162461bcd60e51b815260040161015e90611dbd565b60005b8251811015610712576106ee828483815181106106d657fe5b60200260200101518760600151848151811061034857fe5b61070a5760405162461bcd60e51b815260040161015e90611a10565b6001016106bd565b5050505050565b6040808401510151511561073f5760405162461bcd60e51b815260040161015e90611f00565b835160009061074f904290610a92565b9050836040516020016107629190612023565b60408051808303601f19018152918152815160209283012060008681529283905291206001810191909155805467ffffffffffffffff19166001600160401b0383161790558160028111156107b357fe5b6000848152602081815260409182902080549188015160ff60801b19909216600160801b60ff9590951694909402939093176fffffffffffffffff00000000000000001916600160401b6001600160401b03831602179092555184917ffc420848ac2a3b00435a990dc8d2a2b782b7ef4a2f700cdb3d3c0c821c477b129161083d9190859061207e565b60405180910390a25050505050565b60408051602080820183526000825283830151805190820151935160609485936108799392869101611852565b604051602081830303815290604052905060608460000151856020015183876060015188608001516040516020016108b595949392919061190d565b60408051808303601f1901815291905295945050505050565b6000806108e18580519060200120610ac0565b905060006108ef8286610ad3565b6001600160a01b0390811690851614925050509392505050565b82602001516001016001600160401b031682602001516001600160401b0316146109455760405162461bcd60e51b815260040161015e90611ebb565b6080830151156109675760405162461bcd60e51b815260040161015e90611bad565b61097f83604001518360400151866060015151610c01565b6040808501519051637614eebf60e11b81526001600160a01b0382169063ec29dd7e906109b6908890889088908890600401611fa6565b60006040518083038186803b1580156109ce57600080fd5b505afa1580156109e2573d6000803e3d6000fd5b505050505050505050565b60008381526020819052604090205460ff600160801b9091041660021415610a275760405162461bcd60e51b815260040161015e90611c78565b6000838152602081905260409020805460ff60801b1916600160811b179055610a51838383610e2f565b827f9c10c93b4ed18ef3d175b09b84e7a8a0167ed67174f985922ae3683130c416278260200151604051610a85919061206a565b60405180910390a2505050565b600082820183811015610ab75760405162461bcd60e51b815260040161015e90611b18565b90505b92915050565b60008160405160200161066b9190611821565b60008151604114610af65760405162461bcd60e51b815260040161015e90611a3b565b60208201516040830151606084015160001a7f7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a0821115610b485760405162461bcd60e51b815260040161015e90611cf4565b8060ff16601b14158015610b6057508060ff16601c14155b15610b7d5760405162461bcd60e51b815260040161015e90611d7b565b600060018783868660405160008152602001604052604051610ba2949392919061195c565b6020604051602081039080840390855afa158015610bc4573d6000803e3d6000fd5b5050604051601f1901519150506001600160a01b038116610bf75760405162461bcd60e51b815260040161015e9061197a565b9695505050505050565b81602001515183602001515114610c2a5760405162461bcd60e51b815260040161015e90611f6f565b81515183515114610c4d5760405162461bcd60e51b815260040161015e90611be2565b60005b825151811015610e29578251805182908110610c6857fe5b60200260200101516001600160a01b031684600001518281518110610c8957fe5b60200260200101516001600160a01b031614610cb75760405162461bcd60e51b815260040161015e906119d9565b6000808386602001518481518110610ccb57fe5b60200260200101515114610cf15760405162461bcd60e51b815260040161015e90611d36565b8385602001518481518110610d0257fe5b60200260200101515114610d285760405162461bcd60e51b815260040161015e90611caf565b60005b84811015610db957610d7087602001518581518110610d4657fe5b60200260200101518281518110610d5957fe5b602002602001015184610a9290919063ffffffff16565b9250610daf86602001518581518110610d8557fe5b60200260200101518281518110610d9857fe5b602002602001015183610a9290919063ffffffff16565b9150600101610d2b565b5060408601515115610ddd5760405162461bcd60e51b815260040161015e90611f00565b60408501515115610e005760405162461bcd60e51b815260040161015e90611f00565b808214610e1f5760405162461bcd60e51b815260040161015e90611e24565b5050600101610c50565b50505050565b604081015151516060906001600160401b0381118015610e4e57600080fd5b50604051908082528060200260200182016040528015610e8257816020015b6060815260200190600190039081610e6d5790505b5090506060826040015160400151516001600160401b0381118015610ea657600080fd5b50604051908082528060200260200182016040528015610ed0578160200160208202803683370190505b50905060005b60408401515151811015610fe95760008460400151600001518281518110610efa57fe5b602002602001015190508560600151518560400151602001518381518110610f1e57fe5b60200260200101515114610f445760405162461bcd60e51b815260040161015e90611c12565b806001600160a01b03166379aad62e8888606001518860400151602001518681518110610f6d57fe5b602002602001015187898881518110610f8257fe5b60200260200101516040518663ffffffff1660e01b8152600401610faa95949392919061188b565b600060405180830381600087803b158015610fc457600080fd5b505af1158015610fd8573d6000803e3d6000fd5b505060019093019250610ed6915050565b505050505050565b600082601f830112611001578081fd5b813561101461100f826120ea565b6120c4565b81815291506020808301908481018184028601820187101561103557600080fd5b6000805b858110156110695782356001600160a01b0381168114611057578283fd5b85529383019391830191600101611039565b50505050505092915050565b600082601f830112611085578081fd5b813561109361100f826120ea565b818152915060208083019084810160005b848110156111135781358701604080601f19838c030112156110c557600080fd5b6110ce816120c4565b82860135815290820135906001600160401b038211156110ed57600080fd5b6110fb8b878486010161111e565b818701528652505092820192908201906001016110a4565b505050505092915050565b600082601f83011261112e578081fd5b813561113c61100f826120ea565b81815291506020808301908481018184028601820187101561115d57600080fd5b60005b8481101561111357813584529282019290820190600101611160565b80358015158114610aba57600080fd5b600082601f83011261119c578081fd5b81356001600160401b038111156111b1578182fd5b6111c4601f8201601f19166020016120c4565b91508082528360208285010111156111db57600080fd5b8060208401602084013760009082016020015292915050565b600060608284031215611205578081fd5b61120f60606120c4565b905081356001600160401b038082111561122857600080fd5b61123485838601610ff1565b835260209150818401358181111561124b57600080fd5b8401601f8101861361125c57600080fd5b803561126a61100f826120ea565b8181528481019083860160005b848110156112a05761128e8b89843589010161111e565b84529287019290870190600101611277565b505080868801525050505060408401359150808211156112bf57600080fd5b506112cc84828501611075565b60408301525092915050565b6000608082840312156112e9578081fd5b6112f360806120c4565b8235815260208084013590820152905060408201356001600160a01b038116811461131d57600080fd5b604082015260608201356001600160401b0381111561133b57600080fd5b61134784828501610ff1565b60608301525092915050565b600060a08284031215611364578081fd5b61136e60a06120c4565b90508135815260208201356001600160401b03808216821461138f57600080fd5b81602084015260408401359150808211156113a957600080fd5b6113b5858386016111f4565b604084015260608401359150808211156113ce57600080fd5b506113db8482850161118c565b6060830152506113ee836080840161117c565b608082015292915050565b60006020828403121561140a578081fd5b5035919050565b60008060408385031215611423578081fd5b82356001600160401b0380821115611439578283fd5b611445868387016112d8565b9350602085013591508082111561145a578283fd5b5061146785828601611353565b9150509250929050565b600080600060608486031215611485578081fd5b83356001600160401b038082111561149b578283fd5b6114a7878388016112d8565b94506020915081860135818111156114bd578384fd5b6114c988828901611353565b9450506040860135818111156114dd578384fd5b86019050601f810187136114ef578283fd5b80356114fd61100f826120ea565b81815283810190838501865b84811015611532576115208c88843589010161118c565b84529286019290860190600101611509565b505080955050505050509250925092565b600080600080600060a0868803121561155a578081fd5b85356001600160401b0380821115611570578283fd5b61157c89838a016112d8565b96506020880135915080821115611591578283fd5b61159d89838a01611353565b955060408801359150808211156115b2578283fd5b6115be89838a01611353565b94506060880135935060808801359150808211156115da578283fd5b506115e78882890161118c565b9150509295509295909350565b60008151835260208201516040602085015261161360408501826116b1565b949350505050565b6000815180845260208085019450808401835b838110156116535781516001600160a01b03168752958201959082019060010161162e565b509495945050505050565b6000815180845260208085018081965082840281019150828601855b858110156116a45782840389526116928483516116b1565b9885019893509084019060010161167a565b5091979650505050505050565b6000815180845260208085019450808401835b83811015611653578151875295820195908201906001016116c4565b15159052565b60008151808452815b8181101561170b576020818501810151868301820152016116ef565b8181111561171c5782602083870101525b50601f01601f19169290920160200192915050565b60008151835260206001600160401b03818401511681850152604083015160a060408601528051606060a087015261176d61010087018261161b565b905082820151609f19808884030160c089015261178a838361165e565b925060408401519350808884030160e089015250508082516117ac8184612109565b9385019392508291508085028201865b828110156117e65784820384526117d48287516115f4565b958701959387019391506001016117bc565b506060880151955088810360608a015261180081876116e6565b95505050505050608083015161181960808601826116e0565b509392505050565b7f19457468657265756d205369676e6564204d6573736167653a0a3332000000008152601c810191909152603c0190565b600060608252611865606083018661161b565b8281036020840152611877818661165e565b90508281036040840152610bf781856116e6565b6000868252602060a0818401526118a560a084018861161b565b83810360408501526118b781886116b1565b84810360608601528651808252838801925090830190845b818110156118eb578351835292840192918401916001016118cf565b505084810360808601526118ff81876116b1565b9a9950505050505050505050565b60008682526001600160401b038616602083015260a0604083015261193560a08301866116e6565b828103606084015261194781866116e6565b91505082151560808301529695505050505050565b93845260ff9290921660208401526040830152606082015260800190565b60208082526018908201527f45434453413a20696e76616c6964207369676e61747572650000000000000000604082015260600190565b6020808252600e908201526d1d1a5b595bdd5d081c185cdcd95960921b604082015260600190565b6020808252601a908201527f6173736574735b695d2061646472657373206d69736d61746368000000000000604082015260600190565b602080825260119082015270696e76616c6964207369676e617475726560781b604082015260600190565b6020808252601f908201527f45434453413a20696e76616c6964207369676e6174757265206c656e67746800604082015260600190565b6020808252600f908201526e1cdd185d19481b9bdd08199a5b985b608a1b604082015260600190565b60208082526018908201527f737461746520616c726561647920726567697374657265640000000000000000604082015260600190565b60208082526026908201527f6163746f7249647820646f6573206e6f74206d61746368207369676e65722773604082015265040d2dcc8caf60d31b606082015260800190565b6020808252601b908201527f536166654d6174683a206164646974696f6e206f766572666c6f770000000000604082015260600190565b6020808252600f908201526e77726f6e67206f6c6420737461746560881b604082015260600190565b6020808252818101527f63616e206f6e6c79207265667574652077697468206e65776572207374617465604082015260600190565b6020808252818101527f63616e6e6f742070726f67726573732066726f6d2066696e616c207374617465604082015260600190565b6020808252601690820152750c2e6e6cae8e640d8cadccee8d040dad2e6dac2e8c6d60531b604082015260600190565b6020808252601b908201527f62616c616e6365735b695d206c656e677468206d69736d617463680000000000604082015260600190565b6020808252601590820152746163746f72496478206f7574206f662072616e676560581b604082015260600190565b60208082526019908201527f6368616e6e656c20616c726561647920636f6e636c7564656400000000000000604082015260600190565b60208082526025908201527f6e6577416c6c6f633a2062616c616e6365735b695d206c656e677468206d69736040820152640dac2e8c6d60db1b606082015260800190565b60208082526022908201527f45434453413a20696e76616c6964207369676e6174757265202773272076616c604082015261756560f01b606082015260800190565b60208082526025908201527f6f6c64416c6c6f633a2062616c616e6365735b695d206c656e677468206d69736040820152640dac2e8c6d60db1b606082015260800190565b60208082526022908201527f45434453413a20696e76616c6964207369676e6174757265202776272076616c604082015261756560f01b606082015260800190565b6020808252601a908201527f7369676e617475726573206c656e677468206d69736d61746368000000000000604082015260600190565b6020808252601690820152751d1a5b595bdd5d081b9bdd081c185cdcd959081e595d60521b604082015260600190565b60208082526018908201527f73756d206f662062616c616e636573206d69736d617463680000000000000000604082015260600190565b6020808252818101527f6368616e6e656c206d75737420626520696e2044495350555445207068617365604082015260600190565b6020808252601190820152701a5b9d985b1a590818da185b9b995b1251607a1b604082015260600190565b60208082526025908201527f76657273696f6e20636f756e746572206d75737420696e6372656d656e74206260408201526479206f6e6560d81b606082015260800190565b6020808252601390820152721b9bdd081a5b5c1b195b595b9d1959081e595d606a1b604082015260600190565b60208082526022908201527f6368616e6e656c206d75737420626520696e20464f5243454558454320706861604082015261736560f01b606082015260800190565b60208082526018908201527f62616c616e636573206c656e677468206d69736d617463680000000000000000604082015260600190565b60006080825285516080830152602086015160a083015260018060a01b0360408701511660c08301526060860151608060e0840152611fe961010084018261161b565b90508281036020840152611ffd8187611731565b905082810360408401526120118186611731565b91505082606083015295945050505050565b6000602082526120366020830184611731565b9392505050565b600085825284602083015260018060a01b038416604083015260806060830152610bf7608083018461161b565b6001600160401b0391909116815260200190565b6001600160401b0392831681529116602082015260400190565b6001600160401b03948516815292909316602083015260ff166040820152606081019190915260800190565b6040518181016001600160401b03811182821017156120e257600080fd5b604052919050565b60006001600160401b038211156120ff578081fd5b5060209081020190565b9081526020019056fea26469706673582212207b1f111d9a27aefdcd4d371d9e375efe884f9934db3dffb4f4e73b81b4274f5a64736f6c63430007000033"

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

// Disputes is a free data retrieval call binding the contract method 0x11be1997.
//
// Solidity: function disputes(bytes32 ) view returns(uint64 timeout, uint64 version, uint8 disputePhase, bytes32 stateHash)
func (_Adjudicator *AdjudicatorCaller) Disputes(opts *bind.CallOpts, arg0 [32]byte) (struct {
	Timeout      uint64
	Version      uint64
	DisputePhase uint8
	StateHash    [32]byte
}, error) {
	ret := new(struct {
		Timeout      uint64
		Version      uint64
		DisputePhase uint8
		StateHash    [32]byte
	})
	out := ret
	err := _Adjudicator.contract.Call(opts, out, "disputes", arg0)
	return *ret, err
}

// Disputes is a free data retrieval call binding the contract method 0x11be1997.
//
// Solidity: function disputes(bytes32 ) view returns(uint64 timeout, uint64 version, uint8 disputePhase, bytes32 stateHash)
func (_Adjudicator *AdjudicatorSession) Disputes(arg0 [32]byte) (struct {
	Timeout      uint64
	Version      uint64
	DisputePhase uint8
	StateHash    [32]byte
}, error) {
	return _Adjudicator.Contract.Disputes(&_Adjudicator.CallOpts, arg0)
}

// Disputes is a free data retrieval call binding the contract method 0x11be1997.
//
// Solidity: function disputes(bytes32 ) view returns(uint64 timeout, uint64 version, uint8 disputePhase, bytes32 stateHash)
func (_Adjudicator *AdjudicatorCallerSession) Disputes(arg0 [32]byte) (struct {
	Timeout      uint64
	Version      uint64
	DisputePhase uint8
	StateHash    [32]byte
}, error) {
	return _Adjudicator.Contract.Disputes(&_Adjudicator.CallOpts, arg0)
}

// Conclude is a paid mutator transaction binding the contract method 0x637a1e9d.
//
// Solidity: function conclude((uint256,uint256,address,address[]) params, (bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool) state) returns()
func (_Adjudicator *AdjudicatorTransactor) Conclude(opts *bind.TransactOpts, params ChannelParams, state ChannelState) (*types.Transaction, error) {
	return _Adjudicator.contract.Transact(opts, "conclude", params, state)
}

// Conclude is a paid mutator transaction binding the contract method 0x637a1e9d.
//
// Solidity: function conclude((uint256,uint256,address,address[]) params, (bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool) state) returns()
func (_Adjudicator *AdjudicatorSession) Conclude(params ChannelParams, state ChannelState) (*types.Transaction, error) {
	return _Adjudicator.Contract.Conclude(&_Adjudicator.TransactOpts, params, state)
}

// Conclude is a paid mutator transaction binding the contract method 0x637a1e9d.
//
// Solidity: function conclude((uint256,uint256,address,address[]) params, (bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool) state) returns()
func (_Adjudicator *AdjudicatorTransactorSession) Conclude(params ChannelParams, state ChannelState) (*types.Transaction, error) {
	return _Adjudicator.Contract.Conclude(&_Adjudicator.TransactOpts, params, state)
}

// ConcludeFinal is a paid mutator transaction binding the contract method 0x6bbf706a.
//
// Solidity: function concludeFinal((uint256,uint256,address,address[]) params, (bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool) state, bytes[] sigs) returns()
func (_Adjudicator *AdjudicatorTransactor) ConcludeFinal(opts *bind.TransactOpts, params ChannelParams, state ChannelState, sigs [][]byte) (*types.Transaction, error) {
	return _Adjudicator.contract.Transact(opts, "concludeFinal", params, state, sigs)
}

// ConcludeFinal is a paid mutator transaction binding the contract method 0x6bbf706a.
//
// Solidity: function concludeFinal((uint256,uint256,address,address[]) params, (bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool) state, bytes[] sigs) returns()
func (_Adjudicator *AdjudicatorSession) ConcludeFinal(params ChannelParams, state ChannelState, sigs [][]byte) (*types.Transaction, error) {
	return _Adjudicator.Contract.ConcludeFinal(&_Adjudicator.TransactOpts, params, state, sigs)
}

// ConcludeFinal is a paid mutator transaction binding the contract method 0x6bbf706a.
//
// Solidity: function concludeFinal((uint256,uint256,address,address[]) params, (bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool) state, bytes[] sigs) returns()
func (_Adjudicator *AdjudicatorTransactorSession) ConcludeFinal(params ChannelParams, state ChannelState, sigs [][]byte) (*types.Transaction, error) {
	return _Adjudicator.Contract.ConcludeFinal(&_Adjudicator.TransactOpts, params, state, sigs)
}

// Progress is a paid mutator transaction binding the contract method 0x36995831.
//
// Solidity: function progress((uint256,uint256,address,address[]) params, (bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool) stateOld, (bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool) state, uint256 actorIdx, bytes sig) returns()
func (_Adjudicator *AdjudicatorTransactor) Progress(opts *bind.TransactOpts, params ChannelParams, stateOld ChannelState, state ChannelState, actorIdx *big.Int, sig []byte) (*types.Transaction, error) {
	return _Adjudicator.contract.Transact(opts, "progress", params, stateOld, state, actorIdx, sig)
}

// Progress is a paid mutator transaction binding the contract method 0x36995831.
//
// Solidity: function progress((uint256,uint256,address,address[]) params, (bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool) stateOld, (bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool) state, uint256 actorIdx, bytes sig) returns()
func (_Adjudicator *AdjudicatorSession) Progress(params ChannelParams, stateOld ChannelState, state ChannelState, actorIdx *big.Int, sig []byte) (*types.Transaction, error) {
	return _Adjudicator.Contract.Progress(&_Adjudicator.TransactOpts, params, stateOld, state, actorIdx, sig)
}

// Progress is a paid mutator transaction binding the contract method 0x36995831.
//
// Solidity: function progress((uint256,uint256,address,address[]) params, (bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool) stateOld, (bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool) state, uint256 actorIdx, bytes sig) returns()
func (_Adjudicator *AdjudicatorTransactorSession) Progress(params ChannelParams, stateOld ChannelState, state ChannelState, actorIdx *big.Int, sig []byte) (*types.Transaction, error) {
	return _Adjudicator.Contract.Progress(&_Adjudicator.TransactOpts, params, stateOld, state, actorIdx, sig)
}

// Refute is a paid mutator transaction binding the contract method 0x9b24e091.
//
// Solidity: function refute((uint256,uint256,address,address[]) params, (bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool) state, bytes[] sigs) returns()
func (_Adjudicator *AdjudicatorTransactor) Refute(opts *bind.TransactOpts, params ChannelParams, state ChannelState, sigs [][]byte) (*types.Transaction, error) {
	return _Adjudicator.contract.Transact(opts, "refute", params, state, sigs)
}

// Refute is a paid mutator transaction binding the contract method 0x9b24e091.
//
// Solidity: function refute((uint256,uint256,address,address[]) params, (bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool) state, bytes[] sigs) returns()
func (_Adjudicator *AdjudicatorSession) Refute(params ChannelParams, state ChannelState, sigs [][]byte) (*types.Transaction, error) {
	return _Adjudicator.Contract.Refute(&_Adjudicator.TransactOpts, params, state, sigs)
}

// Refute is a paid mutator transaction binding the contract method 0x9b24e091.
//
// Solidity: function refute((uint256,uint256,address,address[]) params, (bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool) state, bytes[] sigs) returns()
func (_Adjudicator *AdjudicatorTransactorSession) Refute(params ChannelParams, state ChannelState, sigs [][]byte) (*types.Transaction, error) {
	return _Adjudicator.Contract.Refute(&_Adjudicator.TransactOpts, params, state, sigs)
}

// Register is a paid mutator transaction binding the contract method 0x170e6715.
//
// Solidity: function register((uint256,uint256,address,address[]) params, (bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool) state, bytes[] sigs) returns()
func (_Adjudicator *AdjudicatorTransactor) Register(opts *bind.TransactOpts, params ChannelParams, state ChannelState, sigs [][]byte) (*types.Transaction, error) {
	return _Adjudicator.contract.Transact(opts, "register", params, state, sigs)
}

// Register is a paid mutator transaction binding the contract method 0x170e6715.
//
// Solidity: function register((uint256,uint256,address,address[]) params, (bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool) state, bytes[] sigs) returns()
func (_Adjudicator *AdjudicatorSession) Register(params ChannelParams, state ChannelState, sigs [][]byte) (*types.Transaction, error) {
	return _Adjudicator.Contract.Register(&_Adjudicator.TransactOpts, params, state, sigs)
}

// Register is a paid mutator transaction binding the contract method 0x170e6715.
//
// Solidity: function register((uint256,uint256,address,address[]) params, (bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool) state, bytes[] sigs) returns()
func (_Adjudicator *AdjudicatorTransactorSession) Register(params ChannelParams, state ChannelState, sigs [][]byte) (*types.Transaction, error) {
	return _Adjudicator.Contract.Register(&_Adjudicator.TransactOpts, params, state, sigs)
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
	Version   uint64
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterConcluded is a free log retrieval operation binding the contract event 0x9c10c93b4ed18ef3d175b09b84e7a8a0167ed67174f985922ae3683130c41627.
//
// Solidity: event Concluded(bytes32 indexed channelID, uint64 version)
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

// WatchConcluded is a free log subscription operation binding the contract event 0x9c10c93b4ed18ef3d175b09b84e7a8a0167ed67174f985922ae3683130c41627.
//
// Solidity: event Concluded(bytes32 indexed channelID, uint64 version)
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

// ParseConcluded is a log parse operation binding the contract event 0x9c10c93b4ed18ef3d175b09b84e7a8a0167ed67174f985922ae3683130c41627.
//
// Solidity: event Concluded(bytes32 indexed channelID, uint64 version)
func (_Adjudicator *AdjudicatorFilterer) ParseConcluded(log types.Log) (*AdjudicatorConcluded, error) {
	event := new(AdjudicatorConcluded)
	if err := _Adjudicator.contract.UnpackLog(event, "Concluded", log); err != nil {
		return nil, err
	}
	return event, nil
}

// AdjudicatorFinalConcludedIterator is returned from FilterFinalConcluded and is used to iterate over the raw logs and unpacked data for FinalConcluded events raised by the Adjudicator contract.
type AdjudicatorFinalConcludedIterator struct {
	Event *AdjudicatorFinalConcluded // Event containing the contract specifics and raw log

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
func (it *AdjudicatorFinalConcludedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AdjudicatorFinalConcluded)
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
		it.Event = new(AdjudicatorFinalConcluded)
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
func (it *AdjudicatorFinalConcludedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AdjudicatorFinalConcludedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AdjudicatorFinalConcluded represents a FinalConcluded event raised by the Adjudicator contract.
type AdjudicatorFinalConcluded struct {
	ChannelID [32]byte
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterFinalConcluded is a free log retrieval operation binding the contract event 0xc66555504438611dd09cb33bebae3edcd139b3251aacc295dc7958eb55302a95.
//
// Solidity: event FinalConcluded(bytes32 indexed channelID)
func (_Adjudicator *AdjudicatorFilterer) FilterFinalConcluded(opts *bind.FilterOpts, channelID [][32]byte) (*AdjudicatorFinalConcludedIterator, error) {

	var channelIDRule []interface{}
	for _, channelIDItem := range channelID {
		channelIDRule = append(channelIDRule, channelIDItem)
	}

	logs, sub, err := _Adjudicator.contract.FilterLogs(opts, "FinalConcluded", channelIDRule)
	if err != nil {
		return nil, err
	}
	return &AdjudicatorFinalConcludedIterator{contract: _Adjudicator.contract, event: "FinalConcluded", logs: logs, sub: sub}, nil
}

// WatchFinalConcluded is a free log subscription operation binding the contract event 0xc66555504438611dd09cb33bebae3edcd139b3251aacc295dc7958eb55302a95.
//
// Solidity: event FinalConcluded(bytes32 indexed channelID)
func (_Adjudicator *AdjudicatorFilterer) WatchFinalConcluded(opts *bind.WatchOpts, sink chan<- *AdjudicatorFinalConcluded, channelID [][32]byte) (event.Subscription, error) {

	var channelIDRule []interface{}
	for _, channelIDItem := range channelID {
		channelIDRule = append(channelIDRule, channelIDItem)
	}

	logs, sub, err := _Adjudicator.contract.WatchLogs(opts, "FinalConcluded", channelIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AdjudicatorFinalConcluded)
				if err := _Adjudicator.contract.UnpackLog(event, "FinalConcluded", log); err != nil {
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

// ParseFinalConcluded is a log parse operation binding the contract event 0xc66555504438611dd09cb33bebae3edcd139b3251aacc295dc7958eb55302a95.
//
// Solidity: event FinalConcluded(bytes32 indexed channelID)
func (_Adjudicator *AdjudicatorFilterer) ParseFinalConcluded(log types.Log) (*AdjudicatorFinalConcluded, error) {
	event := new(AdjudicatorFinalConcluded)
	if err := _Adjudicator.contract.UnpackLog(event, "FinalConcluded", log); err != nil {
		return nil, err
	}
	return event, nil
}

// AdjudicatorProgressedIterator is returned from FilterProgressed and is used to iterate over the raw logs and unpacked data for Progressed events raised by the Adjudicator contract.
type AdjudicatorProgressedIterator struct {
	Event *AdjudicatorProgressed // Event containing the contract specifics and raw log

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
func (it *AdjudicatorProgressedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AdjudicatorProgressed)
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
		it.Event = new(AdjudicatorProgressed)
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
func (it *AdjudicatorProgressedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AdjudicatorProgressedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AdjudicatorProgressed represents a Progressed event raised by the Adjudicator contract.
type AdjudicatorProgressed struct {
	ChannelID [32]byte
	Version   uint64
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterProgressed is a free log retrieval operation binding the contract event 0x7f204194f5cc64dfd0591b3a8ef245dc7c0f756d71a48344dd6d586d3f290ba9.
//
// Solidity: event Progressed(bytes32 indexed channelID, uint64 version)
func (_Adjudicator *AdjudicatorFilterer) FilterProgressed(opts *bind.FilterOpts, channelID [][32]byte) (*AdjudicatorProgressedIterator, error) {

	var channelIDRule []interface{}
	for _, channelIDItem := range channelID {
		channelIDRule = append(channelIDRule, channelIDItem)
	}

	logs, sub, err := _Adjudicator.contract.FilterLogs(opts, "Progressed", channelIDRule)
	if err != nil {
		return nil, err
	}
	return &AdjudicatorProgressedIterator{contract: _Adjudicator.contract, event: "Progressed", logs: logs, sub: sub}, nil
}

// WatchProgressed is a free log subscription operation binding the contract event 0x7f204194f5cc64dfd0591b3a8ef245dc7c0f756d71a48344dd6d586d3f290ba9.
//
// Solidity: event Progressed(bytes32 indexed channelID, uint64 version)
func (_Adjudicator *AdjudicatorFilterer) WatchProgressed(opts *bind.WatchOpts, sink chan<- *AdjudicatorProgressed, channelID [][32]byte) (event.Subscription, error) {

	var channelIDRule []interface{}
	for _, channelIDItem := range channelID {
		channelIDRule = append(channelIDRule, channelIDItem)
	}

	logs, sub, err := _Adjudicator.contract.WatchLogs(opts, "Progressed", channelIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AdjudicatorProgressed)
				if err := _Adjudicator.contract.UnpackLog(event, "Progressed", log); err != nil {
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

// ParseProgressed is a log parse operation binding the contract event 0x7f204194f5cc64dfd0591b3a8ef245dc7c0f756d71a48344dd6d586d3f290ba9.
//
// Solidity: event Progressed(bytes32 indexed channelID, uint64 version)
func (_Adjudicator *AdjudicatorFilterer) ParseProgressed(log types.Log) (*AdjudicatorProgressed, error) {
	event := new(AdjudicatorProgressed)
	if err := _Adjudicator.contract.UnpackLog(event, "Progressed", log); err != nil {
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
	Version   uint64
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterRefuted is a free log retrieval operation binding the contract event 0x501ceb562a76f0a7471e93a4cfe90c33ac65eeed378064384f3aa80c0c60d464.
//
// Solidity: event Refuted(bytes32 indexed channelID, uint64 version)
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

// WatchRefuted is a free log subscription operation binding the contract event 0x501ceb562a76f0a7471e93a4cfe90c33ac65eeed378064384f3aa80c0c60d464.
//
// Solidity: event Refuted(bytes32 indexed channelID, uint64 version)
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

// ParseRefuted is a log parse operation binding the contract event 0x501ceb562a76f0a7471e93a4cfe90c33ac65eeed378064384f3aa80c0c60d464.
//
// Solidity: event Refuted(bytes32 indexed channelID, uint64 version)
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
	Version   uint64
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterRegistered is a free log retrieval operation binding the contract event 0x614ec80aa8693b5d1c65e212c9150ed89cadbd09945b4c9dc8356ab775feea2c.
//
// Solidity: event Registered(bytes32 indexed channelID, uint64 version)
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

// WatchRegistered is a free log subscription operation binding the contract event 0x614ec80aa8693b5d1c65e212c9150ed89cadbd09945b4c9dc8356ab775feea2c.
//
// Solidity: event Registered(bytes32 indexed channelID, uint64 version)
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

// ParseRegistered is a log parse operation binding the contract event 0x614ec80aa8693b5d1c65e212c9150ed89cadbd09945b4c9dc8356ab775feea2c.
//
// Solidity: event Registered(bytes32 indexed channelID, uint64 version)
func (_Adjudicator *AdjudicatorFilterer) ParseRegistered(log types.Log) (*AdjudicatorRegistered, error) {
	event := new(AdjudicatorRegistered)
	if err := _Adjudicator.contract.UnpackLog(event, "Registered", log); err != nil {
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
	Version   uint64
	Timeout   uint64
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterStored is a free log retrieval operation binding the contract event 0xfc420848ac2a3b00435a990dc8d2a2b782b7ef4a2f700cdb3d3c0c821c477b12.
//
// Solidity: event Stored(bytes32 indexed channelID, uint64 version, uint64 timeout)
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

// WatchStored is a free log subscription operation binding the contract event 0xfc420848ac2a3b00435a990dc8d2a2b782b7ef4a2f700cdb3d3c0c821c477b12.
//
// Solidity: event Stored(bytes32 indexed channelID, uint64 version, uint64 timeout)
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

// ParseStored is a log parse operation binding the contract event 0xfc420848ac2a3b00435a990dc8d2a2b782b7ef4a2f700cdb3d3c0c821c477b12.
//
// Solidity: event Stored(bytes32 indexed channelID, uint64 version, uint64 timeout)
func (_Adjudicator *AdjudicatorFilterer) ParseStored(log types.Log) (*AdjudicatorStored, error) {
	event := new(AdjudicatorStored)
	if err := _Adjudicator.contract.UnpackLog(event, "Stored", log); err != nil {
		return nil, err
	}
	return event, nil
}

// AppABI is the input ABI used to generate the binding from.
const AppABI = "[{\"inputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"challengeDuration\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"app\",\"type\":\"address\"},{\"internalType\":\"address[]\",\"name\":\"participants\",\"type\":\"address[]\"}],\"internalType\":\"structChannel.Params\",\"name\":\"params\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"},{\"components\":[{\"internalType\":\"address[]\",\"name\":\"assets\",\"type\":\"address[]\"},{\"internalType\":\"uint256[][]\",\"name\":\"balances\",\"type\":\"uint256[][]\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"ID\",\"type\":\"bytes32\"},{\"internalType\":\"uint256[]\",\"name\":\"balances\",\"type\":\"uint256[]\"}],\"internalType\":\"structChannel.SubAlloc[]\",\"name\":\"locked\",\"type\":\"tuple[]\"}],\"internalType\":\"structChannel.Allocation\",\"name\":\"outcome\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"appData\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"isFinal\",\"type\":\"bool\"}],\"internalType\":\"structChannel.State\",\"name\":\"from\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"},{\"components\":[{\"internalType\":\"address[]\",\"name\":\"assets\",\"type\":\"address[]\"},{\"internalType\":\"uint256[][]\",\"name\":\"balances\",\"type\":\"uint256[][]\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"ID\",\"type\":\"bytes32\"},{\"internalType\":\"uint256[]\",\"name\":\"balances\",\"type\":\"uint256[]\"}],\"internalType\":\"structChannel.SubAlloc[]\",\"name\":\"locked\",\"type\":\"tuple[]\"}],\"internalType\":\"structChannel.Allocation\",\"name\":\"outcome\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"appData\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"isFinal\",\"type\":\"bool\"}],\"internalType\":\"structChannel.State\",\"name\":\"to\",\"type\":\"tuple\"},{\"internalType\":\"uint256\",\"name\":\"actorIdx\",\"type\":\"uint256\"}],\"name\":\"validTransition\",\"outputs\":[],\"stateMutability\":\"pure\",\"type\":\"function\"}]"

// AppFuncSigs maps the 4-byte function signature to its string representation.
var AppFuncSigs = map[string]string{
	"ec29dd7e": "validTransition((uint256,uint256,address,address[]),(bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool),(bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool),uint256)",
}

// App is an auto generated Go binding around an Ethereum contract.
type App struct {
	AppCaller     // Read-only binding to the contract
	AppTransactor // Write-only binding to the contract
	AppFilterer   // Log filterer for contract events
}

// AppCaller is an auto generated read-only Go binding around an Ethereum contract.
type AppCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AppTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AppTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AppFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AppFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AppSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AppSession struct {
	Contract     *App              // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// AppCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AppCallerSession struct {
	Contract *AppCaller    // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// AppTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AppTransactorSession struct {
	Contract     *AppTransactor    // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// AppRaw is an auto generated low-level Go binding around an Ethereum contract.
type AppRaw struct {
	Contract *App // Generic contract binding to access the raw methods on
}

// AppCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AppCallerRaw struct {
	Contract *AppCaller // Generic read-only contract binding to access the raw methods on
}

// AppTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AppTransactorRaw struct {
	Contract *AppTransactor // Generic write-only contract binding to access the raw methods on
}

// NewApp creates a new instance of App, bound to a specific deployed contract.
func NewApp(address common.Address, backend bind.ContractBackend) (*App, error) {
	contract, err := bindApp(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &App{AppCaller: AppCaller{contract: contract}, AppTransactor: AppTransactor{contract: contract}, AppFilterer: AppFilterer{contract: contract}}, nil
}

// NewAppCaller creates a new read-only instance of App, bound to a specific deployed contract.
func NewAppCaller(address common.Address, caller bind.ContractCaller) (*AppCaller, error) {
	contract, err := bindApp(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AppCaller{contract: contract}, nil
}

// NewAppTransactor creates a new write-only instance of App, bound to a specific deployed contract.
func NewAppTransactor(address common.Address, transactor bind.ContractTransactor) (*AppTransactor, error) {
	contract, err := bindApp(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AppTransactor{contract: contract}, nil
}

// NewAppFilterer creates a new log filterer instance of App, bound to a specific deployed contract.
func NewAppFilterer(address common.Address, filterer bind.ContractFilterer) (*AppFilterer, error) {
	contract, err := bindApp(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AppFilterer{contract: contract}, nil
}

// bindApp binds a generic wrapper to an already deployed contract.
func bindApp(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(AppABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_App *AppRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _App.Contract.AppCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_App *AppRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _App.Contract.AppTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_App *AppRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _App.Contract.AppTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_App *AppCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _App.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_App *AppTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _App.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_App *AppTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _App.Contract.contract.Transact(opts, method, params...)
}

// ValidTransition is a free data retrieval call binding the contract method 0xec29dd7e.
//
// Solidity: function validTransition((uint256,uint256,address,address[]) params, (bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool) from, (bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool) to, uint256 actorIdx) pure returns()
func (_App *AppCaller) ValidTransition(opts *bind.CallOpts, params ChannelParams, from ChannelState, to ChannelState, actorIdx *big.Int) error {
	var ()
	out := &[]interface{}{}
	err := _App.contract.Call(opts, out, "validTransition", params, from, to, actorIdx)
	return err
}

// ValidTransition is a free data retrieval call binding the contract method 0xec29dd7e.
//
// Solidity: function validTransition((uint256,uint256,address,address[]) params, (bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool) from, (bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool) to, uint256 actorIdx) pure returns()
func (_App *AppSession) ValidTransition(params ChannelParams, from ChannelState, to ChannelState, actorIdx *big.Int) error {
	return _App.Contract.ValidTransition(&_App.CallOpts, params, from, to, actorIdx)
}

// ValidTransition is a free data retrieval call binding the contract method 0xec29dd7e.
//
// Solidity: function validTransition((uint256,uint256,address,address[]) params, (bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool) from, (bytes32,uint64,(address[],uint256[][],(bytes32,uint256[])[]),bytes,bool) to, uint256 actorIdx) pure returns()
func (_App *AppCallerSession) ValidTransition(params ChannelParams, from ChannelState, to ChannelState, actorIdx *big.Int) error {
	return _App.Contract.ValidTransition(&_App.CallOpts, params, from, to, actorIdx)
}

// AssetHolderABI is the input ABI used to generate the binding from.
const AssetHolderABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"fundingID\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Deposited\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"}],\"name\":\"OutcomeSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"fundingID\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"receiver\",\"type\":\"address\"}],\"name\":\"Withdrawn\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"adjudicator\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"fundingID\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"deposit\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"holdings\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"},{\"internalType\":\"address[]\",\"name\":\"parts\",\"type\":\"address[]\"},{\"internalType\":\"uint256[]\",\"name\":\"newBals\",\"type\":\"uint256[]\"},{\"internalType\":\"bytes32[]\",\"name\":\"subAllocs\",\"type\":\"bytes32[]\"},{\"internalType\":\"uint256[]\",\"name\":\"subBalances\",\"type\":\"uint256[]\"}],\"name\":\"setOutcome\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"settled\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"channelID\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"participant\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"receiver\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structAssetHolder.WithdrawalAuth\",\"name\":\"authorization\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"withdraw\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// AssetHolderFuncSigs maps the 4-byte function signature to its string representation.
var AssetHolderFuncSigs = map[string]string{
	"53c2ed8e": "adjudicator()",
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

// Adjudicator is a free data retrieval call binding the contract method 0x53c2ed8e.
//
// Solidity: function adjudicator() view returns(address)
func (_AssetHolder *AssetHolderCaller) Adjudicator(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _AssetHolder.contract.Call(opts, out, "adjudicator")
	return *ret0, err
}

// Adjudicator is a free data retrieval call binding the contract method 0x53c2ed8e.
//
// Solidity: function adjudicator() view returns(address)
func (_AssetHolder *AssetHolderSession) Adjudicator() (common.Address, error) {
	return _AssetHolder.Contract.Adjudicator(&_AssetHolder.CallOpts)
}

// Adjudicator is a free data retrieval call binding the contract method 0x53c2ed8e.
//
// Solidity: function adjudicator() view returns(address)
func (_AssetHolder *AssetHolderCallerSession) Adjudicator() (common.Address, error) {
	return _AssetHolder.Contract.Adjudicator(&_AssetHolder.CallOpts)
}

// Holdings is a free data retrieval call binding the contract method 0xae9ee18c.
//
// Solidity: function holdings(bytes32 ) view returns(uint256)
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
// Solidity: function holdings(bytes32 ) view returns(uint256)
func (_AssetHolder *AssetHolderSession) Holdings(arg0 [32]byte) (*big.Int, error) {
	return _AssetHolder.Contract.Holdings(&_AssetHolder.CallOpts, arg0)
}

// Holdings is a free data retrieval call binding the contract method 0xae9ee18c.
//
// Solidity: function holdings(bytes32 ) view returns(uint256)
func (_AssetHolder *AssetHolderCallerSession) Holdings(arg0 [32]byte) (*big.Int, error) {
	return _AssetHolder.Contract.Holdings(&_AssetHolder.CallOpts, arg0)
}

// Settled is a free data retrieval call binding the contract method 0xd945af1d.
//
// Solidity: function settled(bytes32 ) view returns(bool)
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
// Solidity: function settled(bytes32 ) view returns(bool)
func (_AssetHolder *AssetHolderSession) Settled(arg0 [32]byte) (bool, error) {
	return _AssetHolder.Contract.Settled(&_AssetHolder.CallOpts, arg0)
}

// Settled is a free data retrieval call binding the contract method 0xd945af1d.
//
// Solidity: function settled(bytes32 ) view returns(bool)
func (_AssetHolder *AssetHolderCallerSession) Settled(arg0 [32]byte) (bool, error) {
	return _AssetHolder.Contract.Settled(&_AssetHolder.CallOpts, arg0)
}

// Deposit is a paid mutator transaction binding the contract method 0x1de26e16.
//
// Solidity: function deposit(bytes32 fundingID, uint256 amount) payable returns()
func (_AssetHolder *AssetHolderTransactor) Deposit(opts *bind.TransactOpts, fundingID [32]byte, amount *big.Int) (*types.Transaction, error) {
	return _AssetHolder.contract.Transact(opts, "deposit", fundingID, amount)
}

// Deposit is a paid mutator transaction binding the contract method 0x1de26e16.
//
// Solidity: function deposit(bytes32 fundingID, uint256 amount) payable returns()
func (_AssetHolder *AssetHolderSession) Deposit(fundingID [32]byte, amount *big.Int) (*types.Transaction, error) {
	return _AssetHolder.Contract.Deposit(&_AssetHolder.TransactOpts, fundingID, amount)
}

// Deposit is a paid mutator transaction binding the contract method 0x1de26e16.
//
// Solidity: function deposit(bytes32 fundingID, uint256 amount) payable returns()
func (_AssetHolder *AssetHolderTransactorSession) Deposit(fundingID [32]byte, amount *big.Int) (*types.Transaction, error) {
	return _AssetHolder.Contract.Deposit(&_AssetHolder.TransactOpts, fundingID, amount)
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
// Solidity: function withdraw((bytes32,address,address,uint256) authorization, bytes signature) returns()
func (_AssetHolder *AssetHolderTransactor) Withdraw(opts *bind.TransactOpts, authorization AssetHolderWithdrawalAuth, signature []byte) (*types.Transaction, error) {
	return _AssetHolder.contract.Transact(opts, "withdraw", authorization, signature)
}

// Withdraw is a paid mutator transaction binding the contract method 0x4ed4283c.
//
// Solidity: function withdraw((bytes32,address,address,uint256) authorization, bytes signature) returns()
func (_AssetHolder *AssetHolderSession) Withdraw(authorization AssetHolderWithdrawalAuth, signature []byte) (*types.Transaction, error) {
	return _AssetHolder.Contract.Withdraw(&_AssetHolder.TransactOpts, authorization, signature)
}

// Withdraw is a paid mutator transaction binding the contract method 0x4ed4283c.
//
// Solidity: function withdraw((bytes32,address,address,uint256) authorization, bytes signature) returns()
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
	FundingID [32]byte
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterDeposited is a free log retrieval operation binding the contract event 0xcd2fe07293de5928c5df9505b65a8d6506f8668dfe81af09090920687edc48a9.
//
// Solidity: event Deposited(bytes32 indexed fundingID, uint256 amount)
func (_AssetHolder *AssetHolderFilterer) FilterDeposited(opts *bind.FilterOpts, fundingID [][32]byte) (*AssetHolderDepositedIterator, error) {

	var fundingIDRule []interface{}
	for _, fundingIDItem := range fundingID {
		fundingIDRule = append(fundingIDRule, fundingIDItem)
	}

	logs, sub, err := _AssetHolder.contract.FilterLogs(opts, "Deposited", fundingIDRule)
	if err != nil {
		return nil, err
	}
	return &AssetHolderDepositedIterator{contract: _AssetHolder.contract, event: "Deposited", logs: logs, sub: sub}, nil
}

// WatchDeposited is a free log subscription operation binding the contract event 0xcd2fe07293de5928c5df9505b65a8d6506f8668dfe81af09090920687edc48a9.
//
// Solidity: event Deposited(bytes32 indexed fundingID, uint256 amount)
func (_AssetHolder *AssetHolderFilterer) WatchDeposited(opts *bind.WatchOpts, sink chan<- *AssetHolderDeposited, fundingID [][32]byte) (event.Subscription, error) {

	var fundingIDRule []interface{}
	for _, fundingIDItem := range fundingID {
		fundingIDRule = append(fundingIDRule, fundingIDItem)
	}

	logs, sub, err := _AssetHolder.contract.WatchLogs(opts, "Deposited", fundingIDRule)
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

// ParseDeposited is a log parse operation binding the contract event 0xcd2fe07293de5928c5df9505b65a8d6506f8668dfe81af09090920687edc48a9.
//
// Solidity: event Deposited(bytes32 indexed fundingID, uint256 amount)
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

// AssetHolderWithdrawnIterator is returned from FilterWithdrawn and is used to iterate over the raw logs and unpacked data for Withdrawn events raised by the AssetHolder contract.
type AssetHolderWithdrawnIterator struct {
	Event *AssetHolderWithdrawn // Event containing the contract specifics and raw log

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
func (it *AssetHolderWithdrawnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AssetHolderWithdrawn)
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
		it.Event = new(AssetHolderWithdrawn)
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
func (it *AssetHolderWithdrawnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AssetHolderWithdrawnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AssetHolderWithdrawn represents a Withdrawn event raised by the AssetHolder contract.
type AssetHolderWithdrawn struct {
	FundingID [32]byte
	Amount    *big.Int
	Receiver  common.Address
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterWithdrawn is a free log retrieval operation binding the contract event 0xd0b6e7d0170f56c62f87de6a8a47a0ccf41c86ffb5084d399d8eb62e823f2a81.
//
// Solidity: event Withdrawn(bytes32 indexed fundingID, uint256 amount, address receiver)
func (_AssetHolder *AssetHolderFilterer) FilterWithdrawn(opts *bind.FilterOpts, fundingID [][32]byte) (*AssetHolderWithdrawnIterator, error) {

	var fundingIDRule []interface{}
	for _, fundingIDItem := range fundingID {
		fundingIDRule = append(fundingIDRule, fundingIDItem)
	}

	logs, sub, err := _AssetHolder.contract.FilterLogs(opts, "Withdrawn", fundingIDRule)
	if err != nil {
		return nil, err
	}
	return &AssetHolderWithdrawnIterator{contract: _AssetHolder.contract, event: "Withdrawn", logs: logs, sub: sub}, nil
}

// WatchWithdrawn is a free log subscription operation binding the contract event 0xd0b6e7d0170f56c62f87de6a8a47a0ccf41c86ffb5084d399d8eb62e823f2a81.
//
// Solidity: event Withdrawn(bytes32 indexed fundingID, uint256 amount, address receiver)
func (_AssetHolder *AssetHolderFilterer) WatchWithdrawn(opts *bind.WatchOpts, sink chan<- *AssetHolderWithdrawn, fundingID [][32]byte) (event.Subscription, error) {

	var fundingIDRule []interface{}
	for _, fundingIDItem := range fundingID {
		fundingIDRule = append(fundingIDRule, fundingIDItem)
	}

	logs, sub, err := _AssetHolder.contract.WatchLogs(opts, "Withdrawn", fundingIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AssetHolderWithdrawn)
				if err := _AssetHolder.contract.UnpackLog(event, "Withdrawn", log); err != nil {
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
func (_AssetHolder *AssetHolderFilterer) ParseWithdrawn(log types.Log) (*AssetHolderWithdrawn, error) {
	event := new(AssetHolderWithdrawn)
	if err := _AssetHolder.contract.UnpackLog(event, "Withdrawn", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ChannelABI is the input ABI used to generate the binding from.
const ChannelABI = "[]"

// ChannelBin is the compiled bytecode used for deploying new contracts.
var ChannelBin = "0x60566023600b82828239805160001a607314601657fe5b30600052607381538281f3fe73000000000000000000000000000000000000000030146080604052600080fdfea2646970667358221220b4f8b04a9a63efebe931e98c706a7c1297e91530722dd4a8235a133a905b409964736f6c63430007000033"

// DeployChannel deploys a new Ethereum contract, binding an instance of Channel to it.
func DeployChannel(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Channel, error) {
	parsed, err := abi.JSON(strings.NewReader(ChannelABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(ChannelBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Channel{ChannelCaller: ChannelCaller{contract: contract}, ChannelTransactor: ChannelTransactor{contract: contract}, ChannelFilterer: ChannelFilterer{contract: contract}}, nil
}

// Channel is an auto generated Go binding around an Ethereum contract.
type Channel struct {
	ChannelCaller     // Read-only binding to the contract
	ChannelTransactor // Write-only binding to the contract
	ChannelFilterer   // Log filterer for contract events
}

// ChannelCaller is an auto generated read-only Go binding around an Ethereum contract.
type ChannelCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ChannelTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ChannelTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ChannelFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ChannelFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ChannelSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ChannelSession struct {
	Contract     *Channel          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ChannelCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ChannelCallerSession struct {
	Contract *ChannelCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// ChannelTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ChannelTransactorSession struct {
	Contract     *ChannelTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// ChannelRaw is an auto generated low-level Go binding around an Ethereum contract.
type ChannelRaw struct {
	Contract *Channel // Generic contract binding to access the raw methods on
}

// ChannelCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ChannelCallerRaw struct {
	Contract *ChannelCaller // Generic read-only contract binding to access the raw methods on
}

// ChannelTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ChannelTransactorRaw struct {
	Contract *ChannelTransactor // Generic write-only contract binding to access the raw methods on
}

// NewChannel creates a new instance of Channel, bound to a specific deployed contract.
func NewChannel(address common.Address, backend bind.ContractBackend) (*Channel, error) {
	contract, err := bindChannel(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Channel{ChannelCaller: ChannelCaller{contract: contract}, ChannelTransactor: ChannelTransactor{contract: contract}, ChannelFilterer: ChannelFilterer{contract: contract}}, nil
}

// NewChannelCaller creates a new read-only instance of Channel, bound to a specific deployed contract.
func NewChannelCaller(address common.Address, caller bind.ContractCaller) (*ChannelCaller, error) {
	contract, err := bindChannel(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ChannelCaller{contract: contract}, nil
}

// NewChannelTransactor creates a new write-only instance of Channel, bound to a specific deployed contract.
func NewChannelTransactor(address common.Address, transactor bind.ContractTransactor) (*ChannelTransactor, error) {
	contract, err := bindChannel(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ChannelTransactor{contract: contract}, nil
}

// NewChannelFilterer creates a new log filterer instance of Channel, bound to a specific deployed contract.
func NewChannelFilterer(address common.Address, filterer bind.ContractFilterer) (*ChannelFilterer, error) {
	contract, err := bindChannel(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ChannelFilterer{contract: contract}, nil
}

// bindChannel binds a generic wrapper to an already deployed contract.
func bindChannel(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ChannelABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Channel *ChannelRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Channel.Contract.ChannelCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Channel *ChannelRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Channel.Contract.ChannelTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Channel *ChannelRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Channel.Contract.ChannelTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Channel *ChannelCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Channel.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Channel *ChannelTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Channel.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Channel *ChannelTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Channel.Contract.contract.Transact(opts, method, params...)
}

// ECDSAABI is the input ABI used to generate the binding from.
const ECDSAABI = "[]"

// ECDSABin is the compiled bytecode used for deploying new contracts.
var ECDSABin = "0x60566023600b82828239805160001a607314601657fe5b30600052607381538281f3fe73000000000000000000000000000000000000000030146080604052600080fdfea264697066735822122054ff38071b7d64affd4af55783e28f91712386c1e2fd62022fb27eb038cea01a64736f6c63430007000033"

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

// SafeMathABI is the input ABI used to generate the binding from.
const SafeMathABI = "[]"

// SafeMathBin is the compiled bytecode used for deploying new contracts.
var SafeMathBin = "0x60566023600b82828239805160001a607314601657fe5b30600052607381538281f3fe73000000000000000000000000000000000000000030146080604052600080fdfea264697066735822122057a65a7bdabbb92d0dac1b22d801009d8638a978295be7ac76cf75c8d85c137f64736f6c63430007000033"

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

// SigABI is the input ABI used to generate the binding from.
const SigABI = "[]"

// SigBin is the compiled bytecode used for deploying new contracts.
var SigBin = "0x60566023600b82828239805160001a607314601657fe5b30600052607381538281f3fe73000000000000000000000000000000000000000030146080604052600080fdfea2646970667358221220faf262fee47eb63262adaabdf3a4efa7e8e39dbc7d86be68aad3f30b69f06b7264736f6c63430007000033"

// DeploySig deploys a new Ethereum contract, binding an instance of Sig to it.
func DeploySig(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Sig, error) {
	parsed, err := abi.JSON(strings.NewReader(SigABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(SigBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Sig{SigCaller: SigCaller{contract: contract}, SigTransactor: SigTransactor{contract: contract}, SigFilterer: SigFilterer{contract: contract}}, nil
}

// Sig is an auto generated Go binding around an Ethereum contract.
type Sig struct {
	SigCaller     // Read-only binding to the contract
	SigTransactor // Write-only binding to the contract
	SigFilterer   // Log filterer for contract events
}

// SigCaller is an auto generated read-only Go binding around an Ethereum contract.
type SigCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SigTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SigTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SigFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SigFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SigSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SigSession struct {
	Contract     *Sig              // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SigCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SigCallerSession struct {
	Contract *SigCaller    // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// SigTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SigTransactorSession struct {
	Contract     *SigTransactor    // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SigRaw is an auto generated low-level Go binding around an Ethereum contract.
type SigRaw struct {
	Contract *Sig // Generic contract binding to access the raw methods on
}

// SigCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SigCallerRaw struct {
	Contract *SigCaller // Generic read-only contract binding to access the raw methods on
}

// SigTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SigTransactorRaw struct {
	Contract *SigTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSig creates a new instance of Sig, bound to a specific deployed contract.
func NewSig(address common.Address, backend bind.ContractBackend) (*Sig, error) {
	contract, err := bindSig(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Sig{SigCaller: SigCaller{contract: contract}, SigTransactor: SigTransactor{contract: contract}, SigFilterer: SigFilterer{contract: contract}}, nil
}

// NewSigCaller creates a new read-only instance of Sig, bound to a specific deployed contract.
func NewSigCaller(address common.Address, caller bind.ContractCaller) (*SigCaller, error) {
	contract, err := bindSig(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SigCaller{contract: contract}, nil
}

// NewSigTransactor creates a new write-only instance of Sig, bound to a specific deployed contract.
func NewSigTransactor(address common.Address, transactor bind.ContractTransactor) (*SigTransactor, error) {
	contract, err := bindSig(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SigTransactor{contract: contract}, nil
}

// NewSigFilterer creates a new log filterer instance of Sig, bound to a specific deployed contract.
func NewSigFilterer(address common.Address, filterer bind.ContractFilterer) (*SigFilterer, error) {
	contract, err := bindSig(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SigFilterer{contract: contract}, nil
}

// bindSig binds a generic wrapper to an already deployed contract.
func bindSig(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SigABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Sig *SigRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Sig.Contract.SigCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Sig *SigRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Sig.Contract.SigTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Sig *SigRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Sig.Contract.SigTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Sig *SigCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Sig.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Sig *SigTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Sig.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Sig *SigTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Sig.Contract.contract.Transact(opts, method, params...)
}
