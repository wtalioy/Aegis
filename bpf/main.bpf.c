#include "vmlinux.h"
#include <bpf/bpf_core_read.h>
#include <bpf/bpf_endian.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>

#define TASK_COMM_LEN 16
#define PATH_MAX_LEN 256
#define COMMAND_LINE_LEN 512
#define NAME_MAX 128
#define MAX_ARGC_FOR_CMD 16
#define MAX_ARGS_TO_READ 4
#define ARGV0_READ_LEN 256
#define CMD_LINE_SAFETY_MARGIN 64
#define EVENT_TYPE_EXEC 1
#define EVENT_TYPE_FILE_OPEN 2
#define EVENT_TYPE_CONNECT 3

#define EPERM 1
#define AF_INET 2
#define AF_INET6 10

#define ACTION_MONITOR 1
#define ACTION_BLOCK 2

struct event_header {
    u64 timestamp_ns;
    u64 cgroup_id;
    u32 pid;
    u32 tid;
    u32 uid;
    u32 gid;
    u8  type;
    u8  blocked;
    u8  _pad[6];
    char comm[TASK_COMM_LEN];
} __attribute__((packed));

struct exec_event {
    struct event_header hdr;
    u32 ppid;
    u8  _pad[4];
    char pcomm[TASK_COMM_LEN];
    char filename[PATH_MAX_LEN];
    char command_line[COMMAND_LINE_LEN];
} __attribute__((packed));

struct file_event {
    struct event_header hdr;
    u64 ino;
    u64 dev;
    u32 flags;
    u8  _pad[4];
    char filename[PATH_MAX_LEN];
} __attribute__((packed));

struct connect_event {
    struct event_header hdr;
    u32 addr_v4;
    u16 family;
    u16 port;
    u8  addr_v6[16];
} __attribute__((packed));

struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 2 * 1024 * 1024);
} events SEC(".maps");

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

struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __uint(max_entries, 1);
    __type(key, u32);
    __type(value, struct exec_event);
} event_scratch SEC(".maps");

static __always_inline void fill_event_header(
    struct event_header *hdr,
    u8 type,
    struct task_struct *task
) {
    hdr->timestamp_ns = bpf_ktime_get_ns();
    hdr->type = type;
    hdr->blocked = 0;
    
    u64 pid_tgid = bpf_get_current_pid_tgid();
    hdr->pid = pid_tgid >> 32;
    hdr->tid = (u32)pid_tgid;
    
    u64 uid_gid = bpf_get_current_uid_gid();
    hdr->uid = (u32)uid_gid;
    hdr->gid = uid_gid >> 32;
    
    hdr->cgroup_id = bpf_get_current_cgroup_id();
    bpf_get_current_comm(&hdr->comm, sizeof(hdr->comm));
}

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

    struct exec_event* scratch_event = bpf_map_lookup_elem(&event_scratch, &scratch_key);
    if (!scratch_event)
        return ret;

    event = bpf_ringbuf_reserve(&events, sizeof(*event), 0);
    if (!event)
        return ret;

    fill_event_header(&event->hdr, EVENT_TYPE_EXEC, task);
    event->hdr.blocked = blocked;

    parent = BPF_CORE_READ(task, real_parent);
    event->ppid = get_parent_pid(task);
    bpf_map_update_elem(&pid_to_ppid, &pid, &event->ppid, BPF_ANY);
    
    if (parent) {
        BPF_CORE_READ_STR_INTO(&event->pcomm, parent, comm);
    } else {
        event->pcomm[0] = '\0';
    }

    __builtin_memcpy(event->filename, s->path_buf, PATH_MAX_LEN);
    __builtin_memcpy(event->command_line, s->path_buf, PATH_MAX_LEN);
    
    u32 argc = 0;
    if (bpf_probe_read_kernel(&argc, sizeof(argc), (void*)BPF_CORE_READ(bprm, p))) {
        goto submit;
    }
    
    if (argc == 0 || argc > MAX_ARGC_FOR_CMD) {
        goto submit;
    }
    
    u64 argv_start = (u64)BPF_CORE_READ(bprm, p) + sizeof(u32);
    u64 argv0_ptr = 0;
    if (bpf_probe_read_kernel(&argv0_ptr, sizeof(argv0_ptr), (void*)argv_start)) {
        goto submit;
    }
    if (argv0_ptr == 0) {
        goto submit;
    }
    
    u32 argv0_read_size = ARGV0_READ_LEN;
    if (argv0_read_size > COMMAND_LINE_LEN) {
        argv0_read_size = COMMAND_LINE_LEN;
    }
    long n = bpf_probe_read_user_str(event->command_line, argv0_read_size, (void*)argv0_ptr);
    if (n > 0 && n < argv0_read_size) {
        u32 pos = (u32)(n - 1); // n includes null terminator, n > 0 so n-1 >= 0
        if (pos >= COMMAND_LINE_LEN) {
            goto submit;
        }
        u32 max_args = argc < MAX_ARGS_TO_READ ? argc : MAX_ARGS_TO_READ;
        
        for (u32 i = 1; i < max_args && pos < (COMMAND_LINE_LEN - CMD_LINE_SAFETY_MARGIN); i++) {
            u64 argv_i_addr = argv_start + (i * sizeof(u64));
            u64 argv_i_ptr = 0;
            if (bpf_probe_read_kernel(&argv_i_ptr, sizeof(argv_i_ptr), (void*)argv_i_addr)) {
                break;
            }
            if (argv_i_ptr == 0) {
                break;
            }
            
            if (pos >= COMMAND_LINE_LEN - 1) {
                break;
            }
            event->command_line[pos++] = ' ';
            
            if (pos >= COMMAND_LINE_LEN) {
                break;
            }
            u32 remaining;
            if (pos < COMMAND_LINE_LEN) {
                remaining = COMMAND_LINE_LEN - pos;
            } else {
                break;
            }
            if (remaining == 0) {
                break;
            }
            u32 read_size;
            if (remaining > CMD_LINE_SAFETY_MARGIN) {
                read_size = CMD_LINE_SAFETY_MARGIN;
            } else {
                read_size = remaining;
            }
            if (read_size == 0) {
                break;
            }
            if (read_size > remaining) {
                break;
            }
            if (pos + read_size > COMMAND_LINE_LEN) {
                break;
            }
            long arg_len = bpf_probe_read_user_str(&event->command_line[pos], read_size, (void*)argv_i_ptr);
            if (arg_len <= 0) {
                break;
            }
            if ((u32)arg_len > remaining) {
                break;
            }
            pos += (u32)(arg_len - 1);
            if (pos >= COMMAND_LINE_LEN - 1) {
                break;
            }
        }
        if (pos < COMMAND_LINE_LEN) {
            event->command_line[pos] = '\0';
        }
    }

submit:
    bpf_ringbuf_submit(event, 0);
    return ret;
}

SEC("lsm/file_open")
int BPF_PROG(lsm_file_open, struct file* file)
{
    struct file_event* event;
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

    struct task_struct* task = (struct task_struct*)bpf_get_current_task_btf();
    fill_event_header(&event->hdr, EVENT_TYPE_FILE_OPEN, task);
    event->hdr.blocked = blocked;

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

    struct task_struct* task = (struct task_struct*)bpf_get_current_task_btf();
    fill_event_header(&event->hdr, EVENT_TYPE_CONNECT, task);
    event->hdr.blocked = blocked;

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

    bpf_ringbuf_submit(event, 0);
    return ret;
}

char LICENSE[] SEC("license") = "GPL";
