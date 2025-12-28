# Aegis Build System
BPF_SRC = ./bpf/main.bpf.c
BPF_OBJ = ./bpf/main.bpf.o
VMLINUX = ./bpf/vmlinux.h
BUILD   = ./build

all: web

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
web: bpf frontend
	@echo "==> Building Web Server..."
	@mkdir -p $(BUILD)
	@cp -r frontend cmd/
	@go build -tags web -o $(BUILD)/aegis-web ./cmd
	@rm -rf cmd/frontend

# Run
run: web
	@echo "Open http://localhost:3000"
	@sudo $(BUILD)/aegis-web

# 
clean:
	@rm -f $(BPF_OBJ) $(BUILD)/aegis-web
	@rm -rf $(BUILD)/bin cmd/frontend cmd/build

clean-all: clean
	@rm -rf ./frontend/node_modules ./frontend/dist

help:
	@echo "make web     - Web server (:3000)"
	@echo "make run     - Build and run (sudo)"

.PHONY: all bpf frontend web dev run clean clean-all help
