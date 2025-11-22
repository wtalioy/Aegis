#include "vmlinux.h"
#include <bpf/bpf_core_read.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>

#define TASK_COMM_LEN 16

struct exec_event {
    u32 pid;
    u32 ppid;
    u64 cgroup_id;
    char comm[TASK_COMM_LEN];
    char pcomm[TASK_COMM_LEN];
};

struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 256 * 1024);
} events SEC(".maps");

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

    event->pid = pid;
    event->ppid = get_parent_pid(task);
    event->cgroup_id = bpf_get_current_cgroup_id();
    bpf_get_current_comm(&event->comm, sizeof(event->comm));

    // Get parent process name
    parent = BPF_CORE_READ(task, real_parent);
    if (parent) {
        BPF_CORE_READ_STR_INTO(&event->pcomm, parent, comm);
    } else {
        event->pcomm[0] = '\0';
    }

    bpf_ringbuf_submit(event, 0);
    return 0;
}

char LICENSE[] SEC("license") = "Dual BSD/GPL";
