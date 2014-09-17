TEXT ·Counter(SB),$0-8
	RDTSC
        SHLQ    $32, DX
        ADDQ    DX, AX
	MOVQ	AX, count+0(FP)
        RET

TEXT ·cpuid(SB),$0-24
	MOVL	eaxi+0(FP), AX
	MOVL	$0, BX
	MOVL	$0, CX
	MOVL	$0, DX
	CPUID
        MOVL	AX, eax+8(FP)
        MOVL	BX, ebx+12(FP) // Do I need to preserve this?
        MOVL	CX, ecx+16(FP)
	MOVL	DX, edx+20(FP)	
        RET
