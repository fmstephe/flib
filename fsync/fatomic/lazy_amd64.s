TEXT ·LazyStore(SB),$0-16
        MOVQ    addr+0(FP), BP
        MOVQ    val+8(FP), AX
        MOVQ   AX, 0(BP)
        RET
