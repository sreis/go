// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "strings"

// copied from ../../amd64/reg.go
var regNamesAMD64 = []string{
	".AX",
	".CX",
	".DX",
	".BX",
	".SP",
	".BP",
	".SI",
	".DI",
	".R8",
	".R9",
	".R10",
	".R11",
	".R12",
	".R13",
	".R14",
	".R15",
	".X0",
	".X1",
	".X2",
	".X3",
	".X4",
	".X5",
	".X6",
	".X7",
	".X8",
	".X9",
	".X10",
	".X11",
	".X12",
	".X13",
	".X14",
	".X15",

	// pseudo-registers
	".SB",
	".FLAGS",
}

func init() {
	// Make map from reg names to reg integers.
	if len(regNamesAMD64) > 64 {
		panic("too many registers")
	}
	num := map[string]int{}
	for i, name := range regNamesAMD64 {
		if name[0] != '.' {
			panic("register name " + name + " does not start with '.'")
		}
		num[name[1:]] = i
	}
	buildReg := func(s string) regMask {
		m := regMask(0)
		for _, r := range strings.Split(s, " ") {
			if n, ok := num[r]; ok {
				m |= regMask(1) << uint(n)
				continue
			}
			panic("register " + r + " not found")
		}
		return m
	}

	gp := buildReg("AX CX DX BX BP SI DI R8 R9 R10 R11 R12 R13 R14 R15")
	gpsp := gp | buildReg("SP")
	gpspsb := gpsp | buildReg("SB")
	flags := buildReg("FLAGS")
	gp01 := regInfo{[]regMask{}, 0, []regMask{gp}}
	gp11 := regInfo{[]regMask{gpsp}, 0, []regMask{gp}}
	gp11sb := regInfo{[]regMask{gpspsb}, 0, []regMask{gp}}
	gp21 := regInfo{[]regMask{gpsp, gpsp}, 0, []regMask{gp}}
	gp21sb := regInfo{[]regMask{gpspsb, gpsp}, 0, []regMask{gp}}
	gp21shift := regInfo{[]regMask{gpsp, buildReg("CX")}, 0, []regMask{gp}}
	gp2flags := regInfo{[]regMask{gpsp, gpsp}, 0, []regMask{flags}}
	gp1flags := regInfo{[]regMask{gpsp}, 0, []regMask{flags}}
	flagsgp1 := regInfo{[]regMask{flags}, 0, []regMask{gp}}
	gpload := regInfo{[]regMask{gpspsb, 0}, 0, []regMask{gp}}
	gploadidx := regInfo{[]regMask{gpspsb, gpsp, 0}, 0, []regMask{gp}}
	gpstore := regInfo{[]regMask{gpspsb, gpsp, 0}, 0, nil}
	gpstoreconst := regInfo{[]regMask{gpspsb, 0}, 0, nil}
	gpstoreidx := regInfo{[]regMask{gpspsb, gpsp, gpsp, 0}, 0, nil}
	flagsgp := regInfo{[]regMask{flags}, 0, []regMask{gp}}
	cmov := regInfo{[]regMask{flags, gp, gp}, 0, []regMask{gp}}

	// Suffixes encode the bit width of various instructions.
	// Q = 64 bit, L = 32 bit, W = 16 bit, B = 8 bit

	// TODO: 2-address instructions.  Mark ops as needing matching input/output regs.
	var AMD64ops = []opData{
		{name: "ADDQ", reg: gp21},                    // arg0 + arg1
		{name: "ADDQconst", reg: gp11},               // arg0 + auxint
		{name: "SUBQ", reg: gp21, asm: "SUBQ"},       // arg0 - arg1
		{name: "SUBQconst", reg: gp11, asm: "SUBQ"},  // arg0 - auxint
		{name: "MULQ", reg: gp21, asm: "IMULQ"},      // arg0 * arg1
		{name: "MULQconst", reg: gp11, asm: "IMULQ"}, // arg0 * auxint
		{name: "ANDQ", reg: gp21, asm: "ANDQ"},       // arg0 & arg1
		{name: "ANDQconst", reg: gp11, asm: "ANDQ"},  // arg0 & auxint
		{name: "SHLQ", reg: gp21shift, asm: "SHLQ"},  // arg0 << arg1, shift amount is mod 64
		{name: "SHLQconst", reg: gp11, asm: "SHLQ"},  // arg0 << auxint, shift amount 0-63
		{name: "SHRQ", reg: gp21shift, asm: "SHRQ"},  // unsigned arg0 >> arg1, shift amount is mod 64
		{name: "SHRQconst", reg: gp11, asm: "SHRQ"},  // unsigned arg0 >> auxint, shift amount 0-63
		{name: "SARQ", reg: gp21shift, asm: "SARQ"},  // signed arg0 >> arg1, shift amount is mod 64
		{name: "SARQconst", reg: gp11, asm: "SARQ"},  // signed arg0 >> auxint, shift amount 0-63

		{name: "NEGQ", reg: gp11},                   // -arg0
		{name: "XORQconst", reg: gp11, asm: "XORQ"}, // arg0^auxint

		{name: "CMPQ", reg: gp2flags, asm: "CMPQ"},      // arg0 compare to arg1
		{name: "CMPQconst", reg: gp1flags, asm: "CMPQ"}, // arg0 compare to auxint
		{name: "TESTQ", reg: gp2flags, asm: "TESTQ"},    // (arg0 & arg1) compare to 0
		{name: "TESTB", reg: gp2flags, asm: "TESTB"},    // (arg0 & arg1) compare to 0

		{name: "SBBQcarrymask", reg: flagsgp1, asm: "SBBQ"}, // (int64)(-1) if carry is set, 0 if carry is clear.

		{name: "SETEQ", reg: flagsgp}, // extract == condition from arg0
		{name: "SETNE", reg: flagsgp}, // extract != condition from arg0
		{name: "SETL", reg: flagsgp},  // extract signed < condition from arg0
		{name: "SETLE", reg: flagsgp}, // extract signed <= condition from arg0
		{name: "SETG", reg: flagsgp},  // extract signed > condition from arg0
		{name: "SETGE", reg: flagsgp}, // extract signed >= condition from arg0
		{name: "SETB", reg: flagsgp},  // extract unsigned < condition from arg0

		{name: "CMOVQCC", reg: cmov}, // carry clear

		{name: "MOVLQSX", reg: gp11, asm: "MOVLQSX"}, // extend arg0 from int32 to int64
		{name: "MOVWQSX", reg: gp11, asm: "MOVWQSX"}, // extend arg0 from int16 to int64
		{name: "MOVBQSX", reg: gp11, asm: "MOVBQSX"}, // extend arg0 from int8 to int64

		{name: "MOVQconst", reg: gp01}, // auxint
		{name: "LEAQ", reg: gp11sb},    // arg0 + auxint + offset encoded in aux
		{name: "LEAQ1", reg: gp21sb},   // arg0 + arg1 + auxint
		{name: "LEAQ2", reg: gp21sb},   // arg0 + 2*arg1 + auxint
		{name: "LEAQ4", reg: gp21sb},   // arg0 + 4*arg1 + auxint
		{name: "LEAQ8", reg: gp21sb},   // arg0 + 8*arg1 + auxint

		{name: "MOVBload", reg: gpload, asm: "MOVB"},        // load byte from arg0+auxint. arg1=mem
		{name: "MOVBQZXload", reg: gpload},                  // ditto, extend to uint64
		{name: "MOVBQSXload", reg: gpload},                  // ditto, extend to int64
		{name: "MOVWload", reg: gpload, asm: "MOVW"},        // load 2 bytes from arg0+auxint. arg1=mem
		{name: "MOVLload", reg: gpload, asm: "MOVL"},        // load 4 bytes from arg0+auxint. arg1=mem
		{name: "MOVQload", reg: gpload, asm: "MOVQ"},        // load 8 bytes from arg0+auxint. arg1=mem
		{name: "MOVQloadidx8", reg: gploadidx, asm: "MOVQ"}, // load 8 bytes from arg0+8*arg1+auxint. arg2=mem
		{name: "MOVBstore", reg: gpstore, asm: "MOVB"},      // store byte in arg1 to arg0+auxint. arg2=mem
		{name: "MOVWstore", reg: gpstore, asm: "MOVW"},      // store 2 bytes in arg1 to arg0+auxint. arg2=mem
		{name: "MOVLstore", reg: gpstore, asm: "MOVL"},      // store 4 bytes in arg1 to arg0+auxint. arg2=mem
		{name: "MOVQstore", reg: gpstore, asm: "MOVQ"},      // store 8 bytes in arg1 to arg0+auxint. arg2=mem
		{name: "MOVQstoreidx8", reg: gpstoreidx},            // store 8 bytes in arg2 to arg0+8*arg1+auxint. arg3=mem

		{name: "MOVXzero", reg: gpstoreconst}, // store auxint 0 bytes into arg0 using a series of MOV instructions. arg1=mem.
		// TODO: implement this when register clobbering works
		{name: "REPSTOSQ", reg: regInfo{[]regMask{buildReg("DI"), buildReg("CX")}, buildReg("DI AX CX"), nil}}, // store arg1 8-byte words containing zero into arg0 using STOSQ. arg2=mem.

		// Load/store from global. Same as the above loads, but arg0 is missing and
		// aux is a GlobalOffset instead of an int64.
		{name: "MOVQloadglobal"},  // Load from aux.(GlobalOffset).  arg0 = memory
		{name: "MOVQstoreglobal"}, // store arg0 to aux.(GlobalOffset).  arg1=memory, returns memory.

		//TODO: set register clobber to everything?
		{name: "CALLstatic"},                                                            // call static function aux.(*gc.Sym).  arg0=mem, returns mem
		{name: "CALLclosure", reg: regInfo{[]regMask{gpsp, buildReg("DX"), 0}, 0, nil}}, // call function via closure.  arg0=codeptr, arg1=closure, arg2=mem returns mem

		{name: "REPMOVSB", reg: regInfo{[]regMask{buildReg("DI"), buildReg("SI"), buildReg("CX")}, buildReg("DI SI CX"), nil}}, // move arg2 bytes from arg1 to arg0.  arg3=mem, returns memory

		{name: "ADDL", reg: gp21, asm: "ADDL"}, // arg0+arg1
		{name: "ADDW", reg: gp21, asm: "ADDW"}, // arg0+arg1
		{name: "ADDB", reg: gp21, asm: "ADDB"}, // arg0+arg1

		// (InvertFlags (CMPQ a b)) == (CMPQ b a)
		// So if we want (SETL (CMPQ a b)) but we can't do that because a is a constant,
		// then we do (SETL (InvertFlags (CMPQ b a))) instead.
		// Rewrites will convert this to (SETG (CMPQ b a)).
		// InvertFlags is a pseudo-op which can't appear in assembly output.
		{name: "InvertFlags"}, // reverse direction of arg0
	}

	var AMD64blocks = []blockData{
		{name: "EQ"},
		{name: "NE"},
		{name: "LT"},
		{name: "LE"},
		{name: "GT"},
		{name: "GE"},
		{name: "ULT"},
		{name: "ULE"},
		{name: "UGT"},
		{name: "UGE"},
	}

	archs = append(archs, arch{"AMD64", AMD64ops, AMD64blocks, regNamesAMD64})
}
