// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"
#include "funcdata.h"

// methodValueCall is the code half of the function returned by makeMethodValue.
// See the comment on the declaration of methodValueCall in makefunc.go
// for more details.
// No arg size here; runtime pulls arg map out of the func value.
TEXT ·methodValueCall(SB),(NOSPLIT|WRAPPER),$16
	NO_LOCAL_POINTERS
	MOVD	R12, 8(R15)
	MOVD	$argframe+0(FP), R3
	MOVD	R3, 16(R15)
	BL	·callMethod(SB)
	RET
