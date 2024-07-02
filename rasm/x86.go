package rasm

type X86Instruction uint

// Instruction constants (Very WIP)
const (
	_      = iota
	X86Mov = iota << 5
)

var x86InstrSearchMap = map[string]X86Instruction{
	"mov": X86Mov,
}

type X86Register uint

// Register constants (WIP)
const (
	_      = iota
	X86Rax = iota << 5
	X86Rcx
	X86Rdx
	X86Rbx
	X86Rsi
	X86Rdi
	X86Rsp
	X86Rbp
	X86R9
	X86R10
	X86R11
	X86R12
	X86R13
	X86R14
	X86R15

	X86Eax
	X86Ecx
	X86Edx
	X86Ebx
	X86Esi
	X86Edi
	X86Esp
	X86Ebp
	X86R9d
	X86R10d
	X86R11d
	X86R12d
	X86R13d
	X86R14d
	X86R15d

	X86Ax
	X86Cx
	X86Dx
	X86Bx
	X86Si
	X86Di
	X86Sp
	X86Bp
	X86R9w
	X86R10w
	X86R11w
	X86R12w
	X86R13w
	X86R14w
	X86R15w

	X86Al
	X86Cl
	X86Dl
	X86Bl
	X86Sil
	X86Dil
	X86Spl
	X86Bpl
	X86R9b
	X86R10b
	X86R11b
	X86R12b
	X86R13b
	X86R14b
	X86R15b

	X86Ah
	X86Ch
	X86Dh
	X86Bh
)

var x86RegisterSearchMap = map[string]X86Instruction{
	"rax": X86Rax,
	"rcx": X86Rcx,
	"rdx": X86Rdx,
	"rbx": X86Rbx,
	"rsi": X86Rsi,
	"rdi": X86Rdi,
	"rsp": X86Rsp,
	"rbp": X86Rbp,
	"r9":  X86R9,
	"r10": X86R10,
	"r11": X86R11,
	"r12": X86R12,
	"r13": X86R13,
	"r14": X86R14,
	"r15": X86R15,

	"eax":  X86Eax,
	"ecx":  X86Ecx,
	"edx":  X86Edx,
	"ebx":  X86Ebx,
	"esi":  X86Esi,
	"edi":  X86Edi,
	"esp":  X86Esp,
	"ebp":  X86Ebp,
	"r9d":  X86R9d,
	"r10d": X86R10d,
	"r11d": X86R11d,
	"r12d": X86R12d,
	"r13d": X86R13d,
	"r14d": X86R14d,
	"r15d": X86R15d,

	"ax":   X86Ax,
	"cx":   X86Cx,
	"dx":   X86Dx,
	"bx":   X86Bx,
	"si":   X86Si,
	"di":   X86Di,
	"sp":   X86Sp,
	"bp":   X86Bp,
	"r9w":  X86R9w,
	"r10w": X86R10w,
	"r11w": X86R11w,
	"r12w": X86R12w,
	"r13w": X86R13w,
	"r14w": X86R14w,
	"r15w": X86R15w,

	"al":   X86Al,
	"cl":   X86Cl,
	"dl":   X86Dl,
	"bl":   X86Bl,
	"sil":  X86Sil,
	"dil":  X86Dil,
	"spl":  X86Spl,
	"bpl":  X86Bpl,
	"r9b":  X86R9b,
	"r10b": X86R10b,
	"r11b": X86R11b,
	"r12b": X86R12b,
	"r13b": X86R13b,
	"r14b": X86R14b,
	"r15b": X86R15b,

	"ah": X86Ah,
	"ch": X86Ch,
	"dh": X86Dh,
	"bh": X86Bh,
}
