package x86

type Register uint

func (reg Register) EncodeByte() byte {
	switch reg {
	case Al, Ax, Eax, Rax, R8b, R8w, R8d, R8:
		return 0
	case Cl, Cx, Ecx, Rcx, R9b, R9w, R9d, R9:
		return 1
	case Dl, Dx, Edx, Rdx, R10b, R10w, R10d, R10:
		return 2
	case Bl, Bx, Ebx, Rbx, R11b, R11w, R11d, R11:
		return 3
	case Ah, Sp, Esp, Spl, Rsp, R12b, R12w, R12d, R12:
		return 4
	case Ch, Bp, Ebp, Bpl, Rbp, R13b, R13w, R13d, R13:
		return 5
	case Dh, Si, Esi, Sil, Rsi, R14b, R14w, R14d, R14:
		return 6
	case Bh, Di, Edi, Dil, Rdi, R15b, R15w, R15d, R15:
		return 7
	case NilReg:
		return 0
	default:
		panic("given register is unsupported ")
	}
}

func (reg Register) Size() uint {
	switch reg {
	case Al, Cl, Dl, Bl, Sil, Dil, Spl, Bpl, R8b, R9b, R10b, R11b, R12b, R13b, R14b, R15b, Ah, Ch, Dh, Bh:
		return 8
	case Ax, Cx, Dx, Bx, Si, Di, Sp, Bp, R8w, R9w, R10w, R11w, R12w, R13w, R14w, R15w:
		return 16
	case Eax, Ecx, Edx, Ebx, Esi, Edi, Esp, Ebp, R8d, R9d, R10d, R11d, R12d, R13d, R14d, R15d:
		return 32
	case Rax, Rcx, Rdx, Rbx, Rsi, Rdi, Rsp, Rbp, R8, R9, R10, R11, R12, R13, R14, R15:
		return 64
	case NilReg:
		return 0
	}

	panic("unreachable")
}

// Reports if the register requires an REX prefix to be encoded.
func (reg Register) IsRex() bool {
	switch reg {
	case Rax, Rcx, Rdx, Rbx, Rsi, Rdi, Rsp, Rbp, R8b, R9b, R10b, R11b, R12b, R13b, R14b, R15b, R8w,
		R9w, R10w, R11w, R12w, R13w, R14w, R15w, R8d, R9d, R10d, R11d, R12d, R13d, R14d, R15d, R8, R9,
		R10, R11, R12, R13, R14, R15, Sil, Dil, Spl, Bpl:
		return true
	default:
		return false
	}
}

// Reports if the register needs REX.B set (also used to check for the need of REX.R)
func (reg Register) IsRexB() bool {
	switch reg {
	case R8b, R9b, R10b, R11b, R12b, R13b, R14b, R15b, R8w, R9w, R10w, R11w, R12w, R13w, R14w, R15w,
		R8d, R9d, R10d, R11d, R12d, R13d, R14d, R15d, R8, R9, R10, R11, R12, R13, R14, R15:
		return true
	default:
		return false
	}
}

// Reports if the register can not be encoded if a REX byte is present.
func (reg Register) IsRexExcluded() bool {
	switch reg {
	case Ah, Ch, Dh, Bh:
		return true
	default:
		return false
	}
}

func (reg Register) isARegister() bool {
	switch reg {
	case Al, Ax, Eax, Rax:
		return true
	default:
		return false
	}
}

// Register constants (WIP)
const (
	NilReg          = iota
	Rax    Register = iota << 5
	Rcx
	Rdx
	Rbx
	Rsi
	Rdi
	Rsp
	Rbp
	R8
	R9
	R10
	R11
	R12
	R13
	R14
	R15

	Eax
	Ecx
	Edx
	Ebx
	Esi
	Edi
	Esp
	Ebp
	R8d
	R9d
	R10d
	R11d
	R12d
	R13d
	R14d
	R15d

	Ax
	Cx
	Dx
	Bx
	Si
	Di
	Sp
	Bp
	R8w
	R9w
	R10w
	R11w
	R12w
	R13w
	R14w
	R15w

	Al
	Cl
	Dl
	Bl
	Sil
	Dil
	Spl
	Bpl
	R8b
	R9b
	R10b
	R11b
	R12b
	R13b
	R14b
	R15b

	Ah
	Ch
	Dh
	Bh
)

var RegisterSearchMap = map[string]Register{
	"rax": Rax,
	"rcx": Rcx,
	"rdx": Rdx,
	"rbx": Rbx,
	"rsi": Rsi,
	"rdi": Rdi,
	"rsp": Rsp,
	"rbp": Rbp,
	"r8":  R8,
	"r9":  R9,
	"r10": R10,
	"r11": R11,
	"r12": R12,
	"r13": R13,
	"r14": R14,
	"r15": R15,

	"eax":  Eax,
	"ecx":  Ecx,
	"edx":  Edx,
	"ebx":  Ebx,
	"esi":  Esi,
	"edi":  Edi,
	"esp":  Esp,
	"ebp":  Ebp,
	"r8d":  R8d,
	"r9d":  R9d,
	"r10d": R10d,
	"r11d": R11d,
	"r12d": R12d,
	"r13d": R13d,
	"r14d": R14d,
	"r15d": R15d,

	"ax":   Ax,
	"cx":   Cx,
	"dx":   Dx,
	"bx":   Bx,
	"si":   Si,
	"di":   Di,
	"sp":   Sp,
	"bp":   Bp,
	"r8w":  R8w,
	"r9w":  R9w,
	"r10w": R10w,
	"r11w": R11w,
	"r12w": R12w,
	"r13w": R13w,
	"r14w": R14w,
	"r15w": R15w,

	"al":   Al,
	"cl":   Cl,
	"dl":   Dl,
	"bl":   Bl,
	"sil":  Sil,
	"dil":  Dil,
	"spl":  Spl,
	"bpl":  Bpl,
	"r8b":  R8b,
	"r9b":  R9b,
	"r10b": R10b,
	"r11b": R11b,
	"r12b": R12b,
	"r13b": R13b,
	"r14b": R14b,
	"r15b": R15b,

	"ah": Ah,
	"ch": Ch,
	"dh": Dh,
	"bh": Bh,
}
