# EulerGuard Build System
BPF_SRC = ./bpf/main.bpf.c
BPF_OBJ = ./bpf/main.bpf.o
VMLINUX = ./bpf/vmlinux.h
BUILD   = ./build

all: cli

# eBPF
bpf: $(VMLINUX)
	@echo "==> Building eBPF..."
	@clang -g -O2 -target bpf -c -I./bpf -o $(BPF_OBJ) $(BPF_SRC)

$(VMLINUX):
	@echo "==> Generating vmlinux.h..."
	@bpftool btf dump file /sys/kernel/btf/vmlinux format c > $(VMLINUX)

# Frontend
frontend:
	@echo "==> Building frontend..."
	@cd frontend && npm install && npm run build

# Builds
cli: bpf
	@echo "==> Building CLI..."
	@mkdir -p $(BUILD)
	@go build -o $(BUILD)/eulerguard ./cmd

web: bpf frontend
	@echo "==> Building Web Server..."
	@mkdir -p $(BUILD)
	@cp -r frontend cmd/
	@go build -tags web -o $(BUILD)/eulerguard-web ./cmd
	@rm -rf cmd/frontend

# Run
run-cli: cli
	@sudo $(BUILD)/eulerguard

run-web: web
	@echo "Open http://localhost:3000"
	@sudo $(BUILD)/eulerguard-web

# 
clean:
	@rm -f $(BPF_OBJ) $(BUILD)/eulerguard $(BUILD)/eulerguard-web
	@rm -rf $(BUILD)/bin cmd/frontend cmd/build

clean-all: clean
	@rm -rf ./frontend/node_modules ./frontend/dist

help:
	@echo "make cli     - CLI (no frontend)"
	@echo "make web     - Web server (:3000)"
	@echo "make run-*   - Build and run (sudo)"

.PHONY: all bpf frontend cli web dev run-cli run-web clean clean-all help
