package ftime

// A fast timing counter accessing the cheapest fastest
// counter your cpu can provide. Guaranteed to increase monotonically
// across successive calls on the same CPU core.
// On AMD64 CPUs we use the RDTSC instruction
func Counter() (count int64)

// Indicates whether the results returned from a call to Counter()
// increase at a uniform rate, independent of the actual clock speed
// of the CPU it is running on.
// On AMD64 CPUs we test for the 'Invariant TSC' property using CPUID
func IsCounterSteady() bool {
	_, _, _, edx := cpuid(0X80000007)
	return edx&(1<<8) != 0
}

// Indicates whether the results returned from a call to Counter()
// is guaranteed to be monotonically increasing per CPU and across
// multiple CPUs on the same socket. No guarantee is made across CPU
// sockets
// On AMD64 CPUs we test for the 'Invariant TSC' property using CPUID
func IsCounterSMPMonotonic() bool {
	_, _, _, edx := cpuid(0X80000007)
	return edx&(1<<8) != 0
}

func cpuid(eaxi uint32) (eax, ebx, ecx, edx uint32)

// This method will not return until the value returned by Counter()
// has increased by ticks.
// This method is useful as an alternative to time.Sleep() when very short
// pause periods are desired and it is undesirable to have the current
// thread/goroutine descheduled.
func Pause(ticks int64)
