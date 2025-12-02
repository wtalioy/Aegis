#include "vmlinux.h"
#include <bpf/bpf_core_read.h>
#include <bpf/bpf_endian.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>

#define TASK_COMM_LEN 16
#define PATH_MAX_LEN 256
#define NAME_MAX 128
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
    char filename[PATH_MAX_LEN];
    u8 blocked;
} __attribute__((packed));

struct file_open_event {
    u8 type;
    u32 pid;
    u64 cgroup_id;
    u32 flags;
    u64 ino;
    u64 dev;
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

// Key: "parent/filename" for 2-level matching (e.g., "etc/shadow")
struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 1024);
    __type(key, char[PATH_MAX_LEN]);
    __type(value, u8);
} monitored_files SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 1024);
    __type(key, u16);
    __type(value, u8);
} blocked_ports SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_LRU_HASH);
    __uint(max_entries, 32768);
    __type(key, u32);
    __type(value, u32);
} pid_to_ppid SEC(".maps");

struct path_scratch {
    char path_buf[PATH_MAX_LEN];
    char filename[NAME_MAX];
    char parent[NAME_MAX];
};

struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __uint(max_entries, 1);
    __type(key, u32);
    __type(value, struct path_scratch);
} scratch SEC(".maps");

static __always_inline u32 get_parent_pid(struct task_struct* task)
{
    if (!task)
        return 0;
    return BPF_CORE_READ(task, real_parent, tgid);
}

static __always_inline u8 check_file_action(struct dentry* dentry, char* out_path)
{
    if (!dentry)
        return 0;

    u32 key = 0;
    struct path_scratch* s = bpf_map_lookup_elem(&scratch, &key);
    if (!s)
        return 0;
    __builtin_memset(s, 0, sizeof(*s));

    struct qstr d_name = BPF_CORE_READ(dentry, d_name);
    if (!d_name.name || d_name.len == 0 || d_name.len >= NAME_MAX)
        return 0;
    bpf_probe_read_kernel_str(s->filename, NAME_MAX, d_name.name);

    struct dentry* parent_dentry = BPF_CORE_READ(dentry, d_parent);
    if (parent_dentry && parent_dentry != dentry) {
        struct qstr pd_name = BPF_CORE_READ(parent_dentry, d_name);
        if (pd_name.name && pd_name.len > 0 && pd_name.len < NAME_MAX) {
            bpf_probe_read_kernel_str(s->parent, NAME_MAX, pd_name.name);
        }
    }

    int pos = 0;
    if (s->parent[0]) {
        for (int i = 0; i < NAME_MAX - 1 && s->parent[i] && pos < PATH_MAX_LEN - 2; i++) {
            s->path_buf[pos++] = s->parent[i];
        }
        if (s->filename[0] && pos < PATH_MAX_LEN - 1) {
            s->path_buf[pos++] = '/';
        }
    }
    for (int i = 0; i < NAME_MAX - 1 && s->filename[i] && pos < PATH_MAX_LEN - 1; i++) {
        s->path_buf[pos++] = s->filename[i];
    }

    __builtin_memcpy(out_path, s->path_buf, PATH_MAX_LEN);
    u8* action = bpf_map_lookup_elem(&monitored_files, s->path_buf);
    if (action)
        return *action;

    if (s->parent[0]) {
        __builtin_memset(s->path_buf, 0, PATH_MAX_LEN);
        pos = 0;
        for (int i = 0; i < NAME_MAX - 1 && s->parent[i] && pos < PATH_MAX_LEN - 2; i++) {
            s->path_buf[pos++] = s->parent[i];
        }
        action = bpf_map_lookup_elem(&monitored_files, s->path_buf);
        if (action)
            return *action;
    }

    __builtin_memset(s->path_buf, 0, PATH_MAX_LEN);
    __builtin_memcpy(s->path_buf, s->filename, NAME_MAX);
    action = bpf_map_lookup_elem(&monitored_files, s->path_buf);
    if (action)
        return *action;

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

    u32 scratch_key = 0;
    struct path_scratch* s = bpf_map_lookup_elem(&scratch, &scratch_key);
    if (!s)
        return 0;

    __builtin_memset(s->path_buf, 0, PATH_MAX_LEN);

    struct file* file = BPF_CORE_READ(bprm, file);
    if (file) {
        struct dentry* dentry = BPF_CORE_READ(file, f_path.dentry);
        u8 action = check_file_action(dentry, s->path_buf);
        if (action == ACTION_BLOCK) {
            ret = -EPERM;
            blocked = 1;
        }
    }

    event = bpf_ringbuf_reserve(&events, sizeof(*event), 0);
    if (!event)
        return ret;

    event->type = EVENT_TYPE_EXEC;
    event->pid = pid;
    event->ppid = get_parent_pid(task);
    bpf_map_update_elem(&pid_to_ppid, &pid, &event->ppid, BPF_ANY);
    event->cgroup_id = bpf_get_current_cgroup_id();
    __builtin_memcpy(event->filename, s->path_buf, PATH_MAX_LEN);
    bpf_get_current_comm(&event->comm, sizeof(event->comm));

    parent = BPF_CORE_READ(task, real_parent);
    if (parent) {
        BPF_CORE_READ_STR_INTO(&event->pcomm, parent, comm);
    } else {
        event->pcomm[0] = '\0';
    }

    event->blocked = blocked;
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

    u32 scratch_key = 0;
    struct path_scratch* s = bpf_map_lookup_elem(&scratch, &scratch_key);
    if (!s)
        return 0;

    __builtin_memset(s->path_buf, 0, PATH_MAX_LEN);

    struct dentry* dentry = BPF_CORE_READ(file, f_path.dentry);
    u8 action = check_file_action(dentry, s->path_buf);
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
    event->ino = 0;
    event->dev = 0;
    if (file) {
        struct inode* inode = BPF_CORE_READ(file, f_inode);
        if (inode) {
            event->ino = BPF_CORE_READ(inode, i_ino);
            struct super_block* sb = BPF_CORE_READ(inode, i_sb);
            if (sb) {
                event->dev = BPF_CORE_READ(sb, s_dev);
            }
        }
    }
    __builtin_memcpy(event->filename, s->path_buf, PATH_MAX_LEN);
    event->blocked = blocked;
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
    event->addr_v4 = 0;
    __builtin_memset(event->addr_v6, 0, 16);

    if (family == AF_INET) {
        struct sockaddr_in* addr_in = (struct sockaddr_in*)address;
        bpf_probe_read_kernel(&event->addr_v4, sizeof(event->addr_v4), &addr_in->sin_addr.s_addr);
    } else if (family == AF_INET6) {
        struct sockaddr_in6* addr_in6 = (struct sockaddr_in6*)address;
        bpf_probe_read_kernel(event->addr_v6, 16, &addr_in6->sin6_addr);
    }

    event->blocked = blocked;
    bpf_ringbuf_submit(event, 0);
    return ret;
}

char LICENSE[] SEC("license") = "GPL";
