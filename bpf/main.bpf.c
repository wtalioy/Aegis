#include "vmlinux.h"
#include <bpf/bpf_core_read.h>
#include <bpf/bpf_endian.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>

#define TASK_COMM_LEN 16
#define PATH_MAX_LEN 256
#define MAX_PATH_DEPTH 16
#define MAX_NAME_LEN 48
#define EVENT_TYPE_EXEC 1
#define EVENT_TYPE_FILE_OPEN 2
#define EVENT_TYPE_CONNECT 3

#define EPERM 1
#define AF_INET 2
#define AF_INET6 10

#define ACTION_MONITOR 1
#define ACTION_BLOCK 2

struct exec_event {
    u8 type;
    u32 pid;
    u32 ppid;
    u64 cgroup_id;
    char comm[TASK_COMM_LEN];
    char pcomm[TASK_COMM_LEN];
    u8 blocked;
} __attribute__((packed));

struct file_open_event {
    u8 type;
    u32 pid;
    u64 cgroup_id;
    u32 flags;
    char filename[PATH_MAX_LEN];
    u8 blocked;
} __attribute__((packed));

struct connect_event {
    u8 type;
    u32 pid;
    u64 cgroup_id;
    u16 family;
    u16 port;
    u32 addr_v4;
    u8 addr_v6[16];
    u8 blocked;
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
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 1024);
    __type(key, u16);
    __type(value, u8);
} blocked_ports SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __uint(max_entries, 1);
    __type(key, u32);
    __type(value, char[PATH_MAX_LEN]);
} path_buffer SEC(".maps");

struct path_build_ctx {
    char names[MAX_PATH_DEPTH][MAX_NAME_LEN];
    u32 lens[MAX_PATH_DEPTH];
    int count;
};

struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __uint(max_entries, 1);
    __type(key, u32);
    __type(value, struct path_build_ctx);
} path_ctx SEC(".maps");

static __always_inline u32 get_parent_pid(struct task_struct* task)
{
    if (!task)
        return 0;
    return BPF_CORE_READ(task, real_parent, tgid);
}

static __always_inline u8 get_path_action(struct dentry* dentry, char* path_buf)
{
    u32 ctx_key = 0;
    struct path_build_ctx* ctx = bpf_map_lookup_elem(&path_ctx, &ctx_key);
    if (!ctx)
        return 0;

    ctx->count = 0;
    struct dentry* d = dentry;

    // Walk up dentry tree, collect path components
    for (int i = 0; i < MAX_PATH_DEPTH && d; i++) {
        struct dentry* parent = BPF_CORE_READ(d, d_parent);
        if (parent == d)
            break;

        struct qstr d_name;
        bpf_probe_read_kernel(&d_name, sizeof(d_name), &d->d_name);

        u32 len = d_name.len;
        if (len > 0 && len < MAX_NAME_LEN) {
            bpf_probe_read_kernel_str(ctx->names[i], MAX_NAME_LEN, d_name.name);
            ctx->lens[i] = len;
            ctx->count = i + 1;
        }
        d = parent;
    }

    if (ctx->count == 0)
        return 0;

    // Build path: iterate from root (high index) to leaf (index 0)
    __builtin_memset(path_buf, 0, PATH_MAX_LEN);
    int pos = 0;
    int cnt = ctx->count;

    for (int i = MAX_PATH_DEPTH - 1; i >= 0; i--) {
        if (i >= cnt)
            continue;
        if (pos >= PATH_MAX_LEN - 2)
            break;

        path_buf[pos++] = '/';

        u32 len = ctx->lens[i];
        if (len > MAX_NAME_LEN - 1)
            len = MAX_NAME_LEN - 1;

        for (u32 j = 0; j < MAX_NAME_LEN - 1 && j < len && pos < PATH_MAX_LEN - 1; j++) {
            path_buf[pos++] = ctx->names[i][j];
        }
    }

    // look up full path
    u8* val = bpf_map_lookup_elem(&monitored_paths, path_buf);
    if (val)
        return *val;

    // Fallback: try just the filename (leaf = index 0)
    if (ctx->count > 0) {
        char basename[PATH_MAX_LEN] = {};
        u32 len = ctx->lens[0];
        if (len > MAX_NAME_LEN - 1)
            len = MAX_NAME_LEN - 1;
        for (u32 j = 0; j < len && j < MAX_NAME_LEN - 1; j++) {
            basename[j] = ctx->names[0][j];
        }
        val = bpf_map_lookup_elem(&monitored_paths, basename);
        if (val)
            return *val;
    }

    return 0;
}

