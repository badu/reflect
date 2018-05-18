// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build mips mipsle

#include "textflag.h"
#include "funcdata.h"

#define	REGCTXT	R22

// methodValueCall is the code half of the function returned by makeMethodValue.
// See the comment on the declaration of methodValueCall in makefunc.go
// for more details.
// No arg size here; runtime pulls arg map out of the func value.
TEXT ·methodValueCall(SB),(NOSPLIT|WRAPPER),$8
	NO_LOCAL_POINTERS
	MOVW	REGCTXT, 4(R29)
	MOVW	$argframe+0(FP), R1
	MOVW	R1, 8(R29)
	JAL	·callMethod(SB)
	RET
