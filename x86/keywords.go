package x86

type Instruction uint

// Instruction constants (Very WIP)
const (
	_   = iota
	Mov = iota << 5
)

var InstrSearchMap = map[string]Instruction{
	"mov": Mov,
}

type Register uint

// Register constants (WIP)
const (
	_   = iota
	Rax = iota << 5
	Rcx
	Rdx
	Rbx
	Rsi
	Rdi
	Rsp
	Rbp
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

var RegisterSearchMap = map[string]Instruction{
	"rax": Rax,
	"rcx": Rcx,
	"rdx": Rdx,
	"rbx": Rbx,
	"rsi": Rsi,
	"rdi": Rdi,
	"rsp": Rsp,
	"rbp": Rbp,
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
