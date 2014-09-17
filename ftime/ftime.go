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
	_, _, ecx, _ := cpuid(0X80000007)
	return ecx&(1<<8) != 0
}

// Indicates whether the results returned from a call to Counter()
// is guaranteed to be monotonically increasing per CPU and across
// multiple CPUs on the same socket. No guarantee is made across CPU
// sockets
// On AMD64 CPUs we test for the 'Invariant TSC' property using CPUID
func IsCounterSMPMonotonic() bool {
	_, _, ecx, _ := cpuid(0X80000007)
	return ecx&(1<<8) != 0
}

func cpuid(eaxi uint32) (eax, ebx, ecx, edx uint32)
