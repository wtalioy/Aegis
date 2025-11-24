#include "vmlinux.h"
#include <bpf/bpf_core_read.h>
#include <bpf/bpf_endian.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>

#define TASK_COMM_LEN 16
#define PATH_MAX_LEN 256
#define EVENT_TYPE_EXEC 1
#define EVENT_TYPE_FILE_OPEN 2
#define EVENT_TYPE_CONNECT 3

struct exec_event {
    u8 type;
    u32 pid;
    u32 ppid;
    u64 cgroup_id;
    char comm[TASK_COMM_LEN];
    char pcomm[TASK_COMM_LEN];
} __attribute__((packed));

struct file_open_event {
    u8 type;
    u32 pid;
    u64 cgroup_id;
    u32 flags;
    char filename[PATH_MAX_LEN];
} __attribute__((packed));

struct connect_event {
    u8 type;
    u32 pid;
    u64 cgroup_id;
    u16 family;
    u16 port;
    u32 addr_v4;
    u8 addr_v6[16];
} __attribute__((packed));

struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 256 * 1024);
} events SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 1024);
    __type(key, char[PATH_MAX_LEN]);
    __type(value, u8);
} monitored_paths SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __uint(max_entries, 1);
    __type(key, u32);
    __type(value, char[PATH_MAX_LEN]);
} path_buffer SEC(".maps");

static __always_inline u32 get_parent_pid(struct task_struct* task)
{
    if (!task)
        return 0;

    return BPF_CORE_READ(task, real_parent, tgid);
}

SEC("tp/sched/sched_process_exec")
int handle_exec(struct trace_event_raw_sched_process_exec* ctx)
{
    struct exec_event* event;
    struct task_struct* task = (struct task_struct*)bpf_get_current_task_btf();
    struct task_struct* parent;
    u64 pid_tgid = bpf_get_current_pid_tgid();
    u32 pid = pid_tgid >> 32;

    event = bpf_ringbuf_reserve(&events, sizeof(*event), 0);
    if (!event)
        return 0;

    event->type = EVENT_TYPE_EXEC;
    event->pid = pid;
    event->ppid = get_parent_pid(task);
    event->cgroup_id = bpf_get_current_cgroup_id();
    bpf_get_current_comm(&event->comm, sizeof(event->comm));

    parent = BPF_CORE_READ(task, real_parent);
    if (parent) {
        BPF_CORE_READ_STR_INTO(&event->pcomm, parent, comm);
    } else {
        event->pcomm[0] = '\0';
    }

    bpf_ringbuf_submit(event, 0);
    return 0;
}

static __always_inline bool is_monitored_path(const char* userspace_path, char* path_buf)
{
    __builtin_memset(path_buf, 0, PATH_MAX_LEN);
    long ret = bpf_probe_read_user_str(path_buf, PATH_MAX_LEN, userspace_path);
    if (ret <= 0)
        return false;

    // Try exact match first (for specific files like /etc/passwd)
    u8* val = bpf_map_lookup_elem(&monitored_paths, path_buf);
    if (val)
        return true;

// Try prefix matches by temporarily null-terminating at each '/'
// For directory prefixes like /home/ or /etc/
#pragma unroll
    for (int i = 1; i < PATH_MAX_LEN - 1; i++) {
        if (path_buf[i] == '/') {
            char saved = path_buf[i + 1];
            path_buf[i + 1] = '\0';
            val = bpf_map_lookup_elem(&monitored_paths, path_buf);
            path_buf[i + 1] = saved;
            if (val)
                return true;
        }
        if (path_buf[i] == '\0')
            break;
    }

    return false;
}

SEC("tp/syscalls/sys_enter_openat")
int tracepoint_openat(struct trace_event_raw_sys_enter* ctx)
{
    struct file_open_event* event;
    u64 pid_tgid = bpf_get_current_pid_tgid();
    u32 pid = pid_tgid >> 32;

    const char* filename = (const char*)ctx->args[1];

    u32 key = 0;
    char* path_buf = bpf_map_lookup_elem(&path_buffer, &key);
    if (!path_buf)
        return 0;
    if (!is_monitored_path(filename, path_buf))
        return 0;

    event = bpf_ringbuf_reserve(&events, sizeof(*event), 0);
    if (!event)
        return 0;

    event->type = EVENT_TYPE_FILE_OPEN;
    event->pid = pid;
    event->cgroup_id = bpf_get_current_cgroup_id();
    event->flags = (u32)ctx->args[2];

    __builtin_memcpy(event->filename, path_buf, PATH_MAX_LEN);

    bpf_ringbuf_submit(event, 0);
    return 0;
}

SEC("tp/syscalls/sys_enter_connect")
int tracepoint_connect(struct trace_event_raw_sys_enter* ctx)
{
    struct connect_event* event;
    u64 pid_tgid = bpf_get_current_pid_tgid();
    u32 pid = pid_tgid >> 32;

    struct sockaddr* addr = (struct sockaddr*)ctx->args[1];
    if (!addr)
        return 0;

    event = bpf_ringbuf_reserve(&events, sizeof(*event), 0);
    if (!event)
        return 0;

    event->type = EVENT_TYPE_CONNECT;
    event->pid = pid;
    event->cgroup_id = bpf_get_current_cgroup_id();
    event->family = 0;
    event->port = 0;
    event->addr_v4 = 0;
    __builtin_memset(event->addr_v6, 0, 16);

    u16 sa_family = 0;
    long ret = bpf_probe_read_user(&sa_family, sizeof(sa_family), &addr->sa_family);
    if (ret < 0) {
        bpf_ringbuf_discard(event, 0);
        return 0;
    }
    event->family = sa_family;

    // Handle IPv4 (AF_INET = 2)
    if (sa_family == 2) {
        struct sockaddr_in* addr_in = (struct sockaddr_in*)addr;
        u16 port_net = 0;
        u32 addr_net = 0;

        bpf_probe_read_user(&port_net, sizeof(port_net), &addr_in->sin_port);
        bpf_probe_read_user(&addr_net, sizeof(addr_net), &addr_in->sin_addr.s_addr);

        event->port = __bpf_ntohs(port_net);
        event->addr_v4 = addr_net;
    }
    // Handle IPv6 (AF_INET6 = 10)
    else if (sa_family == 10) {
        struct sockaddr_in6* addr_in6 = (struct sockaddr_in6*)addr;
        u16 port_net = 0;

        bpf_probe_read_user(&port_net, sizeof(port_net), &addr_in6->sin6_port);
        bpf_probe_read_user(event->addr_v6, 16, &addr_in6->sin6_addr);

        event->port = __bpf_ntohs(port_net);
    }

    bpf_ringbuf_submit(event, 0);
    return 0;
}

char LICENSE[] SEC("license") = "Dual BSD/GPL";