SEC("lsm/bprm_check_security")
int BPF_PROG(lsm_bprm_check, struct linux_binprm* bprm)
{
    struct exec_event* event;
    struct task_struct* task = (struct task_struct*)bpf_get_current_task_btf();
    struct task_struct* parent;
    u64 pid_tgid = bpf_get_current_pid_tgid();
    u32 pid = pid_tgid >> 32;
    int ret = 0;
    u8 blocked = 0;

    struct file* file = BPF_CORE_READ(bprm, file);
    if (file) {
        struct dentry* dentry = BPF_CORE_READ(file, f_path.dentry);
        if (dentry) {
            u32 key = 0;
            char* path_buf = bpf_map_lookup_elem(&path_buffer, &key);
            if (path_buf) {
                u8 action = get_path_action(dentry, path_buf);
                if (action == ACTION_BLOCK) {
                    ret = -EPERM;
                    blocked = 1;
                }
            }
        }
    }

    event = bpf_ringbuf_reserve(&events, sizeof(*event), 0);
    if (!event)
        return ret;

    event->type = EVENT_TYPE_EXEC;
    event->pid = pid;
    event->ppid = get_parent_pid(task);
    event->cgroup_id = bpf_get_current_cgroup_id();
    event->blocked = blocked;
    bpf_get_current_comm(&event->comm, sizeof(event->comm));

    parent = BPF_CORE_READ(task, real_parent);
    if (parent) {
        BPF_CORE_READ_STR_INTO(&event->pcomm, parent, comm);
    } else {
        event->pcomm[0] = '\0';
    }

    bpf_ringbuf_submit(event, 0);
    return ret;
}

SEC("lsm/file_open")
int BPF_PROG(lsm_file_open, struct file* file)
{
    struct file_open_event* event;
    u64 pid_tgid = bpf_get_current_pid_tgid();
    u32 pid = pid_tgid >> 32;
    int ret = 0;
    u8 blocked = 0;

    struct dentry* dentry = BPF_CORE_READ(file, f_path.dentry);
    if (!dentry)
        return 0;

    u32 key = 0;
    char* path_buf = bpf_map_lookup_elem(&path_buffer, &key);
    if (!path_buf)
        return 0;

    u8 action = get_path_action(dentry, path_buf);
    if (!action)
        return 0;

    if (action == ACTION_BLOCK) {
        ret = -EPERM;
        blocked = 1;
    }

    event = bpf_ringbuf_reserve(&events, sizeof(*event), 0);
    if (!event)
        return ret;

    event->type = EVENT_TYPE_FILE_OPEN;
    event->pid = pid;
    event->cgroup_id = bpf_get_current_cgroup_id();
    event->flags = BPF_CORE_READ(file, f_flags);
    event->blocked = blocked;
    __builtin_memcpy(event->filename, path_buf, PATH_MAX_LEN);
    bpf_ringbuf_submit(event, 0);

    return ret;
}

SEC("lsm/socket_connect")
int BPF_PROG(lsm_socket_connect, struct socket* sock, struct sockaddr* address, int addrlen)
{
    struct connect_event* event;
    u64 pid_tgid = bpf_get_current_pid_tgid();
    u32 pid = pid_tgid >> 32;
    int ret = 0;
    u8 blocked = 0;
    u16 port = 0;
    u16 family = 0;

    if (!address)
        return 0;

    bpf_probe_read_kernel(&family, sizeof(family), &address->sa_family);

    if (family == AF_INET) {
        struct sockaddr_in* addr_in = (struct sockaddr_in*)address;
        u16 port_net = 0;
        bpf_probe_read_kernel(&port_net, sizeof(port_net), &addr_in->sin_port);
        port = __bpf_ntohs(port_net);
    } else if (family == AF_INET6) {
        struct sockaddr_in6* addr_in6 = (struct sockaddr_in6*)address;
        u16 port_net = 0;
        bpf_probe_read_kernel(&port_net, sizeof(port_net), &addr_in6->sin6_port);
        port = __bpf_ntohs(port_net);
    } else {
        return 0;
    }

    u8* port_action = bpf_map_lookup_elem(&blocked_ports, &port);
    if (!port_action)
        return 0;

    if (*port_action == ACTION_BLOCK) {
        ret = -EPERM;
        blocked = 1;
    }

    event = bpf_ringbuf_reserve(&events, sizeof(*event), 0);
    if (!event)
        return ret;

    event->type = EVENT_TYPE_CONNECT;
    event->pid = pid;
    event->cgroup_id = bpf_get_current_cgroup_id();
    event->family = family;
    event->port = port;
    event->blocked = blocked;
    event->addr_v4 = 0;
    __builtin_memset(event->addr_v6, 0, 16);

    if (family == AF_INET) {
        struct sockaddr_in* addr_in = (struct sockaddr_in*)address;
        bpf_probe_read_kernel(&event->addr_v4, sizeof(event->addr_v4), &addr_in->sin_addr.s_addr);
    } else if (family == AF_INET6) {
        struct sockaddr_in6* addr_in6 = (struct sockaddr_in6*)address;
        bpf_probe_read_kernel(event->addr_v6, 16, &addr_in6->sin6_addr);
    }

    bpf_ringbuf_submit(event, 0);
    return ret;
}

char LICENSE[] SEC("license") = "GPL";
