GO_BUILD_CMD = go build -o ./build/eulerguard ./cmd/eulerguard

BPF_C_SRC = ./bpf/eulerguard.bpf.c
BPF_O_OBJ = ./bpf/eulerguard.bpf.o

BPF_CFLAGS = -g -O2 -target bpf -c

all: bpf go

bpf:
	@echo "==> Build eBPF (C)..."
	@clang $(BPF_CFLAGS) -o $(BPF_O_OBJ) $(BPF_C_SRC)

go:
	@echo "==> Build Go..."
	@$(GO_BUILD_CMD)

clean:
	@echo "==> Clean..."
	@rm -f $(BPF_O_OBJ) ./build/eulerguard

.PHONY: all bpf go clean