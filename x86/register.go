package x86

type Register uint

func (reg Register) EncodeByte() byte {
	switch reg {
	case AL, AX, EAX, RAX, R8B, R8W, R8D, R8:
		return 0
	case CL, CX, ECX, RCX, R9B, R9W, R9D, R9:
		return 1
	case DL, DX, EDX, RDX, R10B, R10W, R10D, R10:
		return 2
	case BL, BX, EBX, RBX, R11B, R11W, R11D, R11:
		return 3
	case AH, SP, ESP, SPL, RSP, R12B, R12W, R12D, R12:
		return 4
	case CH, BP, EBP, BPL, RBP, R13B, R13W, R13D, R13:
		return 5
	case DH, SI, ESI, SIL, RSI, R14B, R14W, R14D, R14:
		return 6
	case BH, DI, EDI, DIL, RDI, R15B, R15W, R15D, R15:
		return 7
	case NilReg:
		return 0
	default:
		panic("given register is unsupported ")
	}
}

func (reg Register) Size() uint {
	switch reg {
	case AL, CL, DL, BL, SIL, DIL, SPL, BPL, R8B, R9B, R10B,
		R11B, R12B, R13B, R14B, R15B, AH, CH, DH, BH:
		return 8
	case AX, CX, DX, BX, SI, DI, SP, BP, R8W, R9W, R10W, R11W, R12W, R13W, R14W, R15W:
		return 16
	case EAX, ECX, EDX, EBX, ESI, EDI, ESP, EBP, R8D, R9D, R10D, R11D, R12D, R13D, R14D, R15D:
		return 32
	case RAX, RCX, RDX, RBX, RSI, RDI, RSP, RBP, R8, R9, R10, R11, R12, R13, R14, R15:
		return 64
	case NilReg:
		return 0
	}

	panic("unreachable")
}

// Reports if the register requires an REX prefix to be encoded.
func (reg Register) IsRex() bool {
	switch reg {
	case RAX, RCX, RDX, RBX, RSI, RDI, RSP, RBP, R8B, R9B, R10B, R11B, R12B, R13B, R14B, R15B, R8W,
		R9W, R10W, R11W, R12W, R13W, R14W, R15W, R8D, R9D, R10D, R11D, R12D, R13D, R14D, R15D, R8, R9,
		R10, R11, R12, R13, R14, R15, SIL, DIL, SPL, BPL:
		return true
	default:
		return false
	}
}

// Reports if the register needs REX.B set (also used to check for the need of REX.R)
func (reg Register) IsRexB() bool {
	switch reg {
	case R8B, R9B, R10B, R11B, R12B, R13B, R14B, R15B, R8W, R9W, R10W, R11W, R12W, R13W, R14W, R15W,
		R8D, R9D, R10D, R11D, R12D, R13D, R14D, R15D, R8, R9, R10, R11, R12, R13, R14, R15:
		return true
	default:
		return false
	}
}

// Reports if the register can not be encoded if a REX byte is present.
func (reg Register) IsRexExcluded() bool {
	switch reg {
	case AH, CH, DH, BH:
		return true
	default:
		return false
	}
}

func (reg Register) isARegister() bool {
	switch reg {
	case AL, AX, EAX, RAX:
		return true
	default:
		return false
	}
}

// Register constants (WIP)
const (
	NilReg          = iota
	RAX    Register = iota << 5
	RCX
	RDX
	RBX
	RSI
	RDI
	RSP
	RBP
	R8
	R9
	R10
	R11
	R12
	R13
	R14
	R15

	EAX
	ECX
	EDX
	EBX
	ESI
	EDI
	ESP
	EBP
	R8D
	R9D
	R10D
	R11D
	R12D
	R13D
	R14D
	R15D

	AX
	CX
	DX
	BX
	SI
	DI
	SP
	BP
	R8W
	R9W
	R10W
	R11W
	R12W
	R13W
	R14W
	R15W

	AL
	CL
	DL
	BL
	SIL
	DIL
	SPL
	BPL
	R8B
	R9B
	R10B
	R11B
	R12B
	R13B
	R14B
	R15B

	AH
	CH
	DH
	BH
)

var RegisterSearchMap = map[string]Register{
	"rax": RAX,
	"rcx": RCX,
	"rdx": RDX,
	"rbx": RBX,
	"rsi": RSI,
	"rdi": RDI,
	"rsp": RSP,
	"rbp": RBP,
	"r8":  R8,
	"r9":  R9,
	"r10": R10,
	"r11": R11,
	"r12": R12,
	"r13": R13,
	"r14": R14,
	"r15": R15,

	"eax":  EAX,
	"ecx":  ECX,
	"edx":  EDX,
	"ebx":  EBX,
	"esi":  ESI,
	"edi":  EDI,
	"esp":  ESP,
	"ebp":  EBP,
	"r8d":  R8D,
	"r9d":  R9D,
	"r10d": R10D,
	"r11d": R11D,
	"r12d": R12D,
	"r13d": R13D,
	"r14d": R14D,
	"r15d": R15D,

	"ax":   AX,
	"cx":   CX,
	"dx":   DX,
	"bx":   BX,
	"si":   SI,
	"di":   DI,
	"sp":   SP,
	"bp":   BP,
	"r8w":  R8W,
	"r9w":  R9W,
	"r10w": R10W,
	"r11w": R11W,
	"r12w": R12W,
	"r13w": R13W,
	"r14w": R14W,
	"r15w": R15W,

	"al":   AL,
	"cl":   CL,
	"dl":   DL,
	"bl":   BL,
	"sil":  SIL,
	"dil":  DIL,
	"spl":  SPL,
	"bpl":  BPL,
	"r8b":  R8B,
	"r9b":  R9B,
	"r10b": R10B,
	"r11b": R11B,
	"r12b": R12B,
	"r13b": R13B,
	"r14b": R14B,
	"r15b": R15B,

	"ah": AH,
	"ch": CH,
	"dh": DH,
	"bh": BH,
}
