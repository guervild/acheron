// func getNtdllBaseAddr() uintptr
TEXT ·getNtdllBaseAddr(SB),$0

    // TEB->ProcessEnvironmentBlock
    XORQ AX, AX
    MOVQ 0x30(GS), AX
    MOVQ 0x60(AX), AX

    // PEB->Ldr
    MOVQ 0x18(AX), AX

    // PEB->Ldr->InMemoryOrderModuleList
    MOVQ 0x20(AX), AX

    // PEB->Ldr->InMemoryOrderModuleList->Flink (ntdll.dll)
    MOVQ (AX), AX

    // PEB->Ldr->InMemoryOrderModuleList->Flink DllBase
    MOVQ 0x20(AX), AX

    MOVQ AX, ret+0(FP)
    RET


// func getModuleEATAddr (moduleBase uintptr) uintptr
TEXT ·getModuleEATAddr(SB),$0-8
    MOVQ moduleBase+0(FP), AX

    XORQ R15, R15
    XORQ R14, R14

    // AX = IMAGE_DOS_HEADER->e_lfanew offset
    MOVB 0x3C(AX), R15

    // R15 = ntdll base + R15
    ADDQ AX, R15

    // R15 = R15 + OptionalHeader + DataDirectory offset
    ADDQ $0x88, R15

    // AX = ntdll base + IMAGE_DATA_DIRECTORY.VirtualAddress
    ADDL 0x0(R15), R14
    ADDQ R14, AX

    MOVQ AX, ret+8(FP)
    RET


// func getEATNumberOfFunctions(exportsBase uintptr) uint32
TEXT ·getEATNumberOfFunctions(SB),$0-8
    MOVQ exportsBase+0(FP), AX

    XORQ R15, R15

    // R15 = exportsBase + IMAGE_EXPORT_DIRECTORY.NumberOfFunctions
    MOVL 0x14(AX), R15

    MOVL R15, ret+8(FP)
    RET


// func getEATAddressOfFunctions(moduleBase,exportsBase uintptr) uintptr
TEXT ·getEATAddressOfFunctions(SB),$0-16
    MOVQ moduleBase+0(FP), AX
    MOVQ exportsBase+8(FP), R8

    XORQ SI, SI

    // R15 = exportsBase + IMAGE_EXPORT_DIRECTORY.AddressOfFunctions
    MOVL 0x1c(R8), SI

    // AX = exportsBase + AddressOfFunctions offset
    ADDQ SI, AX

    MOVQ AX, ret+16(FP)
    RET


// func getEATAddressOfNames(moduleBase,exportsBase uintptr) uintptr
TEXT ·getEATAddressOfNames(SB),$0-16
    MOVQ moduleBase+0(FP), AX
    MOVQ exportsBase+8(FP), R8

    XORQ SI, SI

    // SI = exportsBase + IMAGE_EXPORT_DIRECTORY.AddressOfNames
    MOVL 0x20(R8), SI

    // AX = exportsBase + AddressOfNames offset
    ADDQ SI, AX

    MOVQ AX, ret+16(FP)
    RET
