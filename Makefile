BPF_C_SRC = ./bpf/main.bpf.c
BPF_O_OBJ = ./bpf/main.bpf.o

VMLINUX_H = ./bpf/vmlinux.h

BPF_CFLAGS = -g -O2 -target bpf -c
BPF_CFLAGS += -I./bpf

all: bpf

bpf: $(VMLINUX_H)
	@echo "==> Build eBPF (C)..."
	@clang $(BPF_CFLAGS) -o $(BPF_O_OBJ) $(BPF_C_SRC)

$(VMLINUX_H):
	@echo "==> [Dependency missing] Generating vmlinux.h from kernel BTF ..."
	@echo "    (This may take a few seconds, but will only be executed once)"
	@bpftool btf dump file /sys/kernel/btf/vmlinux format c > $(VMLINUX_H)

GO_BUILD_CMD = go build -o ./build/eulerguard ./cmd/eulerguard
go:
	@echo "==> Build Go..."
	@$(GO_BUILD_CMD)

clean:
	@echo "==> Clean..."
	@rm -f $(BPF_O_OBJ) ./build/eulerguard $(VMLINUX_H)

.PHONY: all bpf go clean