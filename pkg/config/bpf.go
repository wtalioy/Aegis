package config

type BPFOptions struct {
	RingBufferSize    int  `yaml:"ring_buffer_size"`    // Default: 2MB
	ProcessCacheSize  int  `yaml:"process_cache_size"`  // Default: 16384
	EnableArgv        bool `yaml:"enable_argv"`         // Whether to collect argv
	BatchSize         int  `yaml:"batch_size"`          // Batch processing size
	SkipKernelThreads bool `yaml:"skip_kernel_threads"` // Skip kernel threads
}

var DefaultBPFOptions = BPFOptions{
	RingBufferSize:    2 * 1024 * 1024, // 2MB
	ProcessCacheSize:  16384,
	EnableArgv:        true,
	BatchSize:         100,
	SkipKernelThreads: true,
}
