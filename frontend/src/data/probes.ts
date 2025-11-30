// Static probe dictionary for KernelXRay educational content

export interface ProbeInfo {
  id: string
  name: string
  hook: string
  hookType: 'lsm' | 'tracepoint'
  description: string
  capability: 'monitor' | 'block'
  sourceCode: string
  kernelStructs: string[]
  category: 'process' | 'file' | 'network'
}

export const probes: ProbeInfo[] = [
  {
    id: 'exec',
    name: 'Process Execution',
    hook: 'lsm/bprm_check_security',
    hookType: 'lsm',
    capability: 'block',
    description: 'LSM hook that intercepts all process execution via execve(). Can actively BLOCK malicious binaries from executing or ALERT on suspicious execution patterns. Captures PID, parent process, command name, and cgroup ID for container detection.',
    sourceCode: `SEC("lsm/bprm_check_security")
int BPF_PROG(lsm_bprm_check, struct linux_binprm* bprm, int ret) {
    struct exec_event* event;
    struct task_struct* task = (struct task_struct*)bpf_get_current_task_btf();
    
    // Check if this binary should be blocked
    struct file* file = bprm->file;
    struct dentry* dentry = BPF_CORE_READ(file, f_path.dentry);
    u8 action = check_file_action(dentry, event->filename);
    
    event = bpf_ringbuf_reserve(&events, sizeof(*event), 0);
    if (!event) return 0;
    
    event->type = EVENT_TYPE_EXEC;
    event->pid = bpf_get_current_pid_tgid() >> 32;
    event->ppid = BPF_CORE_READ(task, real_parent, tgid);
    event->blocked = (action == ACTION_BLOCK) ? 1 : 0;
    
    bpf_ringbuf_submit(event, 0);
    
    // Return -EPERM to block execution
    return (action == ACTION_BLOCK) ? -EPERM : 0;
}`,
    kernelStructs: [
      'linux_binprm.file',
      'task_struct.pid',
      'task_struct.real_parent',
      'dentry.d_name'
    ],
    category: 'process'
  },
  {
    id: 'openat',
    name: 'File Access',
    hook: 'lsm/file_open',
    hookType: 'lsm',
    capability: 'block',
    description: 'LSM hook that intercepts all file open operations. Can actively BLOCK access to sensitive files like /etc/shadow or ALERT on suspicious file access patterns. Uses parent/filename matching for efficient kernel-side rule enforcement.',
    sourceCode: `SEC("lsm/file_open")
int BPF_PROG(lsm_file_open, struct file* file, int ret) {
    struct file_open_event* event;
    
    struct dentry* dentry = BPF_CORE_READ(file, f_path.dentry);
    char path_buf[PATH_MAX_LEN];
    u8 action = check_file_action(dentry, path_buf);
    
    // Only emit event if monitored or blocked
    if (action == 0) return 0;
    
    event = bpf_ringbuf_reserve(&events, sizeof(*event), 0);
    if (!event) return (action == ACTION_BLOCK) ? -EPERM : 0;
    
    event->type = EVENT_TYPE_FILE;
    event->pid = bpf_get_current_pid_tgid() >> 32;
    event->blocked = (action == ACTION_BLOCK) ? 1 : 0;
    __builtin_memcpy(event->filename, path_buf, PATH_MAX_LEN);
    
    bpf_ringbuf_submit(event, 0);
    
    // Return -EPERM to deny file access
    return (action == ACTION_BLOCK) ? -EPERM : 0;
}`,
    kernelStructs: [
      'file.f_path.dentry',
      'dentry.d_parent',
      'dentry.d_name.name',
      'qstr.len'
    ],
    category: 'file'
  },
  {
    id: 'connect',
    name: 'Network Connection',
    hook: 'lsm/socket_connect',
    hookType: 'lsm',
    capability: 'block',
    description: 'LSM hook that intercepts outbound network connections. Can actively BLOCK connections to malicious ports (e.g., reverse shells on 4444) or ALERT on suspicious network activity. Critical for preventing C2 callbacks and data exfiltration.',
    sourceCode: `SEC("lsm/socket_connect")
int BPF_PROG(lsm_socket_connect, struct socket* sock, 
             struct sockaddr* address, int addrlen, int ret) {
    struct connect_event* event;
    
    u16 family = address->sa_family;
    if (family != AF_INET && family != AF_INET6) return 0;
    
    u16 port = 0;
    if (family == AF_INET) {
        struct sockaddr_in* addr4 = (struct sockaddr_in*)address;
        port = __builtin_bswap16(addr4->sin_port);
    }
    
    // Check if port should be blocked
    u8* action = bpf_map_lookup_elem(&blocked_ports, &port);
    u8 should_block = action && (*action == ACTION_BLOCK);
    
    event = bpf_ringbuf_reserve(&events, sizeof(*event), 0);
    if (!event) return should_block ? -EPERM : 0;
    
    event->type = EVENT_TYPE_CONNECT;
    event->port = port;
    event->blocked = should_block ? 1 : 0;
    
    bpf_ringbuf_submit(event, 0);
    
    // Return -EPERM to block the connection
    return should_block ? -EPERM : 0;
}`,
    kernelStructs: [
      'socket.sk',
      'sockaddr.sa_family',
      'sockaddr_in.sin_port',
      'sockaddr_in6.sin6_port'
    ],
    category: 'network'
  }
]

// Kernel structure information for educational display
export interface KernelStruct {
  name: string
  description: string
  fields: { name: string; type: string; description: string }[]
}

export const kernelStructs: KernelStruct[] = [
  {
    name: 'linux_binprm',
    description: 'Binary program descriptor used during execve(). Contains all information needed to execute a new program.',
    fields: [
      { name: 'file', type: 'struct file*', description: 'File being executed' },
      { name: 'filename', type: 'const char*', description: 'Name of file to execute' },
      { name: 'interp', type: 'const char*', description: 'Interpreter name (for scripts)' },
      { name: 'cred', type: 'struct cred*', description: 'Credentials for the new process' }
    ]
  },
  {
    name: 'dentry',
    description: 'Directory entry - represents a path component in the filesystem hierarchy.',
    fields: [
      { name: 'd_name', type: 'struct qstr', description: 'Filename component' },
      { name: 'd_parent', type: 'struct dentry*', description: 'Parent directory entry' },
      { name: 'd_inode', type: 'struct inode*', description: 'Associated inode' }
    ]
  },
  {
    name: 'task_struct',
    description: 'The fundamental process descriptor in Linux. Contains all information about a process/thread.',
    fields: [
      { name: 'pid', type: 'pid_t', description: 'Process ID (thread ID)' },
      { name: 'tgid', type: 'pid_t', description: 'Thread Group ID (process ID)' },
      { name: 'real_parent', type: 'struct task_struct*', description: 'Pointer to parent process' },
      { name: 'comm[16]', type: 'char[16]', description: 'Executable name (max 16 chars)' }
    ]
  },
  {
    name: 'sockaddr_in',
    description: 'IPv4 socket address structure used in network syscalls.',
    fields: [
      { name: 'sin_family', type: 'sa_family_t', description: 'Address family (AF_INET)' },
      { name: 'sin_port', type: '__be16', description: 'Port number (network byte order)' },
      { name: 'sin_addr', type: 'struct in_addr', description: 'IPv4 address' }
    ]
  }
]
