package safe

import "sync/atomic"

type AtomicBool int32

func (a *AtomicBool) IsSet() bool { return atomic.LoadInt32((*int32)(a)) != 0 }
func (a *AtomicBool) SetFalse()   { atomic.StoreInt32((*int32)(a), 0) }
func (a *AtomicBool) SetTrue()    { atomic.StoreInt32((*int32)(a), 1) }
