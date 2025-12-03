package mempool

import (
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"

	"golang.org/x/sys/cpu"
)

// Pool provides a form of SimplePool
// with the addition of concurrency safety.
type Pool[T any] struct {
	UnsafePool

	// New is an optionally provided
	// allocator used when no value
	// is available for use in pool.
	New func() T

	// Reset is an optionally provided
	// value resetting function called
	// on passed value to Put().
	Reset func(T) bool
}

func NewPool[T any](new func() T, reset func(T) bool, check func(current, victim int) bool) Pool[T] {
	return Pool[T]{
		New:        new,
		Reset:      reset,
		UnsafePool: NewUnsafePool(check),
	}
}

func (p *Pool[T]) Get() T {
	if ptr := p.UnsafePool.Get(); ptr != nil {
		return *(*T)(ptr)
	}
	var t T
	if p.New != nil {
		t = p.New()
	}
	return t
}

func (p *Pool[T]) Put(t T) {
	if p.Reset != nil && !p.Reset(t) {
		return
	}
	ptr := unsafe.Pointer(&t)
	p.UnsafePool.Put(ptr)
}

// UnsafePool provides a form of UnsafeSimplePool
// with the addition of concurrency safety.
type UnsafePool struct {
	internal
	_ [cache_line_bytes - unsafe.Sizeof(internal{})%cache_line_bytes]byte
}

func NewUnsafePool(check func(current, victim int) bool) UnsafePool {
	return UnsafePool{internal: internal{
		pool: UnsafeSimplePool{Check: check},
	}}
}

const (
	// platform CPU cache line size to avoid false sharing.
	cache_line_bytes = unsafe.Sizeof(cpu.CacheLinePad{})
)

type internal struct {
	// underlying pool and
	// slow mutex protection.
	pool  UnsafeSimplePool
	mutex sync.Mutex

	// fast-access ring-buffer of
	// pointers accessible by PID
	// (running goroutine index).
	ring atomic_pointer
}

func (p *internal) Check(fn func(current, victim int) bool) func(current, victim int) bool {
	p.mutex.Lock()
	if fn == nil {
		if p.pool.Check == nil {
			fn = defaultCheck
		} else {
			fn = p.pool.Check
		}
	} else {
		p.pool.Check = fn
	}
	p.mutex.Unlock()
	return fn
}

func (p *internal) Get() unsafe.Pointer {
	pid := procPin()
	ptr := p.local(pid).Swap(nil)
	procUnpin()

	if ptr != nil {
		return ptr
	}

	p.mutex.Lock()
	ptr = p.pool.Get()
	p.mutex.Unlock()
	return ptr
}

func (p *internal) Put(ptr unsafe.Pointer) {
	pid := procPin()
	ptr = p.local(pid).Swap(ptr)
	procUnpin()

	if ptr == nil {
		return
	}

	p.mutex.Lock()
	p.pool.Put(ptr)
	p.mutex.Unlock()
}

func (p *internal) GC() {
	p.ring.Store(nil)
	p.mutex.Lock()
	p.pool.GC()
	p.mutex.Unlock()
}

func (p *internal) Size() (sz int) {
	if ptr := p.ring.Load(); ptr != nil {
		sz += len(*(*[]pointer_elem)(ptr))
	}
	p.mutex.Lock()
	sz += p.pool.Size()
	p.mutex.Unlock()
	return
}

// local returns an atomic_pointer from the fast-access
// ring buffer for the given goroutine PID index.
func (p *internal) local(pid uint) *pointer_elem {
	for {
		// Load current ring.
		ptr := p.ring.Load()
		if ptr != nil {

			// Check if pid within ring length.
			ring := *(*[]pointer_elem)(ptr)
			if pid < uint(len(ring)) {
				return &ring[pid]
			}
		}

		// Unpin before calling GOMAXPROCS,
		// which acquires a blocking mutex
		// lock on scheduler and may cause
		// goroutine to be rescheduled.
		procUnpin()

		// Allocate new ring buffer capable
		// of accomodating an index of 'pid'.
		ring := make([]pointer_elem, maxprocs())

		// Repin and get a (potentially)
		// different goroutine PID index.
		pid = procPin()

		// Attempt to replace
		// ring ptr with new.
		if pid < uint(len(ring)) &&
			p.ring.CAS(ptr, unsafe.Pointer(&ring)) {
			return &ring[pid]
		}
	}
}

// atomic_pointer wraps an unsafe.Pointer with
// receiver methods for their atomic counterparts.
type atomic_pointer struct{ p unsafe.Pointer }

func (p *atomic_pointer) Load() unsafe.Pointer {
	return atomic.LoadPointer(&p.p)
}

func (p *atomic_pointer) Store(ptr unsafe.Pointer) {
	atomic.StorePointer(&p.p, ptr)
}

func (p *atomic_pointer) CAS(old, new unsafe.Pointer) bool {
	return atomic.CompareAndSwapPointer(&p.p, old, new)
}

// pointer_elem wraps an unsafe.Pointer to make
// swapping of a slice element nicer on the eyes.
//
// THIS IS THE TRUE VIBE CODING, NONE OF THAT LLM
// DOG-ARSE BULLSHIT. WRITE CODE WITH *NICE VIBES*.
type pointer_elem struct{ p unsafe.Pointer }

func (p *pointer_elem) Swap(new unsafe.Pointer) unsafe.Pointer {
	old := p.p
	p.p = new
	return old
}

// maxprocs prevents runtime.GOMAXPROCS() from
// being inlined, making it more likely for its
// caller to be capable of being inlined.
//
//go:noinline
func maxprocs() int {
	return runtime.GOMAXPROCS(0)
}

// note is int in runtime, but should never be negative.
//
//go:linkname procPin runtime.procPin
func procPin() uint

//go:linkname procUnpin runtime.procUnpin
func procUnpin()
